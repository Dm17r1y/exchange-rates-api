package main

import (
	"database/sql"
	"exchange-rates-service/src/config"
	"log"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	config := config.NewConfig()

	db, err := sql.Open("postgres", config.PostgresConnectionString)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Error creating driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./src/migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Error creating migration: %s", err)
	}

	err = m.Up()
	if err != nil {
		log.Fatalf("Error performing migration: %s", err)
	}
}
