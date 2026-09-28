package main

import (
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"

	"moneychange/internal/domain"
)

func runCalculate(args []string, log *slog.Logger) int {
	start := time.Now()
	fs := flag.NewFlagSet("calculate", flag.ContinueOnError)
	amountStr := fs.String("amount", "", "сумма, не больше двух знаков после точки")
	rateStr := fs.String("rate", "", "курс: сколько единиц целевой валюты за одну исходную")
	if ok, code := parseFlags(fs, args, log); !ok {
		return code
	}
	log.Debug("calculate input", "amount", *amountStr, "rate", *rateStr)

	result, err := calculate(*amountStr, *rateStr)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		log.Error("calculation failed", "error", err.Error(), "duration_ms", duration)
		return 1
	}
	log.Info("calculation succeeded", "records", 1, "duration_ms", duration)
	fmt.Println(result.StringFixed(2))
	return 0
}

func calculate(amountStr, rateStr string) (decimal.Decimal, error) {
	amount, err := domain.ParseAmount(amountStr)
	if err != nil {
		return decimal.Decimal{}, err
	}
	rate, err := domain.ParseRate(rateStr)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return domain.Calculate(amount, rate), nil
}
