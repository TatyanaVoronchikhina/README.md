package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

func Result(amountStr, rateStr string) (string, error) {
	amount, err := makeDigit(amountStr)
	if err != nil {
		return "", fmt.Errorf("error by makeDigit: %w", err)
	}
	rate, err := makeDigit(rateStr)
	if err != nil {
		return "", fmt.Errorf("error by makeDigit: %w", err)
	}
	result := amount.Mul(rate).Round(2)
	formattedResult := result.StringFixed(2)
	return formattedResult, nil
}

func makeDigit(dataString string) (decimal.Decimal, error) {
	dataString = strings.TrimSpace(dataString)
	if dataString == "" {
		return decimal.Decimal{}, errors.New("value is empty")
	}
	dataString = strings.ReplaceAll(dataString, ",", ".")
	punctuation := strings.Count(dataString, ".")
	if punctuation > 1 {
		return decimal.Decimal{}, errors.New("too many punctuation marks")
	}
	dataArray := strings.Split(dataString, ".")
	if len(dataArray) > 1 && len(dataArray[1]) > 2 {
		return decimal.Decimal{}, errors.New("too many digits after the comma")
	}
	decimalValue, err := decimal.NewFromString(dataString)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("value is invalid `%s`: %w", dataString, err)
	}
	if decimalValue.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, errors.New("value is less or equal zero")
	}
	return decimalValue, nil
}
