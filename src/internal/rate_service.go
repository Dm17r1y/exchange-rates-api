package internal

import (
	"github.com/shopspring/decimal"
	"time"
)

type Rate struct {
	Rate decimal.Decimal
	UpdateDateTime time.Time
}

func StartUpdateRate(from string, to string) (string, error) {
	return "123", nil
}

func CheckUpdateRate(updateId string) (Rate, error) {
	
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

func GetLastRate(from string, to string) (Rate, error) {
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