package domain

import (
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type OfferFilter struct {
	From, To string
	Channel  Channel
	City     string
	Bank     string // пусто — любой банк
	Now      time.Time
	MaxAge   time.Duration
}

// Offer — предложение в направлении запроса From → To.
type Offer struct {
	Record  RateRecord
	Rate    decimal.Decimal // To за одну единицу From
	Inverse bool            // 1/sell_rate обратной пары
}

func (o Offer) EstimatedReceive(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(o.Rate).RoundBank(2)
}

func LatestByKey(records []RateRecord) []RateRecord {
	latest := make(map[Key]RateRecord, len(records))
	for _, r := range records {
		if cur, ok := latest[r.Key()]; !ok || r.ObservedAt.After(cur.ObservedAt) {
			latest[r.Key()] = r
		}
	}
	out := make([]RateRecord, 0, len(latest))
	for _, r := range latest {
		out = append(out, r)
	}
	return out
}

func BestOffers(records []RateRecord, f OfferFilter) []Offer {
	var offers []Offer
	for _, r := range LatestByKey(records) {
		if !f.matches(r) {
			continue
		}
		if o, ok := toOffer(r, f.From, f.To); ok {
			offers = append(offers, o)
		}
	}
	sort.Slice(offers, func(i, j int) bool {
		a, b := offers[i], offers[j]
		if c := a.Rate.Cmp(b.Rate); c != 0 {
			return c > 0
		}
		if !a.Record.ObservedAt.Equal(b.Record.ObservedAt) {
			return a.Record.ObservedAt.After(b.Record.ObservedAt)
		}
		if a.Record.Bank != b.Record.Bank {
			return a.Record.Bank < b.Record.Bank
		}
		return a.Record.Source < b.Record.Source
	})
	return offers
}

func (f OfferFilter) matches(r RateRecord) bool {
	if r.Kind != KindBank || r.City != f.City || r.Channel != f.Channel {
		return false
	}
	if f.Bank != "" && !strings.EqualFold(r.Bank, f.Bank) {
		return false
	}
	return f.Now.Sub(r.ObservedAt) <= f.MaxAge
}

// toOffer: прямая пара — buy_rate, обратная — 1/sell_rate.
func toOffer(r RateRecord, from, to string) (Offer, bool) {
	switch {
	case r.Base == from && r.Quote == to && r.BuyRate.Valid:
		return Offer{Record: r, Rate: r.BuyRate.Decimal}, true
	case r.Base == to && r.Quote == from && r.SellRate.Valid:
		rate, err := Invert(r.SellRate.Decimal)
		if err != nil {
			return Offer{}, false
		}
		return Offer{Record: r, Rate: rate, Inverse: true}, true
	default:
		return Offer{}, false
	}
}
