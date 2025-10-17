package storage

import (
	"exchange-rates-service/src/internal/model"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

func TestGetOrCreateRateUpdate_ShouldSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

	updateId, from, to := "test-update-id", "USD", "EUR"
	rows := sqlmock.NewRows([]string{"id"}).AddRow(updateId)

	mock.ExpectPrepare(regexp.QuoteMeta(getOrCreateRateUpdateSql)).
		ExpectQuery().
		WithArgs(updateId, from, to).
		WillReturnRows(rows)

	update, err := storage.GetOrCreateRateUpdate(updateId, from, to)

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if update.Id != updateId {
		t.Fatalf("update id: expected %s but got %s", updateId, update.Id)
	}
	if update.FromCurrency != from {
		t.Fatalf("from currency: expected %s but got %s", from, update.FromCurrency)
	}
	if update.ToCurrency != to {
		t.Fatalf("to currency: expected %s but got %s", to, update.ToCurrency)
	}
	if update.Status != model.StatusUpdating {
		t.Fatalf("status: expected %v but got %v", model.StatusUpdating, update.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %s", err)
	}
}

func TestGetRateUpdate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

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

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if update.Id != updateId {
		t.Fatalf("update id: expected %s but got %s", updateId, update.Id)
	}
	if update.FromCurrency != from {
		t.Fatalf("from currency: expected %s but got %s", from, update.FromCurrency)
	}
	if update.ToCurrency != to {
		t.Fatalf("to currency: expected %s but got %s", to, update.ToCurrency)
	}
	if update.Status != status {
		t.Fatalf("status: expected %v but got %v", status, update.Status)
	}
	if !update.RateValue.Equal(rateValue) {
		t.Fatalf("rate value: expected %v but got %v", rateValue, update.RateValue)
	}
	if !update.UpdateTime.Equal(updateTime) {
		t.Fatalf("update time: expected %v but got %v", updateTime, update.UpdateTime)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %s", err)
	}
}

func TestGetRateUpdate_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

	updateId := "non-existent-id"
	rows := sqlmock.NewRows([]string{"from_currency", "to_currency", "status", "rate_value", "update_time"})

	mock.ExpectPrepare(regexp.QuoteMeta(getRateUpdateSql)).
		ExpectQuery().
		WithArgs(updateId).
		WillReturnRows(rows)

	update, err := storage.GetRateUpdate(updateId)

	if update != nil || err == nil {
		t.Fatalf("expected not found error, got update=%v err=%v", update, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
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
		WithArgs(fetchSize).
		WillReturnRows(rows)

	updates, err := storage.GetRatesForUpdate(fetchSize)

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(updates) != 3 {
		t.Fatalf("expected 3 updates but got %d", len(updates))
	}
	if updates[0].Id != "update-1" || updates[0].FromCurrency != "USD" || updates[0].ToCurrency != "EUR" {
		t.Fatalf("unexpected first update: %v", updates[0])
	}
	if updates[1].Id != "update-2" || updates[1].FromCurrency != "EUR" || updates[1].ToCurrency != "USD" {
		t.Fatalf("unexpected second update: %v", updates[1])
	}
	if updates[2].Id != "update-3" || updates[2].FromCurrency != "EUR" || updates[2].ToCurrency != "MXN" {
		t.Fatalf("unexpected third update: %v", updates[2])
	}
	for _, update := range updates {
		if update.Status != model.StatusUpdating {
			t.Fatalf("expected status %v but got %v", model.StatusUpdating, update.Status)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %s", err)
	}
}

func TestGetRatesForUpdate_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

	fetchSize := 10
	rows := sqlmock.NewRows([]string{"id", "from_currency", "to_currency"})

	mock.ExpectPrepare(regexp.QuoteMeta(getRatesForUpdateSql)).
		ExpectQuery().
		WithArgs(fetchSize).
		WillReturnRows(rows)

	updates, err := storage.GetRatesForUpdate(fetchSize)

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(updates) != 0 {
		t.Fatalf("expected empty result but got %d updates", len(updates))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %s", err)
	}
}

func TestUpdateRateTx_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

	updateId := "test-update-id"
	rateValue := decimal.NewFromFloat(1.2345)
	updateTime := time.Now()
	status := model.StatusDone

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
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	if err := storage.UpdateRateTx(tx, &updateDbo); err != nil {
		t.Fatalf("error on UpdateRateTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("error on commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetError_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	storage := NewExchangeRateUpdateStorage(db)

	updateId := "test-update-id"

	mock.ExpectPrepare(regexp.QuoteMeta(setErrorSql)).
		ExpectExec().
		WithArgs(updateId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = storage.SetError(updateId)

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %s", err)
	}
}
