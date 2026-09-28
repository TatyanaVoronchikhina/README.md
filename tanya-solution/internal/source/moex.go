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
	moexURL = "https://iss.moex.com/iss/engines/currency/markets/selt/securities/USD000UTSTOM.json"
	// Основной режим торгов USD/RUB TOM.
	moexBoard = "CETS"
)

// moexTable — формат ISS: колонки отдельно от строк.
type moexTable struct {
	Columns []string `json:"columns"`
	Data    [][]any  `json:"data"`
}

type moexResponse struct {
	MarketData moexTable `json:"marketdata"`
}

func parseMOEX(r io.Reader, m Meta) ([]domain.RateRecord, error) {
	var doc moexResponse
	if err := decodeJSON(r, &doc); err != nil {
		return nil, err
	}
	row, err := doc.MarketData.row("BOARDID", moexBoard)
	if err != nil {
		return nil, err
	}

	price := row["WAPRICE"]
	if price == "" {
		price = row["LAST"]
	}
	if price == "" {
		return nil, errors.New("both WAPRICE and LAST are null")
	}
	rate, err := domain.ParseRate(price)
	if err != nil {
		return nil, fmt.Errorf("price: %w", err)
	}

	var sourceAt time.Time
	if v := row["SYSTIME"]; v != "" {
		t, err := time.ParseInLocation(time.DateTime, v, domain.Moscow)
		if err != nil {
			return nil, fmt.Errorf("SYSTIME %q: %w", v, err)
		}
		sourceAt = t
	}
	return []domain.RateRecord{referenceRecord(m, domain.KindMarketReference, "USD", rate, sourceAt)}, nil
}

// row — первая строка с key=value как map колонка → значение; null → "".
func (t moexTable) row(key, value string) (map[string]string, error) {
	idx := -1
	for i, c := range t.Columns {
		if c == key {
			idx = i
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("column %s not found", key)
	}
	for _, data := range t.Data {
		if len(data) != len(t.Columns) || fmt.Sprint(data[idx]) != value {
			continue
		}
		row := make(map[string]string, len(t.Columns))
		for i, c := range t.Columns {
			switch v := data[i].(type) {
			case nil:
				row[c] = ""
			case json.Number:
				row[c] = v.String()
			case string:
				row[c] = v
			default:
				row[c] = fmt.Sprint(v)
			}
		}
		return row, nil
	}
	return nil, fmt.Errorf("row with %s=%s not found", key, value)
}
