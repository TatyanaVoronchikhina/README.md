package importer

import (
	"context"
	"errors"
	"fmt"
	"moneychange/internal/domain"
	"moneychange/internal/source"
	"strings"
	"time"
)

type RateRecord struct {
	Observed_at    time.Time
	Source_at      time.Time
	Source         string
	Kind           string
	Bank           string
	City           string
	Channel        string
	Base           string
	Quote          string
	Buy_rate       string
	Sell_rate      string
	Reference_rate string
	Source_url     string
}

const CbrUrl = "https://cbr.ru"

func Run(ctx context.Context, filePath, sourceName string) ([]RateRecord, error) {
	observedAt := time.Now()
	sourceURL := CbrUrl
	if filePath != "" {
		sourceURL = filePath
	}
	rawBCR, err := source.GetRawCBR(ctx, filePath, CbrUrl)
	if err != nil {
		return nil, fmt.Errorf("Error getting raw BCR: %v", err)
	}
	defer rawBCR.Close()
	parseDataCbr, err := source.ParseRawDataCBR(rawBCR)
	if err != nil {
		return nil, fmt.Errorf("Error parsing raw BCR: %v", err)
	}
	sourceAt := observedAt
	if parseDataCbr.Date != "" {
		if t, err := time.Parse("02.01.2006", parseDataCbr.Date); err == nil {
			sourceAt = t
		}
	}
	var records []RateRecord
	for _, el := range parseDataCbr.Valute {
		upperCharcode := strings.ToUpper(el.Charcode)
		valDecimal, err := domain.ParsePositiveDecimal(el.Value, -1)
		if err != nil {
			return nil, fmt.Errorf("Error parsing value %v: %v", el.Value, err)
		}
		nomDecimal, err := domain.ParsePositiveDecimal(el.Nominal, -1)
		if err != nil {
			return nil, fmt.Errorf("Error parsing nominal %v: %v", el.Nominal, err)
		}
		rateDecimal := valDecimal.Div(nomDecimal)
		rateRecord := RateRecord{
			Observed_at:    observedAt,
			Source_at:      sourceAt,
			Source:         sourceName,
			Kind:           "reference",
			Base:           upperCharcode,
			Quote:          "RUB",
			Reference_rate: rateDecimal.StringFixed(4),
			Source_url:     sourceURL,
		}
		records = append(records, rateRecord)
	}
	if len(records) == 0 {
		return nil, errors.New("No records found")
	}
	return records, nil
}
