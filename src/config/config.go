package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	PostgresConnectionString string
}

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Env file not found")
	}

	config := Config{
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
	}

	return &config
}
