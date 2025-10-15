package service

import (
	"github.com/shopspring/decimal"
	"time"
)

type RateService struct {}

type Rate struct {
	Rate decimal.Decimal
	UpdateDateTime time.Time
}

func NewRateService() *RateService {
	return &RateService{}
}

func (r *RateService) StartUpdateRate(from string, to string) (string, error) {
	return "123", nil
}

func (r *RateService) CheckUpdateRate(updateId string) (Rate, error) {
	
	rateVal, err := decimal.NewFromString("19.99")
	if err != nil {
		return Rate{}, err
	}

	rate := Rate{
		Rate: rateVal,
		UpdateDateTime: time.Now(),
	}
	return rate, nil
}

func (r *RateService) GetLastRate(from string, to string) (Rate, error) {
	rateVal, err := decimal.NewFromString("19.99")
	if err != nil {
		return Rate{}, err
	}

	rate := Rate{
		Rate: rateVal,
		UpdateDateTime: time.Now(),
	}
	return rate, nil
}