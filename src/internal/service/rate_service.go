package service

import (
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/repository"
)

type RateService struct {
	repository *repository.ExchangeRateRepository
}

func NewRateService(config *config.Config) *RateService {
	return &RateService{
		repository: repository.NewExchangeRateRepository(config),
	}
}

func (service *RateService) StartUpdateRate(from string, to string) (string, error) {
	return service.repository.GetOrCreateRateUpdate(from, to)
}

func (service *RateService) GetRateUpdate(updateId string) (*repository.ExchangeRate, error) {
	return service.repository.GetRateUpdate(updateId)
}

func (service *RateService) GetLastRate(from string, to string) (*repository.ExchangeRate, error) {
	return service.repository.GetLastRate(from, to)

}
