package main

import (
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/service"
	"log"
	"time"
)

var serviceConfig = config.NewConfig()
var rateServiceWorker *service.RateServiceWorker

func main() {
	var err error

	ticker := time.NewTicker(serviceConfig.WorkerTickInterval)

	rateServiceWorker, err = service.NewRateServiceWorker(serviceConfig)
	if err != nil {
		panic(err)
	}

	for {
		<-ticker.C

		updated, err := rateServiceWorker.ExecuteUpdate()
		if err != nil {
			log.Println(err)
		}

		if updated > 0 {
			log.Printf("Updated %d rates", updated)
		}
	}
}
