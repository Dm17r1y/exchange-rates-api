package repository

import (
	"context"
	"database/sql"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/model"
	"exchange-rates-service/src/internal/storage"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"time"
)

type ExchangeRateRepository struct {
	db *sql.DB
	rateStorage   *storage.ExchangeRateStorage
	updateStorage *storage.ExchangeRateUpdateStorage
}

func NewExchangeRateRepository(config *config.Config) (*ExchangeRateRepository, error) {
	db, err := sql.Open("postgres", config.PostgresConnectionString)
	if err != nil {
		return nil, err
	}
	repository := ExchangeRateRepository{
		db: db,
		rateStorage:   storage.NewExchangeRateStorage(db),
		updateStorage: storage.NewExchangeRateUpdateStorage(db),
	}
	return &repository, nil
}

func (r *ExchangeRateRepository) GetOrCreateRateUpdate(from string, to string) (string, error) {
	updateId := uuid.New()
	update, err := r.updateStorage.GetOrCreateRateUpdate(updateId.String(), from, to)
	if err != nil {
		return "", err
	}
	return update.Id, nil
}

func (r *ExchangeRateRepository) GetRateUpdate(updateId string) (model.ExchangeRate, error) {
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


func (r *ExchangeRateRepository) GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error) {
	return r.updateStorage.GetRatesForUpdate(fetchSize)
}

func (r *ExchangeRateRepository) SetUpdateError(updateId string) error {
	return r.updateStorage.SetError(updateId)
}

func (r *ExchangeRateRepository) UpdateRate(updateId string, from string, to string, rate decimal.Decimal) error {
	
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateTime := time.Now().UTC()

	updateRateDbo := model.ExchangeRateUpdateDbo{
		Id: updateId,
		FromCurrency: from,
		ToCurrency: to,
		Status: model.StatusDone,
		UpdateTime: &updateTime,
		RateValue: &rate,
	}

	rateDbo := model.ExchangeRateDbo{
		FromCurrency: from,
		ToCurrency: to,
		RateValue: &rate,
		UpdateTime: &updateTime,
	}

	if err := r.updateStorage.UpdateRateTx(tx, &updateRateDbo); err != nil {
		return err
	}

	if err := r.rateStorage.SetRateTx(tx, &rateDbo); err != nil {
		return err
	}

	return tx.Commit()
}


func (r *ExchangeRateRepository) GetLastRate(from string, to string) (model.ExchangeRate, error) {
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

