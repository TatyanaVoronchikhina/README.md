// Package application — сценарии convert/best; CLI и HTTP используют одни и те же функции.
package application

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"moneychange/internal/domain"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNoFreshOffers   = errors.New("no fresh offers")
)

type RecordReader interface {
	ReadSince(since time.Time) ([]domain.RateRecord, error)
}

type Converter struct {
	Records RecordReader
	Now     func() time.Time
	MaxAge  time.Duration
}

type OfferRequest struct {
	From, To string
	Channel  string
	City     string // пусто — Moscow
	Bank     string // пусто — любой банк
}

type ConvertRequest struct {
	OfferRequest
	Amount string
}

type ConvertResult struct {
	Amount           decimal.Decimal
	EstimatedReceive decimal.Decimal
	Best             domain.Offer
	Offers           []domain.Offer
}

// BestOffers: пустой список — не ошибка.
func (c Converter) BestOffers(req OfferRequest) ([]domain.Offer, error) {
	filter, err := c.filter(req)
	if err != nil {
		return nil, err
	}
	records, err := c.Records.ReadSince(filter.Now.Add(-filter.MaxAge))
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	return domain.BestOffers(records, filter), nil
}

func (c Converter) Convert(req ConvertRequest) (ConvertResult, error) {
	amount, err := domain.ParseAmount(req.Amount)
	if err != nil {
		return ConvertResult{}, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}
	offers, err := c.BestOffers(req.OfferRequest)
	if err != nil {
		return ConvertResult{}, err
	}
	if len(offers) == 0 {
		return ConvertResult{}, ErrNoFreshOffers
	}
	return ConvertResult{
		Amount:           amount,
		EstimatedReceive: offers[0].EstimatedReceive(amount),
		Best:             offers[0],
		Offers:           offers,
	}, nil
}

func (c Converter) filter(req OfferRequest) (domain.OfferFilter, error) {
	from, err := domain.ParseCurrency(req.From)
	if err != nil {
		return domain.OfferFilter{}, fmt.Errorf("%w: from: %v", ErrInvalidArgument, err)
	}
	to, err := domain.ParseCurrency(req.To)
	if err != nil {
		return domain.OfferFilter{}, fmt.Errorf("%w: to: %v", ErrInvalidArgument, err)
	}
	if from == to {
		return domain.OfferFilter{}, fmt.Errorf("%w: from and to must differ", ErrInvalidArgument)
	}
	channel, err := domain.ParseChannel(req.Channel)
	if err != nil {
		return domain.OfferFilter{}, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}
	city := req.City
	if city == "" {
		city = domain.CityMoscow
	}
	return domain.OfferFilter{
		From: from, To: to, Channel: channel, City: city, Bank: req.Bank,
		Now: c.Now(), MaxAge: c.MaxAge,
	}, nil
}
