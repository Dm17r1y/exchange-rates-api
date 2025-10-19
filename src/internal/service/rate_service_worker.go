package service

import (
	"database/sql"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/integration"
	"exchange-rates-service/src/internal/repository"
	"exchange-rates-service/src/internal/storage"
	"log"
)

type RateServiceWorker struct {
	config     *config.Config
	repository *repository.ExchangeRateRepository
	client     *integration.ExchangeRateApiClient
}

func NewRateServiceWorker(config *config.Config) (*RateServiceWorker, error) {
	db, err := sql.Open("postgres", config.PostgresConnectionString)
	if err != nil {
		return nil, err
	}

	exchangeRateStorage := storage.NewExchangeRateStorage(db)
	exchangeRateUpdateStorage := storage.NewExchangeRateUpdateStorage(db)

	r := repository.NewExchangeRateRepository(db, exchangeRateStorage, exchangeRateUpdateStorage)

	serviceWorker := RateServiceWorker{
		config:     config,
		repository: r,
		client:     integration.NewExchangeRateIoClient(config),
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
