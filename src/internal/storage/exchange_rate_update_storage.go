package storage

import (
	"context"
	"database/sql"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"
)

type SqlExchangeRateUpdateStorage struct {
	db *sql.DB
}

type ExchangeRateUpdateStorage interface {
	GetOrCreateRateUpdate(updateId string, from string, to string) (*model.ExchangeRateUpdateDbo, error)
	GetRateUpdate(updateId string) (*model.ExchangeRateUpdateDbo, error)
	GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error)
	UpdateRateTx(tx *sql.Tx, model *model.ExchangeRateUpdateDbo) error
	SetError(updateId string) error
}

func NewExchangeRateUpdateStorage(db *sql.DB) ExchangeRateUpdateStorage {
	return &SqlExchangeRateUpdateStorage{db: db}
}

const getOrCreateRateUpdateSql = `
WITH new_update AS (
	MERGE INTO exchange_rate_update
	USING (VALUES ($1, $2, $3, $4)) AS update(id, from_currency, to_currency, status)
	ON exchange_rate_update.from_currency = update.from_currency 
		AND exchange_rate_update.to_currency = update.to_currency 
		AND exchange_rate_update.status = update.status
	WHEN NOT MATCHED THEN INSERT (id, from_currency, to_currency, status) 
		VALUES (update.id, update.from_currency, update.to_currency, update.status)
	RETURNING exchange_rate_update.id AS id
)
SELECT id FROM exchange_rate_update 
WHERE from_currency = $2 AND to_currency = $3 AND status = $4
UNION ALL
SELECT id FROM new_update
`

func (storage *SqlExchangeRateUpdateStorage) GetOrCreateRateUpdate(updateId string, from string, to string) (*model.ExchangeRateUpdateDbo, error) {
	stmt, err := storage.db.PrepareContext(context.Background(), getOrCreateRateUpdateSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), updateId, from, to, model.StatusUpdating)
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
SELECT from_currency, to_currency, status, rate_value, update_time 
FROM exchange_rate_update
WHERE id = $1
`

func (storage *SqlExchangeRateUpdateStorage) GetRateUpdate(updateId string) (*model.ExchangeRateUpdateDbo, error) {
	stmt, err := storage.db.PrepareContext(context.Background(), getRateUpdateSql)
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

const getRatesForUpdateSql = `
SELECT id, from_currency, to_currency
FROM exchange_rate_update
WHERE status = $2
LIMIT $1
`

func (storage *SqlExchangeRateUpdateStorage) GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error) {
	stmt, err := storage.db.PrepareContext(context.Background(), getRatesForUpdateSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), fetchSize, model.StatusUpdating)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbos := make([]model.ExchangeRateUpdateDbo, 0, fetchSize)
	for rows.Next() {
		update := model.ExchangeRateUpdateDbo{
			Status: 0,
		}

		if err := rows.Scan(&update.Id, &update.FromCurrency, &update.ToCurrency); err != nil {
			return nil, err
		}

		dbos = append(dbos, update)
	}

	return dbos, nil
}

const updateRateSql = `
UPDATE exchange_rate_update 
SET rate_value = $2, update_time = $3, status = $4
WHERE id = $1
`

func (storage *SqlExchangeRateUpdateStorage) UpdateRateTx(tx *sql.Tx, model *model.ExchangeRateUpdateDbo) error {
	stmt, err := tx.PrepareContext(context.Background(), updateRateSql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(context.Background(), model.Id, model.RateValue, model.UpdateTime, model.Status)
	return err
}

const setErrorSql = `
UPDATE exchange_rate_update
SET status = $2
WHERE id = $1
`

func (storage *SqlExchangeRateUpdateStorage) SetError(updateId string) error {
	stmt, err := storage.db.PrepareContext(context.Background(), setErrorSql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(context.Background(), updateId, model.StatusError)
	return err
}
