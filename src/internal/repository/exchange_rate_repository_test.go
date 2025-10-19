package repository

import (
	"database/sql"
	"errors"
	"exchange-rates-service/src/internal/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockExchangeRateStorage struct {
	mock.Mock
}

func (m *MockExchangeRateStorage) GetRate(from string, to string) (*model.ExchangeRateDbo, error) {
	args := m.Called(from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ExchangeRateDbo), args.Error(1)
}

func (m *MockExchangeRateStorage) SetRateTx(tx *sql.Tx, rateDbo *model.ExchangeRateDbo) error {
	args := m.Called(tx, rateDbo)
	return args.Error(0)
}

type MockExchangeRateUpdateStorage struct {
	mock.Mock
}

func (m *MockExchangeRateUpdateStorage) GetOrCreateRateUpdate(updateId string, from string, to string) (*model.ExchangeRateUpdateDbo, error) {
	args := m.Called(updateId, from, to)
	return args.Get(0).(*model.ExchangeRateUpdateDbo), args.Error(1)
}

func (m *MockExchangeRateUpdateStorage) GetRateUpdate(updateId string) (*model.ExchangeRateUpdateDbo, error) {
	args := m.Called(updateId)
	return args.Get(0).(*model.ExchangeRateUpdateDbo), args.Error(1)
}

func (m *MockExchangeRateUpdateStorage) GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error) {
	args := m.Called(fetchSize)
	return args.Get(0).([]model.ExchangeRateUpdateDbo), args.Error(1)
}

func (m *MockExchangeRateUpdateStorage) UpdateRateTx(tx *sql.Tx, updateDbo *model.ExchangeRateUpdateDbo) error {
	args := m.Called(tx, updateDbo)
	return args.Error(0)
}

func (m *MockExchangeRateUpdateStorage) SetError(updateId string) error {
	args := m.Called(updateId)
	return args.Error(0)
}

func TestGetOrCreateRateUpdate_ShouldReturnUpdateIdFromStorage(t *testing.T) {
	_, mockUpdateStorage, repo, _, _ := createMocks(t)

	expectedUpdateId := "update-id-123"
	update := &model.ExchangeRateUpdateDbo{
		Id:           expectedUpdateId,
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Status:       model.StatusUpdating,
	}

	mockUpdateStorage.On("GetOrCreateRateUpdate", mock.AnythingOfType("string"), update.FromCurrency, update.ToCurrency).
		Return(update, nil)

	updateId, err := repo.GetOrCreateRateUpdate(update.FromCurrency, update.ToCurrency)

	assert.NoError(t, err)
	assert.Equal(t, expectedUpdateId, updateId)
	mockUpdateStorage.AssertExpectations(t)
}

func TestGetRateUpdate_Success(t *testing.T) {
	_, mockUpdateStorage, repo, _, _ := createMocks(t)

	updateId := "update-123"
	updateTime := time.Now().UTC()
	rate := decimal.NewFromFloat(1.25)

	updateDbo := &model.ExchangeRateUpdateDbo{
		Id:           updateId,
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Status:       model.StatusDone,
		RateValue:    &rate,
		UpdateTime:   &updateTime,
	}

	mockUpdateStorage.On("GetRateUpdate", updateId).Return(updateDbo, nil)

	result, err := repo.GetRateUpdate(updateId)

	assert.NoError(t, err)
	assert.Equal(t, &rate, result.Rate)
	assert.Equal(t, &updateTime, result.UpdateDateTime)
	mockUpdateStorage.AssertExpectations(t)
}

func TestGetRateUpdate_ReturnsEmptyWhenStatusNotDone(t *testing.T) {
	_, mockUpdateStorage, repo, _, _ := createMocks(t)

	updateId := "update-123"
	updateDbo := &model.ExchangeRateUpdateDbo{
		Id:           updateId,
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Status:       model.StatusUpdating,
	}

	mockUpdateStorage.On("GetRateUpdate", updateId).Return(updateDbo, nil)

	result, err := repo.GetRateUpdate(updateId)

	assert.NoError(t, err)
	assert.Nil(t, result.Rate)
	assert.Nil(t, result.UpdateDateTime)
	mockUpdateStorage.AssertExpectations(t)
}

