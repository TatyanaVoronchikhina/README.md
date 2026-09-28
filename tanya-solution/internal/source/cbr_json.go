package source

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"moneychange/internal/domain"
)

const cbrJSONURL = "https://www.cbr-xml-daily.ru/daily_json.js"

type cbrJSONDaily struct {
	Date   string                   `json:"Date"`
	Valute map[string]cbrJSONValute `json:"Valute"`
}

type cbrJSONValute struct {
	CharCode string      `json:"CharCode"`
	Nominal  json.Number `json:"Nominal"`
	Value    json.Number `json:"Value"`
}

func parseCBRJSON(r io.Reader, m Meta) ([]domain.RateRecord, error) {
	var doc cbrJSONDaily
	if err := decodeJSON(r, &doc); err != nil {
		return nil, err
	}
	if len(doc.Valute) == 0 {
		return nil, fmt.Errorf("field Valute is empty")
	}
	var sourceAt time.Time
	if doc.Date != "" {
		t, err := time.Parse(time.RFC3339, doc.Date)
		if err != nil {
			return nil, fmt.Errorf("Date %q: %w", doc.Date, err)
		}
		sourceAt = t
	}

	codes := make([]string, 0, len(doc.Valute))
	for code := range doc.Valute {
		codes = append(codes, code)
	}
	sort.Strings(codes) // map без порядка, строки снимка — стабильно

	records := make([]domain.RateRecord, 0, len(codes))
	for _, code := range codes {
		v := doc.Valute[code]
		rate, err := perUnit(v.Value.String(), v.Nominal.String())
		if err != nil {
			return nil, fmt.Errorf("valute %s: %w", code, err)
		}
		records = append(records, referenceRecord(m, domain.KindReference, v.CharCode, rate, sourceAt))
	}
	return records, nil
}
