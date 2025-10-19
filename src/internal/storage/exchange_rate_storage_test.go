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

func TestGetRate_Success(t *testing.T) {
	storage, _, mock := createRateMockStorage(t)

	from, to, rateValue, updateTime := "USD", "EUR", "12345", time.Now()
	rows := sqlmock.NewRows([]string{"rate_value", "update_time"}).AddRow(rateValue, updateTime)

	mock.ExpectPrepare(regexp.QuoteMeta(getRateSql)).
		ExpectQuery().
		WithArgs(from, to).
		WillReturnRows(rows)

	rate, err := storage.GetRate(from, to)

	assert.NoError(t, err)
	assert.Equal(t, from, rate.FromCurrency)
	assert.Equal(t, to, rate.ToCurrency)

	expectedRateValue, _ := decimal.NewFromString(rateValue)
	assert.Equal(t, &expectedRateValue, rate.RateValue)
	assert.Equal(t, &updateTime, rate.UpdateTime)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRate_NotFound(t *testing.T) {
	storage, _, mock := createRateMockStorage(t)

	from, to := "USD", "EUR"
	rows := sqlmock.NewRows([]string{"rate_value", "update_time"})

	mock.ExpectPrepare(regexp.QuoteMeta(getRateSql)).
		ExpectQuery().
		WithArgs(from, to).
		WillReturnRows(rows)

	rate, err := storage.GetRate(from, to)
	assert.Nil(t, rate)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func SetRateTx_Success(t *testing.T) {
	storage, db, mock := createRateMockStorage(t)
	rateValue, updateTime := decimal.NewFromFloat(123.45), time.Now()

	dbo := model.ExchangeRateDbo{
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		RateValue:    &rateValue,
		UpdateTime:   &updateTime,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta(setRateSql)).
		ExpectExec().
		WithArgs(dbo.FromCurrency, dbo.ToCurrency, dbo.RateValue, dbo.UpdateTime).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	err = storage.SetRateTx(tx, &dbo)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func createRateMockStorage(t *testing.T) (ExchangeRateStorage, *sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	storage := NewExchangeRateStorage(db)
	return storage, db, mock
}
