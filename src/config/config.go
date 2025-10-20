package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresConnectionString string
	ExchangeIoApiKey         string
	WorkerFetchSize          int
	WorkerTickInterval       time.Duration
	HttpClientTimeout 	time.Duration
}

func NewConfig() *Config {
	err := godotenv.Load(".env", ".env.secret")
	if err != nil {
		log.Println("Warning:", err)
	}

	workerFetchSize, err := strconv.Atoi(os.Getenv("WORKER_FETCH_SIZE"))
	if err != nil {
		log.Fatalf("Unable to parse WORKER_FETCH_SIZE: %s", err)
	}

	WorkerTickInterval, err := strconv.Atoi(os.Getenv("WORKER_TICK_INTERVAL_MILLISECONDS"))
	if err != nil {
		log.Fatalf("Unable to parse WORKER_TICK_INTERVAL_MILLISECONDS: %s", err)
	}

	postgresConnectionString := os.Getenv("POSTGRES_CONNECTION_STRING")
	if postgresConnectionString == "" {
		log.Fatal("POSTGRES_CONNECTION_STRING is not set")
	}

	exchangeRatesApiKey := os.Getenv("EXCHANGE_RATES_IO_API_KEY")
	if exchangeRatesApiKey == "" {
		log.Println("EXCHANGE_RATES_IO_API_KEY is not set. currency-api will be used")
	} else {
		log.Println("EXCHANGE_RATES_IO_API_KEY found. exchangeratesapi.io api will be used")

	}

	httpClientTimeout, err := strconv.Atoi(os.Getenv("HTTP_CLIENT_TIMEOUT_MS"))
	if err != nil {
		log.Fatalf("Unable to parse HTTP_CLIENT_TIMEOUT_MS: %s", err)
	}
	
	config := Config{
		PostgresConnectionString: postgresConnectionString,
		WorkerFetchSize:          workerFetchSize,
		WorkerTickInterval:       time.Duration(WorkerTickInterval) * time.Millisecond,
		ExchangeIoApiKey:         exchangeRatesApiKey,
		HttpClientTimeout: 		  time.Duration(httpClientTimeout) * time.Millisecond,
	}

	return &config
}
