package source

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"moneychange/internal/domain"
)

func referenceRecord(m Meta, kind domain.Kind, base string, rate decimal.Decimal, sourceAt time.Time) domain.RateRecord {
	return domain.RateRecord{
		ObservedAt:    m.ObservedAt,
		SourceAt:      sourceAt,
		Source:        m.SourceID,
		Kind:          kind,
		Bank:          domain.Unknown,
		City:          domain.Unknown,
		Channel:       domain.ChannelReference,
		Base:          strings.ToUpper(base),
		Quote:         "RUB",
		ReferenceRate: decimal.NewNullDecimal(rate),
		SourceURL:     m.SourceURL,
	}
}

// bankRecord: город unknown, пока применимость к Москве не подтверждена.
func bankRecord(m Meta, bank string, ch domain.Channel, base, quote string, buy, sell decimal.NullDecimal, sourceAt time.Time) domain.RateRecord {
	return domain.RateRecord{
		ObservedAt: m.ObservedAt,
		SourceAt:   sourceAt,
		Source:     m.SourceID,
		Kind:       domain.KindBank,
		Bank:       bank,
		City:       domain.Unknown,
		Channel:    ch,
		Base:       strings.ToUpper(base),
		Quote:      strings.ToUpper(quote),
		BuyRate:    buy,
		SellRate:   sell,
		SourceURL:  m.SourceURL,
	}
}

func perUnit(value, nominal string) (decimal.Decimal, error) {
	v, err := domain.ParseRate(value)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("value: %w", err)
	}
	n, err := domain.ParsePositiveDecimal(nominal, domain.NoFractionLimit)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("nominal: %w", err)
	}
	return domain.DivHalfEven(v, n, domain.RatePrecision)
}

// decodeJSON: json.Number, чтобы курсы не проходили через float64.
func decodeJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	return nil
}

func optionalRate(n json.Number) (decimal.NullDecimal, error) {
	if n == "" {
		return decimal.NullDecimal{}, nil
	}
	v, err := domain.ParseRate(n.String())
	if err != nil {
		return decimal.NullDecimal{}, err
	}
	return decimal.NewNullDecimal(v), nil
}
