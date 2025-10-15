package storage

import (
	"context"
	"database/sql"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"

	_ "github.com/lib/pq"
)

type ExchangeRateUpdateStorage struct {
	config *config.Config
}

func NewExchangeRateUpdateStorage(config *config.Config) *ExchangeRateUpdateStorage {
	return &ExchangeRateUpdateStorage{
		config: config,
	}
}

const getOrCreateRateUpdateSql = `
WITH new_update AS (
	MERGE INTO exchange_rate_update
	USING (VALUES ($1, $2, $3, 0)) AS update(id, from_currency, to_currency, status)
	ON exchange_rate_update.from_currency = update.from_currency 
		AND exchange_rate_update.to_currency = update.to_currency 
		AND exchange_rate_update.status = update.status
	WHEN NOT MATCHED THEN INSERT (id, from_currency, to_currency, status) 
		VALUES (update.id, update.from_currency, update.to_currency, update.status)
	RETURNING exchange_rate_update.id AS id
)
SELECT id FROM exchange_rate_update 
WHERE from_currency = $2 AND to_currency = $3 AND status = 0
UNION ALL
SELECT id FROM new_update
`

func (storage *ExchangeRateUpdateStorage) GetOrCreateRateUpdate(updateId string, from string, to string) (*model.ExchangeRateUpdateDbo, error) {
	db, err := sql.Open("postgres", storage.config.PostgresConnectionString)
	if err != nil {
		return nil, err
	}

	stmt, err := db.PrepareContext(context.Background(), getOrCreateRateUpdateSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), updateId, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()

	update := model.ExchangeRateUpdateDbo{
		FromCurrency: from,
		ToCurrency:   to,
		Status:       model.StatusUpdating,
	}
	err = rows.Scan(&update.Id)
	return &update, err
}

const getRateUpdateSql = `
SELECT from_currency, to_currency, status, rate_value, update_time FROM exchange_rate_update
WHERE id = $1
`

func (storage *ExchangeRateUpdateStorage) GetRateUpdate(updateId string) (*model.ExchangeRateUpdateDbo, error) {
	db, err := sql.Open("postgres", storage.config.PostgresConnectionString)
	if err != nil {
		return nil, err
	}

	stmt, err := db.PrepareContext(context.Background(), getRateUpdateSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), updateId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasData := rows.Next()
	if !hasData {
		return nil, internal.NewNotFoundError("update not found")
	}

	update := model.ExchangeRateUpdateDbo{Id: updateId}

	err = rows.Scan(&update.FromCurrency, &update.ToCurrency, &update.Status, &update.RateValue, &update.UpdateTime)
	return &update, err
}
