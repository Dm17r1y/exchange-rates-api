package storage

import (
	"context"
	"database/sql"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal"
	"time"

	_ "github.com/lib/pq"
)

type ExchangeRateStorage struct {
	config *config.Config
}

func NewExchangeRateStorage(config *config.Config) *ExchangeRateStorage {
	return &ExchangeRateStorage{config: config}
}

type ExchangeRate struct {
	FromCurrency string
	ToCurrency   string
	RateValue    []byte
	UpdateTime   *time.Time
}

const getRateSql = `
SELECT rate_value, update_time FROM exchange_rate
WHERE from_currency = $1 AND to_currency = $2
`

func (storage *ExchangeRateStorage) GetRate(from string, to string) (*ExchangeRate, error) {
	db, err := sql.Open("postgres", storage.config.PostgresConnectionString)
	if err != nil {
		return nil, err
	}

	stmt, err := db.PrepareContext(context.Background(), getRateSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasData := rows.Next()
	if !hasData {
		return nil, internal.NewNotFoundError("rate updates not found")
	}

	rate := ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
	}
	err = rows.Scan(&rate.RateValue, &rate.UpdateTime)
	return &rate, err
}
