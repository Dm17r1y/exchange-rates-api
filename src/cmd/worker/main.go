package main

import (
	"database/sql"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/integration"
	"exchange-rates-service/src/internal/repository"
	"exchange-rates-service/src/internal/service"
	"exchange-rates-service/src/internal/storage"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {

	serviceConfig := config.NewConfig()
	db, err := sql.Open("postgres", serviceConfig.PostgresConnectionString)
	if err != nil {
		panic(err)
	}

	exchangeRateStorage := storage.NewExchangeRateStorage(db)
	exchangeRateUpdateStorage := storage.NewExchangeRateUpdateStorage(db)
	repo := repository.NewExchangeRateRepository(db, exchangeRateStorage, exchangeRateUpdateStorage)
	
	var client integration.ExchangeRateApiClient
	if serviceConfig.ExchangeIoApiKey != "" {
		client = integration.NewExchangeRateApiIoClient(serviceConfig)
	} else {
		client = integration.NewCurrencyApiClient(serviceConfig)
	}

	rateServiceWorker := service.NewRateServiceWorker(serviceConfig, repo, client)
	ticker := time.NewTicker(serviceConfig.WorkerTickInterval)

	for {
		<-ticker.C

		for {
			updated, err := rateServiceWorker.ExecuteUpdate()
			if err != nil {
				log.Println(err)
			}

			if updated == 0 {
				break
			}

			log.Printf("Updated %d rates", updated)
		}
	}
}
