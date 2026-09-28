package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Kind string

const (
	KindBank            Kind = "bank"
	KindReference       Kind = "reference"
	KindMarketReference Kind = "market_reference"
)

type Channel string

const (
	ChannelCash      Channel = "cash"
	ChannelCard      Channel = "card"
	ChannelAccount   Channel = "account"
	ChannelReference Channel = "reference"
)

const (
	CityMoscow = "Moscow"
	// Unknown — bank/city не подтверждены источником.
	Unknown = "unknown"
)

func ParseChannel(s string) (Channel, error) {
	switch c := Channel(strings.ToLower(strings.TrimSpace(s))); c {
	case ChannelCash, ChannelCard, ChannelAccount:
		return c, nil
	default:
		return "", fmt.Errorf("channel %q must be one of cash, card, account", s)
	}
}

var ErrInvalidRecord = errors.New("invalid rate record")

var sourceIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateSourceID: ID — часть имени файла снимка.
func ValidateSourceID(id string) error {
	if !sourceIDPattern.MatchString(id) {
		return fmt.Errorf("source id %q must match %s", id, sourceIDPattern)
	}
	return nil
}

// RateRecord — колонки CSV-контракта; курсы за одну единицу Base.
type RateRecord struct {
	ObservedAt    time.Time // получение сервисом
	SourceAt      time.Time // публикация источником; zero — неизвестно
	Source        string
	Kind          Kind
	Bank          string
	City          string
	Channel       Channel
	Base          string
	Quote         string
	BuyRate       decimal.NullDecimal
	SellRate      decimal.NullDecimal
	ReferenceRate decimal.NullDecimal
	SourceURL     string
}

func (r RateRecord) Validate() error {
	if err := r.validate(); err != nil {
		return fmt.Errorf("%w: %s %s/%s: %v", ErrInvalidRecord, r.Source, r.Base, r.Quote, err)
	}
	return nil
}

func (r RateRecord) validate() error {
	if err := ValidateSourceID(r.Source); err != nil {
		return err
	}
	if r.ObservedAt.IsZero() {
		return errors.New("observed_at is empty")
	}
	if _, err := ParseCurrency(r.Base); err != nil || r.Base != strings.ToUpper(r.Base) {
		return fmt.Errorf("base %q is not a currency code", r.Base)
	}
	if _, err := ParseCurrency(r.Quote); err != nil || r.Quote != strings.ToUpper(r.Quote) {
		return fmt.Errorf("quote %q is not a currency code", r.Quote)
	}
	if r.Base == r.Quote {
		return errors.New("base and quote are equal")
	}
	if r.Bank == "" || r.City == "" {
		return errors.New("bank and city must be set (use unknown)")
	}
	for name, v := range map[string]decimal.NullDecimal{"buy_rate": r.BuyRate, "sell_rate": r.SellRate, "reference_rate": r.ReferenceRate} {
		if v.Valid && !v.Decimal.IsPositive() {
			return fmt.Errorf("%s must be greater than zero", name)
		}
	}

	switch r.Kind {
	case KindBank:
		switch r.Channel {
		case ChannelCash, ChannelCard, ChannelAccount:
		default:
			return fmt.Errorf("bank record has channel %q", r.Channel)
		}
		if !r.BuyRate.Valid && !r.SellRate.Valid {
			return errors.New("bank record has neither buy_rate nor sell_rate")
		}
		if r.ReferenceRate.Valid {
			return errors.New("bank record must not have reference_rate")
		}
		// buy > sell — перепутанная семантика источника.
		if r.BuyRate.Valid && r.SellRate.Valid && r.BuyRate.Decimal.GreaterThan(r.SellRate.Decimal) {
			return errors.New("buy_rate is greater than sell_rate")
		}
	case KindReference, KindMarketReference:
		if r.Channel != ChannelReference {
			return fmt.Errorf("%s record must have channel reference", r.Kind)
		}
		if !r.ReferenceRate.Valid {
			return errors.New("reference_rate is empty")
		}
		if r.BuyRate.Valid || r.SellRate.Valid {
			return fmt.Errorf("%s record must not have buy/sell rates", r.Kind)
		}
	default:
		return fmt.Errorf("unknown kind %q", r.Kind)
	}
	return nil
}

// Key — последняя запись ищется отдельно для каждого ключа.
type Key struct {
	Source, Bank, City string
	Channel            Channel
	Base, Quote        string
}

func (r RateRecord) Key() Key {
	return Key{Source: r.Source, Bank: r.Bank, City: r.City, Channel: r.Channel, Base: r.Base, Quote: r.Quote}
}

// Snapshot — одно получение одного источника; сохраняется целиком.
type Snapshot struct {
	Source     string
	ObservedAt time.Time
	Records    []RateRecord
}

func NewSnapshot(source string, records []RateRecord) (Snapshot, error) {
	if err := ValidateSourceID(source); err != nil {
		return Snapshot{}, err
	}
	if len(records) == 0 {
		return Snapshot{}, fmt.Errorf("source %s returned no records", source)
	}
	observedAt := records[0].ObservedAt
	for _, r := range records {
		if r.Source != source {
			return Snapshot{}, fmt.Errorf("record of source %q in snapshot of %q", r.Source, source)
		}
		if !r.ObservedAt.Equal(observedAt) {
			return Snapshot{}, errors.New("records of one snapshot have different observed_at")
		}
		if err := r.Validate(); err != nil {
			return Snapshot{}, err
		}
	}
	return Snapshot{Source: source, ObservedAt: observedAt, Records: records}, nil
}
