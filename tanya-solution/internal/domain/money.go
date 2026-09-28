package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

const NoFractionLimit = -1

// RatePrecision — деление курсов (1/sell, Value/Nominal): 18 знаков, half-even.
const RatePrecision = 18

var one = decimal.NewFromInt(1)

func ParseAmount(s string) (decimal.Decimal, error) {
	v, err := ParsePositiveDecimal(s, 2)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid amount: %w", err)
	}
	return v, nil
}

func ParseRate(s string) (decimal.Decimal, error) {
	v, err := ParsePositiveDecimal(s, NoFractionLimit)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid rate: %w", err)
	}
	return v, nil
}

// ParsePositiveDecimal: без float64, десятичная запятая допускается.
func ParsePositiveDecimal(s string, maxFraction int) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Decimal{}, errors.New("value is empty")
	}
	s = strings.ReplaceAll(s, ",", ".")
	if strings.Count(s, ".") > 1 {
		return decimal.Decimal{}, errors.New("too many decimal separators")
	}
	if maxFraction >= 0 {
		if parts := strings.Split(s, "."); len(parts) > 1 && len(parts[1]) > maxFraction {
			return decimal.Decimal{}, fmt.Errorf("more than %d digits after the decimal point", maxFraction)
		}
	}
	value, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("value %q is not a number", s)
	}
	if !value.IsPositive() {
		return decimal.Decimal{}, errors.New("value must be greater than zero")
	}
	return value, nil
}

// DivHalfEven: decimal.DivRound округляет half-up, поэтому банковское округление своё.
func DivHalfEven(a, b decimal.Decimal, prec int32) (decimal.Decimal, error) {
	if b.IsZero() {
		return decimal.Decimal{}, errors.New("division by zero")
	}
	// a = b·q + r, где q кратно 10^-prec и 0 ≤ |r| < |b|·10^-prec.
	q, r := a.QuoRem(b, prec)
	if r.IsZero() {
		return q, nil
	}
	cmp := r.Abs().Mul(decimal.NewFromInt(2)).Cmp(b.Abs().Shift(-prec))
	lastDigitOdd := q.Shift(prec).BigInt().Bit(0) == 1
	if cmp > 0 || (cmp == 0 && lastDigitOdd) {
		unit := decimal.New(1, -prec)
		if a.Sign()*b.Sign() < 0 {
			q = q.Sub(unit)
		} else {
			q = q.Add(unit)
		}
	}
	return q, nil
}

func Invert(rate decimal.Decimal) (decimal.Decimal, error) {
	return DivHalfEven(one, rate, RatePrecision)
}

func ParseCurrency(s string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(s))
	if len(code) != 3 {
		return "", fmt.Errorf("currency %q must be a 3-letter code", s)
	}
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return "", fmt.Errorf("currency %q must be a 3-letter code", s)
		}
	}
	return code, nil
}
