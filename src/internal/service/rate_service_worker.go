package service

import (
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/integration"
	"exchange-rates-service/src/internal/repository"
	"log"
)

type RateServiceWorker struct {
	config     *config.Config
	repository repository.ExchangeRateRepository
	client     integration.ExchangeRateApiClient
}

func NewRateServiceWorker(config *config.Config, repo repository.ExchangeRateRepository, client integration.ExchangeRateApiClient) *RateServiceWorker {

	serviceWorker := RateServiceWorker{
		config:     config,
		repository: repo,
		client:     client,
	}

	return &serviceWorker
}

func (s *RateServiceWorker) ExecuteUpdate() (int, error) {
	rateUpdates, err := s.repository.GetRatesForUpdate(s.config.WorkerFetchSize)
	if err != nil {
		return 0, err
	}
	updateCount := 0

	for _, rateUpdate := range rateUpdates {
		rate, err := s.client.GetRate(rateUpdate.FromCurrency, rateUpdate.ToCurrency)
		if err != nil {
			s.repository.SetUpdateError(rateUpdate.Id)
			log.Println(err)
			updateCount++
			continue
		}

		if err := s.repository.UpdateRate(rateUpdate.Id, rateUpdate.FromCurrency, rateUpdate.ToCurrency, rate); err != nil {
			return updateCount, err
		}
		updateCount++
	}

	return updateCount, nil
}
