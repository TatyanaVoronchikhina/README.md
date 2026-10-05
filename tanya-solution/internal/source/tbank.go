package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"moneychange/internal/domain"
)

const (
	tbankURL = "https://www.tbank.ru/api/v1/currency_rates/"
	// Единственная категория MVP: обмен между счетами → channel=account.
	tbankCategory = "DepositPayments"
	tbankBank     = "T-Bank"
)

type tbankResponse struct {
	ResultCode string `json:"resultCode"`
	Payload    struct {
		LastUpdate struct {
			Milliseconds int64 `json:"milliseconds"`
		} `json:"lastUpdate"`
		Rates []struct {
			Category     string        `json:"category"`
			FromCurrency tbankCurrency `json:"fromCurrency"`
			ToCurrency   tbankCurrency `json:"toCurrency"`
			Buy          json.Number   `json:"buy"`
			Sell         json.Number   `json:"sell"`
		} `json:"rates"`
	} `json:"payload"`
}

type tbankCurrency struct {
	Name string `json:"name"`
}

func parseTBank(r io.Reader, m Meta) ([]domain.RateRecord, error) {
	var doc tbankResponse
	if err := decodeJSON(r, &doc); err != nil {
		return nil, err
	}
	if doc.ResultCode != "OK" {
		return nil, fmt.Errorf("resultCode %q", doc.ResultCode)
	}
	var sourceAt time.Time
	if ms := doc.Payload.LastUpdate.Milliseconds; ms > 0 {
		sourceAt = time.UnixMilli(ms).UTC()
	}

	var records []domain.RateRecord
	for _, rate := range doc.Payload.Rates {
		if rate.Category != tbankCategory {
			continue
		}
		pair := rate.FromCurrency.Name + "/" + rate.ToCurrency.Name
		buy, err := optionalRate(rate.Buy)
		if err != nil {
			return nil, fmt.Errorf("%s buy: %w", pair, err)
		}
		sell, err := optionalRate(rate.Sell)
		if err != nil {
			return nil, fmt.Errorf("%s sell: %w", pair, err)
		}
		records = append(records, bankRecord(m, tbankBank, domain.ChannelAccount,
			rate.FromCurrency.Name, rate.ToCurrency.Name, buy, sell, sourceAt))
	}
	if len(records) == 0 {
		return nil, errors.New("no " + tbankCategory + " rates in response")
	}
	return records, nil
}
