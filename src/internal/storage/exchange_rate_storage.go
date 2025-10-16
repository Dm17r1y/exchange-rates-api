package storage

import (
	"context"
	"database/sql"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"

	_ "github.com/lib/pq"
)

type ExchangeRateStorage struct {
	db *sql.DB
}

func NewExchangeRateStorage(db *sql.DB) *ExchangeRateStorage {
	return &ExchangeRateStorage{db: db}
}


const getRateSql = `
SELECT rate_value, update_time FROM exchange_rate
WHERE from_currency = $1 AND to_currency = $2
`

func (storage *ExchangeRateStorage) GetRate(from string, to string) (*model.ExchangeRateDbo, error) {
	stmt, err := storage.db.PrepareContext(context.Background(), getRateSql)
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

	rate := model.ExchangeRateDbo{
		FromCurrency: from,
		ToCurrency:   to,
	}
	err = rows.Scan(&rate.RateValue, &rate.UpdateTime)
	return &rate, err
}

const setRateSql = `
INSERT INTO exchange_rate(from_currency, to_currency, rate_value, update_time)
VALUES ($1, $2, $3, $4) 
ON CONFLICT(from_currency, to_currency) 
DO UPDATE SET rate_value = $3, update_time = $4
`

func (storage *ExchangeRateStorage) SetRateTx(tx *sql.Tx, model *model.ExchangeRateDbo) error {
	stmt, err := tx.PrepareContext(context.Background(), setRateSql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(context.Background(), model.FromCurrency, model.ToCurrency, model.RateValue, model.UpdateTime)
	return err
}