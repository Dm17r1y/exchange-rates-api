package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type ExchangeRate struct {
	Rate           decimal.Decimal
	UpdateDateTime *time.Time
}

type ExchangeRateDbo struct {
	FromCurrency string
	ToCurrency   string
	RateValue    []byte
	UpdateTime   *time.Time
}

type ExchangeRateUpdateStatus int

const (
	StatusUpdating ExchangeRateUpdateStatus = iota
	StatusDone
	StatusError
)

type ExchangeRateUpdateDbo struct {
	Id           string
	FromCurrency string
	ToCurrency   string
	Status       ExchangeRateUpdateStatus
	RateValue    []byte
	UpdateTime   *time.Time
}
