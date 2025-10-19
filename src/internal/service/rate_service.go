package service

import (
	"database/sql"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"
	"exchange-rates-service/src/internal/repository"
	"exchange-rates-service/src/internal/storage"
	"fmt"
)

type RateService struct {
	supportedCurrencies map[string]bool
	repository          *repository.ExchangeRateRepository
}

func NewRateService(config *config.Config) (*RateService, error) {
	db, err := sql.Open("postgres", config.PostgresConnectionString)
	if err != nil {
		return nil, err
	}

	exchangeRateStorage := storage.NewExchangeRateStorage(db)
	exchangeRateUpdateStorage := storage.NewExchangeRateUpdateStorage(db)

	r := repository.NewExchangeRateRepository(db, exchangeRateStorage, exchangeRateUpdateStorage)

	return &RateService{
		supportedCurrencies: map[string]bool{
			"EUR": true,
			"USD": true,
			"MXN": true,
		},
		repository: r,
	}, nil
}

func (service *RateService) StartUpdateRate(from string, to string) (string, error) {
	if _, ok := service.supportedCurrencies[from]; !ok {
		return "", internal.NewBadRequestError(fmt.Sprintf("currency %s not supported", from))
	}

	if _, ok := service.supportedCurrencies[to]; !ok {
		return "", internal.NewBadRequestError(fmt.Sprintf("currency %s not supported", to))
	}

	if from == to {
		return "", internal.NewBadRequestError(fmt.Sprintf("trying to convert same currency: %s to %s", from, to))
	}

	return service.repository.GetOrCreateRateUpdate(from, to)
}

func (service *RateService) GetRateUpdate(updateId string) (model.ExchangeRate, error) {
	return service.repository.GetRateUpdate(updateId)
}

func (service *RateService) GetLastRate(from string, to string) (model.ExchangeRate, error) {
	if _, ok := service.supportedCurrencies[from]; !ok {
		return model.ExchangeRate{}, internal.NewBadRequestError(fmt.Sprintf("currency %s not supported", from))
	}

	if _, ok := service.supportedCurrencies[to]; !ok {
		return model.ExchangeRate{}, internal.NewBadRequestError(fmt.Sprintf("currency %s not supported", to))
	}

	if from == to {
		return model.ExchangeRate{}, internal.NewBadRequestError(fmt.Sprintf("trying to get same currency rate: %s to %s", from, to))
	}

	return service.repository.GetLastRate(from, to)

}
