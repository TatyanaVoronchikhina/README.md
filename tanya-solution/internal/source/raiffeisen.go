package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/shopspring/decimal"

	"moneychange/internal/domain"
)

const (
	// source=CASH → channel=cash.
	raiffeisenURL  = "https://www.raiffeisen.ru/oapi/currency_rate/get/?counterCurrency=RUB&source=CASH&currencies=USD,EUR"
	raiffeisenBank = "Raiffeisenbank"
)

type raiffeisenResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Rates []struct {
			Code      string `json:"code"`
			UpdatedAt int64  `json:"updatedAt"`
			Exchange  []struct {
				Code  string `json:"code"`
				Rates struct {
					Buy  *raiffeisenRate `json:"buy"`
					Sell *raiffeisenRate `json:"sell"`
				} `json:"rates"`
			} `json:"exchange"`
		} `json:"rates"`
	} `json:"data"`
}

type raiffeisenRate struct {
	Value      json.Number `json:"value"`
	Multiplier json.Number `json:"multiplier"`
}

// counterCurrency=RUB: exchange[].code — base; курс за единицу — value / multiplier.
func parseRaiffeisen(r io.Reader, m Meta) ([]domain.RateRecord, error) {
	var doc raiffeisenResponse
	if err := decodeJSON(r, &doc); err != nil {
		return nil, err
	}
	if !doc.Success {
		return nil, errors.New("success=false")
	}

	var records []domain.RateRecord
	for _, group := range doc.Data.Rates {
		if group.Code != "RUB" {
			return nil, fmt.Errorf("unexpected counter currency %q", group.Code)
		}
		var sourceAt time.Time
		if group.UpdatedAt > 0 {
			sourceAt = time.Unix(group.UpdatedAt, 0).UTC()
		}
		for _, ex := range group.Exchange {
			buy, err := raiffeisenPerUnit(ex.Rates.Buy)
			if err != nil {
				return nil, fmt.Errorf("%s buy: %w", ex.Code, err)
			}
			sell, err := raiffeisenPerUnit(ex.Rates.Sell)
			if err != nil {
				return nil, fmt.Errorf("%s sell: %w", ex.Code, err)
			}
			records = append(records, bankRecord(m, raiffeisenBank, domain.ChannelCash, ex.Code, "RUB", buy, sell, sourceAt))
		}
	}
	if len(records) == 0 {
		return nil, errors.New("no rates in response")
	}
	return records, nil
}

func raiffeisenPerUnit(r *raiffeisenRate) (decimal.NullDecimal, error) {
	if r == nil || r.Value == "" {
		return decimal.NullDecimal{}, nil
	}
	if r.Multiplier == "" {
		return decimal.NullDecimal{}, errors.New("multiplier is missing")
	}
	v, err := perUnit(r.Value.String(), r.Multiplier.String())
	if err != nil {
		return decimal.NullDecimal{}, err
	}
	return decimal.NewNullDecimal(v), nil
}
