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
	ExchangeIoApiKey string
	WorkerFetchSize int
	WorkerTickInterval time.Duration
}

func NewConfig() *Config {
	err := godotenv.Load(".env", ".env.secret")
	if err != nil {
		log.Fatal(err)
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
		log.Fatal("EXCHANGE_RATES_IO_API_KEY is not set")
	}

	config := Config{
		PostgresConnectionString: postgresConnectionString,
		WorkerFetchSize: workerFetchSize,
		WorkerTickInterval: time.Duration(WorkerTickInterval) * time.Millisecond,
		ExchangeIoApiKey: exchangeRatesApiKey,
	}

	return &config
}
