package repository

import (
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/storage"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExchangeRateRepository struct {
	rateStorage   *storage.ExchangeRateStorage
	updateStorage *storage.ExchangeRateUpdateStorage
}

func NewExchangeRateRepository(config *config.Config) *ExchangeRateRepository {
	return &ExchangeRateRepository{
		rateStorage:   storage.NewExchangeRateStorage(config),
		updateStorage: storage.NewExchangeRateUpdateStorage(config),
	}
}

type ExchangeRate struct {
	Rate           decimal.Decimal
	UpdateDateTime *time.Time
}

func (r *ExchangeRateRepository) GetOrCreateRateUpdate(from string, to string) (string, error) {
	updateId := uuid.New()
	update, err := r.updateStorage.GetOrCreateRateUpdate(updateId.String(), from, to)
	if err != nil {
		return "", err
	}
	return update.Id, nil
}

func (r *ExchangeRateRepository) GetRateUpdate(updateId string) (ExchangeRate, error) {
	update, err := r.updateStorage.GetRateUpdate(updateId)
	if err != nil {
		return ExchangeRate{}, err
	}

	if update.Status != storage.Done {
		return ExchangeRate{}, nil
	}

	rateValue, err := decimal.NewFromString(string(update.RateValue))
	if err != nil {
		return ExchangeRate{}, err
	}

	rate := ExchangeRate{
		Rate:           rateValue,
		UpdateDateTime: update.UpdateTime,
	}

	return rate, nil
}

func (r *ExchangeRateRepository) GetLastRate(from string, to string) (ExchangeRate, error) {
	rate, err := r.rateStorage.GetRate(from, to)

	if err != nil {
		return ExchangeRate{}, err
	}

	if rate == nil {
		return ExchangeRate{}, nil
	}

	rateValue, err := decimal.NewFromString(string(rate.RateValue))
	if err != nil {
		return ExchangeRate{}, err
	}

	resultRate := ExchangeRate{
		Rate:           rateValue,
		UpdateDateTime: rate.UpdateTime,
	}

	return resultRate, nil
}
