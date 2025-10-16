package service

import (
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/integration"
	"exchange-rates-service/src/internal/repository"
	"log"
)

type RateServiceWorker struct {
	config *config.Config
	repository *repository.ExchangeRateRepository
	client *integration.ExchangeRateClient
}

func NewRateServiceWorker(config *config.Config) (*RateServiceWorker, error) {
	r, err := repository.NewExchangeRateRepository(config)
	if err != nil {
		return nil, err
	}

	serviceWorker := RateServiceWorker{
		config: config,
		repository: r,
		client: integration.NewExchangeRateClient(config),
	}

	return &serviceWorker, nil
}

func (s *RateServiceWorker) ExecuteUpdate() (int, error) {
	rateUpdates, err := s.repository.GetRatesForUpdate(s.config.WorkerFetchSize)
	if err != nil {
		return 0, err
	}
	updateCount := 0	

	for _, rateUpdate := range rateUpdates {
		response, err := s.client.GetRates(rateUpdate.FromCurrency, rateUpdate.ToCurrency)
		if err != nil {
			s.repository.SetUpdateError(rateUpdate.Id)
			log.Println(err)
			updateCount++
		}

		if err := s.repository.UpdateRate(rateUpdate.Id, rateUpdate.FromCurrency, rateUpdate.ToCurrency, response.Value); err != nil {
			return updateCount, err
		}
		updateCount++
	}

	return updateCount, nil
}