func TestUpdateRate_Success(t *testing.T) {
	mockRateStorage, mockUpdateStorage, repo, _, sqlMock := createMocks(t)

	rate := decimal.NewFromFloat(1.35)
	updateId := "update-123"
	fromCurrency := "USD"
	toCurrency := "EUR"	

	sqlMock.ExpectBegin()

	mockUpdateStorage.On("UpdateRateTx", mock.AnythingOfType("*sql.Tx"), mock.MatchedBy(func(dbo *model.ExchangeRateUpdateDbo) bool {
		return dbo.Id == updateId &&
			dbo.FromCurrency == fromCurrency &&
			dbo.ToCurrency == toCurrency &&
			dbo.Status == model.StatusDone &&
			dbo.RateValue.Equal(rate)
	})).Return(nil)

	mockRateStorage.On("SetRateTx", mock.AnythingOfType("*sql.Tx"), mock.MatchedBy(func(dbo *model.ExchangeRateDbo) bool {
		return dbo.FromCurrency == fromCurrency &&
			dbo.ToCurrency == toCurrency &&
			dbo.RateValue.Equal(rate)
	})).Return(nil)

	sqlMock.ExpectCommit()

	err := repo.UpdateRate(updateId, fromCurrency, toCurrency, rate)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	mockUpdateStorage.AssertExpectations(t)
	mockRateStorage.AssertExpectations(t)
}

func TestUpdateRate_ShouldRollbackWhenError(t *testing.T) {
	_, mockUpdateStorage, repo, _, sqlMock := createMocks(t)

	rate := decimal.NewFromFloat(1.35)
	expectedError := errors.New("error")

	sqlMock.ExpectBegin()

	mockUpdateStorage.On("UpdateRateTx", mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(expectedError)

	sqlMock.ExpectRollback()

	err := repo.UpdateRate("update-123", "USD", "EUR", rate)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	mockUpdateStorage.AssertExpectations(t)
}

func TestUpdateRate_ShouldRollbackWhenSetRateError(t *testing.T) {
	mockRateStorage, mockUpdateStorage, repo, _, sqlMock := createMocks(t)

	rate := decimal.NewFromFloat(1.35)
	expectedError := errors.New("set rate error")

	sqlMock.ExpectBegin()

	mockUpdateStorage.On("UpdateRateTx", mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(nil)

	mockRateStorage.On("SetRateTx", mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(expectedError)

	sqlMock.ExpectRollback()

	err := repo.UpdateRate("update-123", "USD", "EUR", rate)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	mockUpdateStorage.AssertExpectations(t)
	mockRateStorage.AssertExpectations(t)
}

func TestGetLastRate_Success(t *testing.T) {
	mockRateStorage, _, repo, _, _ := createMocks(t)

	fromCurrency := "USD"
	toCurrency := "EUR"
	updateTime := time.Now().UTC()
	rateValue := decimal.NewFromFloat(1.45)

	rateDbo := &model.ExchangeRateDbo{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		RateValue:    &rateValue,
		UpdateTime:   &updateTime,
	}

	mockRateStorage.On("GetRate", fromCurrency, toCurrency).Return(rateDbo, nil)

	result, err := repo.GetLastRate(fromCurrency, toCurrency)

	assert.NoError(t, err)
	assert.Equal(t, &rateValue, result.Rate)
	assert.Equal(t, &updateTime, result.UpdateDateTime)
	mockRateStorage.AssertExpectations(t)
}

func TestGetLastRate_ReturnsEmptyWhenRateNotFound(t *testing.T) {
	mockRateStorage, _, repo, _, _ := createMocks(t)

	from := "USD"
	to := "EUR"
	mockRateStorage.On("GetRate", from, to).Return(nil, nil)

	result, err := repo.GetLastRate(from, to)

	assert.NoError(t, err)
	assert.Nil(t, result.Rate)
	assert.Nil(t, result.UpdateDateTime)
	mockRateStorage.AssertExpectations(t)
}


func createMocks(t *testing.T) (
	*MockExchangeRateStorage, 
	*MockExchangeRateUpdateStorage, 
	*ExchangeRateRepository, 
	*sql.DB, 
	sqlmock.Sqlmock) {
	mockUpdateStorage := new(MockExchangeRateUpdateStorage)
	mockRateStorage := new(MockExchangeRateStorage)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewExchangeRateRepository(db, mockRateStorage, mockUpdateStorage)
	return mockRateStorage, mockUpdateStorage, repo, db, mock
}
