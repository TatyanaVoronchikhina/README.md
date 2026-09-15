package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// noFractionLimit — rate, в отличие от amount, не ограничен двумя знаками после точки.
const noFractionLimit = -1

func Convert(amountStr, rateStr string) (string, error) {
	amount, err := parsePositiveDecimal(amountStr, 2)
	if err != nil {
		return "", fmt.Errorf("invalid amount: %w", err)
	}
	rate, err := parsePositiveDecimal(rateStr, noFractionLimit)
	if err != nil {
		return "", fmt.Errorf("invalid rate: %w", err)
	}
	// RoundBank — банковское округление half-even по постановке: 0.125 → 0.12, 0.135 → 0.14.
	result := amount.Mul(rate).RoundBank(2)
	return result.StringFixed(2), nil
}

func parsePositiveDecimal(s string, maxFraction int) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Decimal{}, errors.New("value is empty")
	}
	s = strings.ReplaceAll(s, ",", ".")
	if strings.Count(s, ".") > 1 {
		return decimal.Decimal{}, errors.New("too many punctuation marks")
	}
	if maxFraction >= 0 {
		if parts := strings.Split(s, "."); len(parts) > 1 && len(parts[1]) > maxFraction {
			return decimal.Decimal{}, errors.New("too many digits after the comma")
		}
	}
	value, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("value is invalid `%s`: %w", s, err)
	}
	if value.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, errors.New("value is less or equal zero")
	}
	return value, nil
}
