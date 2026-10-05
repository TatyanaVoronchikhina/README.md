package domain

import "github.com/shopspring/decimal"

func Calculate(amount, rate decimal.Decimal) decimal.Decimal {
	return amount.Mul(rate).RoundBank(2)
}
