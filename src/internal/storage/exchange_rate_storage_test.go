package storage

import (
	"exchange-rates-service/src/internal/model"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

func TestGetRate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateStorage(db)

	from, to, rateValue, updateTime := "USD", "EUR", "12345", time.Now()
	rows := sqlmock.NewRows([]string{"rate_value", "update_time"}).AddRow(rateValue, updateTime)

	mock.ExpectPrepare(regexp.QuoteMeta(getRateSql)).
		ExpectQuery().
		WithArgs(from, to).
		WillReturnRows(rows)

	rate, err := storage.GetRate(from, to)

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rate.FromCurrency != from {
		t.Fatalf("from currency: expected %s but got %s", from, rate.FromCurrency)
	}
	if rate.ToCurrency != to {
		t.Fatalf("to currency: expected %s but got %s", to, rate.ToCurrency)
	}
	expectedRateValue, _ := decimal.NewFromString(rateValue)
	if !rate.RateValue.Equal(expectedRateValue) {
		t.Fatalf("rate value: expected %v but got %v", expectedRateValue, rate.RateValue)
	}
	if !rate.UpdateTime.Equal(updateTime) {
		t.Fatalf("update time: expected %v but got %v", updateTime, rate.UpdateTime)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %s", err)
	}
}

func TestGetRate_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()
	storage := NewExchangeRateStorage(db)

	from, to := "USD", "EUR"
	rows := sqlmock.NewRows([]string{"rate_value", "update_time"})

	mock.ExpectPrepare(regexp.QuoteMeta(getRateSql)).
		ExpectQuery().
		WithArgs(from, to).
		WillReturnRows(rows)

	rate, err := storage.GetRate(from, to)
	if rate != nil || err == nil {
		t.Fatalf("expected not found error, got rate=%v err=%v", rate, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func SetRateTx_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateStorage(db)

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
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	if err := storage.SetRateTx(tx, &dbo); err != nil {
		t.Fatalf("error on SetRateTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("error on commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
