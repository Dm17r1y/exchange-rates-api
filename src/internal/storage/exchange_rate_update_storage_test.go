package storage

import (
	"database/sql"
	"exchange-rates-service/src/internal/model"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrCreateRateUpdate_ShouldSuccess(t *testing.T) {
	storage, _, mock := createUpdateMockStorage(t)

	updateId, from, to := "test-update-id", "USD", "EUR"
	rows := sqlmock.NewRows([]string{"id"}).AddRow(updateId)

	mock.ExpectPrepare(regexp.QuoteMeta(getOrCreateRateUpdateSql)).
		ExpectQuery().
		WithArgs(updateId, from, to, model.StatusUpdating).
		WillReturnRows(rows)

	update, err := storage.GetOrCreateRateUpdate(updateId, from, to)

	assert.NoError(t, err)
	assert.Equal(t, updateId, update.Id)
	assert.Equal(t, from, update.FromCurrency)
	assert.Equal(t, to, update.ToCurrency)
	assert.Equal(t, model.StatusUpdating, update.Status)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetRateUpdate_Success(t *testing.T) {
	storage, _, mock := createUpdateMockStorage(t)

	updateId, from, to := "test-update-id", "USD", "EUR"
	rateValue := decimal.NewFromFloat(1.2345)
	updateTime := time.Now()
	status := model.StatusDone

	rows := sqlmock.NewRows([]string{"from_currency", "to_currency", "status", "rate_value", "update_time"}).
		AddRow(from, to, status, rateValue, updateTime)

	mock.ExpectPrepare(regexp.QuoteMeta(getRateUpdateSql)).
		ExpectQuery().
		WithArgs(updateId).
		WillReturnRows(rows)

	update, err := storage.GetRateUpdate(updateId)

	assert.NoError(t, err)
	assert.Equal(t, updateId, update.Id)
	assert.Equal(t, from, update.FromCurrency)
	assert.Equal(t, to, update.ToCurrency)
	assert.Equal(t, status, update.Status)
	assert.Equal(t, &rateValue, update.RateValue)
	assert.Equal(t, &updateTime, update.UpdateTime)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRateUpdate_NotFound(t *testing.T) {
	storage, _, mock := createUpdateMockStorage(t)

	updateId := "non-existent-id"
	rows := sqlmock.NewRows([]string{"from_currency", "to_currency", "status", "rate_value", "update_time"})

	mock.ExpectPrepare(regexp.QuoteMeta(getRateUpdateSql)).
		ExpectQuery().
		WithArgs(updateId).
		WillReturnRows(rows)

	update, err := storage.GetRateUpdate(updateId)

	assert.Nil(t, update)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRatesForUpdate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

	fetchSize := 10
	rows := sqlmock.NewRows([]string{"id", "from_currency", "to_currency"}).
		AddRow("update-1", "USD", "EUR").
		AddRow("update-2", "EUR", "USD").
		AddRow("update-3", "EUR", "MXN")

	mock.ExpectPrepare(regexp.QuoteMeta(getRatesForUpdateSql)).
		ExpectQuery().
		WithArgs(fetchSize, model.StatusUpdating).
		WillReturnRows(rows)

	updates, err := storage.GetRatesForUpdate(fetchSize)
	assert.NoError(t, err)
	assert.Len(t, updates, 3)
	assert.Equal(t, updates[0], model.ExchangeRateUpdateDbo{Id: "update-1", FromCurrency: "USD", ToCurrency: "EUR"})
	assert.Equal(t, updates[1], model.ExchangeRateUpdateDbo{Id: "update-2", FromCurrency: "EUR", ToCurrency: "USD"})
	assert.Equal(t, updates[2], model.ExchangeRateUpdateDbo{Id: "update-3", FromCurrency: "EUR", ToCurrency: "MXN"})

	for _, update := range updates {
		assert.Equal(t, model.StatusUpdating, update.Status)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRatesForUpdate_EmptyResult(t *testing.T) {
	storage, _, mock := createUpdateMockStorage(t)

	fetchSize := 10
	rows := sqlmock.NewRows([]string{"id", "from_currency", "to_currency"})

	mock.ExpectPrepare(regexp.QuoteMeta(getRatesForUpdateSql)).
		ExpectQuery().
		WithArgs(fetchSize, model.StatusUpdating).
		WillReturnRows(rows)

	updates, err := storage.GetRatesForUpdate(fetchSize)

	assert.NoError(t, err)
	assert.Len(t, updates, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRateTx_Success(t *testing.T) {
	storage, db, mock := createUpdateMockStorage(t)

	updateId, rateValue := "test-update-id", decimal.NewFromFloat(1.2345)
	updateTime, status := time.Now(), model.StatusDone

	updateDbo := model.ExchangeRateUpdateDbo{
		Id:         updateId,
		RateValue:  &rateValue,
		UpdateTime: &updateTime,
		Status:     status,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta(updateRateSql)).
		ExpectExec().
		WithArgs(updateId, &rateValue, &updateTime, status).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	err = storage.UpdateRateTx(tx, &updateDbo)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestSetError_Success(t *testing.T) {
	storage, _, mock := createUpdateMockStorage(t)

	updateId := "test-update-id"

	mock.ExpectPrepare(regexp.QuoteMeta(setErrorSql)).
		ExpectExec().
		WithArgs(updateId, model.StatusError).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := storage.SetError(updateId)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func createUpdateMockStorage(t *testing.T) (ExchangeRateUpdateStorage, *sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	storage := NewExchangeRateUpdateStorage(db)
	return storage, db, mock
}
