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
	WorkerFetchSize int
	WorkerTickInterval time.Duration
}

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Env file not found")
	}

	workerFetchSize, err := strconv.Atoi(os.Getenv("WORKER_FETCH_SIZE"))
	if err != nil {
		log.Fatal("Unable to parse WORKER_FETCH_SIZE")
	}

	WorkerTickInterval, err := strconv.Atoi(os.Getenv("WORKER_TICK_INTERVAL_MILLISECONDS"))
	if err != nil {
		log.Fatal("Unable to parse WORKER_TICK_INTERVAL_MILLISECONDS")
	}

	config := Config{
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
		WorkerFetchSize: workerFetchSize,
		WorkerTickInterval: time.Duration(WorkerTickInterval) * time.Millisecond,
	}

	return &config
}
