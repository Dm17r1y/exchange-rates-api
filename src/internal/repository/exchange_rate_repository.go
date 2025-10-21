package repository

import (
	"context"
	"database/sql"
	"exchange-rates-service/src/internal/model"
	"exchange-rates-service/src/internal/storage"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"time"
)

type ExchangeRateRepository interface {
	GetOrCreateRateUpdate(from string, to string) (string, error)
	GetRateUpdate(updateId string) (model.ExchangeRate, error)
	GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error)
	SetUpdateError(updateId string) error
	UpdateRate(updateId string, from string, to string, rate decimal.Decimal) error
	GetLastRate(from string, to string) (model.ExchangeRate, error)
}

type PostgresExchangeRateRepository struct {
	db            *sql.DB
	rateStorage   storage.RateStorage
	updateStorage storage.UpdateStorage
}

func NewExchangeRateRepository(
	db *sql.DB,
	rateStorage storage.RateStorage,
	rateUpdateStorage storage.UpdateStorage) *PostgresExchangeRateRepository {
	repository := PostgresExchangeRateRepository{
		db:            db,
		rateStorage:   rateStorage,
		updateStorage: rateUpdateStorage,
	}
	return &repository
}

func (r *PostgresExchangeRateRepository) GetOrCreateRateUpdate(from string, to string) (string, error) {
	updateId := uuid.New()
	update, err := r.updateStorage.GetOrCreateRateUpdate(updateId.String(), from, to)
	if err != nil {
		return "", err
	}
	return update.Id, nil
}

func (r *PostgresExchangeRateRepository) GetRateUpdate(updateId string) (model.ExchangeRate, error) {
	update, err := r.updateStorage.GetRateUpdate(updateId)
	if err != nil {
		return model.ExchangeRate{}, err
	}

	if update.Status != model.StatusDone {
		return model.ExchangeRate{}, nil
	}

	rate := model.ExchangeRate{
		Rate:           update.RateValue,
		UpdateDateTime: update.UpdateTime,
	}

	return rate, nil
}

func (r *PostgresExchangeRateRepository) GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error) {
	return r.updateStorage.GetRatesForUpdate(fetchSize)
}

func (r *PostgresExchangeRateRepository) SetUpdateError(updateId string) error {
	return r.updateStorage.SetError(updateId)
}

func (r *PostgresExchangeRateRepository) UpdateRate(updateId string, from string, to string, rate decimal.Decimal) error {

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateTime := time.Now().UTC()

	updateRateDbo := model.ExchangeRateUpdateDbo{
		Id:           updateId,
		FromCurrency: from,
		ToCurrency:   to,
		Status:       model.StatusDone,
		UpdateTime:   &updateTime,
		RateValue:    &rate,
	}

	rateDbo := model.ExchangeRateDbo{
		FromCurrency: from,
		ToCurrency:   to,
		RateValue:    &rate,
		UpdateTime:   &updateTime,
	}

	if err := r.updateStorage.UpdateRateTx(tx, &updateRateDbo); err != nil {
		return err
	}

	if err := r.rateStorage.SetRateTx(tx, &rateDbo); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresExchangeRateRepository) GetLastRate(from string, to string) (model.ExchangeRate, error) {
	rate, err := r.rateStorage.GetRate(from, to)

	if err != nil {
		return model.ExchangeRate{}, err
	}

	if rate == nil {
		return model.ExchangeRate{}, nil
	}

	resultRate := model.ExchangeRate{
		Rate:           rate.RateValue,
		UpdateDateTime: rate.UpdateTime,
	}

	return resultRate, nil
}
