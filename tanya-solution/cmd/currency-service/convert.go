package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"time"

	"moneychange/internal/application"
	"moneychange/internal/archive"
	"moneychange/internal/config"
	"moneychange/internal/domain"
)

const rateDisplayPlaces = 12

func runConvert(args []string, cfg config.Config, log *slog.Logger) int {
	start := time.Now()
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	var req application.ConvertRequest
	fs.StringVar(&req.From, "from", "", "валюта, которую отдаём (base)")
	fs.StringVar(&req.To, "to", "", "валюта, которую получаем (quote)")
	fs.StringVar(&req.Amount, "amount", "", "сумма в валюте from")
	fs.StringVar(&req.Channel, "channel", "", "cash, card или account")
	fs.StringVar(&req.City, "city", domain.CityMoscow, "город предложения; unknown — записи без подтверждённой географии")
	fs.StringVar(&req.Bank, "bank", "", "только указанный банк")
	if ok, code := parseFlags(fs, args, log); !ok {
		return code
	}

	conv := application.Converter{Records: archive.New(cfg.DataDir), Now: time.Now, MaxAge: cfg.MaxRateAge}
	res, err := conv.Convert(req)
	duration := time.Since(start).Milliseconds()
	switch {
	case errors.Is(err, application.ErrNoFreshOffers):
		log.Warn("no fresh offers", "from", req.From, "to", req.To, "channel", req.Channel, "city", req.City, "duration_ms", duration)
		return 1
	case err != nil:
		log.Error("conversion failed", "error", err.Error(), "duration_ms", duration)
		return 1
	}

	best := res.Best.Record
	log.Info("conversion succeeded", "source", best.Source, "records", len(res.Offers), "duration_ms", duration)
	fmt.Printf("%s %s | %s (%s) %s %s | rate %s | observed_at %s | source_at %s | offers %d | ориентировочно, без комиссий\n",
		res.EstimatedReceive.StringFixed(2), req.To,
		best.Bank, best.Source, best.Channel, best.City,
		res.Best.Rate.StringFixedBank(rateDisplayPlaces),
		best.ObservedAt.UTC().Format(time.RFC3339), formatSourceAt(best.SourceAt),
		len(res.Offers))
	return 0
}

func formatSourceAt(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.UTC().Format(time.RFC3339)
}
