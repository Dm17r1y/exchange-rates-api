package service

import (
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetOrCreateRateUpdate(from string, to string) (string, error) {
	args := m.Called(from, to)
	return args.String(0), args.Error(1)
}

func (m *mockRepository) GetRateUpdate(updateId string) (model.ExchangeRate, error) {
	args := m.Called(updateId)
	return args.Get(0).(model.ExchangeRate), args.Error(1)
}

func (m *mockRepository) GetRatesForUpdate(fetchSize int) ([]model.ExchangeRateUpdateDbo, error) {
	args := m.Called(fetchSize)
	return args.Get(0).([]model.ExchangeRateUpdateDbo), args.Error(1)
}

func (m *mockRepository) SetUpdateError(updateId string) error {
	args := m.Called(updateId)
	return args.Error(0)
}

func (m *mockRepository) UpdateRate(updateId string, from string, to string, rate decimal.Decimal) error {
	args := m.Called(updateId, from, to, rate)
	return args.Error(0)
}

func (m *mockRepository) GetLastRate(from string, to string) (model.ExchangeRate, error) {
	args := m.Called(from, to)
	return args.Get(0).(model.ExchangeRate), args.Error(1)
}

func TestStartUpdateRate_ThrowsErrorWhenUnknownCurrency(t *testing.T) {
	service := createMockService()

	updateId, err := service.StartUpdateRate("UNKNOWN", "USD")
	assert.Equal(t, updateId, "")
	assert.Error(t, err)
	assert.Equal(t, err.(*internal.ServiceError).ErrorType, internal.BadRequest)
}

func TestStartUpdateRate_ThrowsErrorOnConvertingSameCurrency(t *testing.T) {
	service := createMockService()

	updateId, err := service.StartUpdateRate("USD", "USD")
	assert.Equal(t, updateId, "")
	assert.Error(t, err)
	assert.Equal(t, err.(*internal.ServiceError).ErrorType, internal.BadRequest)
}

func TestGetLastRate_ThrowsErrorWhenUnknownCurrency(t *testing.T) {
	service := createMockService()

	_, err := service.GetLastRate("UNKNOWN", "USD")
	assert.Error(t, err)
	assert.Equal(t, err.(*internal.ServiceError).ErrorType, internal.BadRequest)
}

func TestGetLastRate_ThrowsErrorOnConvertingSameCurrency(t *testing.T) {
	service := createMockService()

	_, err := service.GetLastRate("EUR", "EUR")
	assert.Error(t, err)
	assert.Equal(t, err.(*internal.ServiceError).ErrorType, internal.BadRequest)
}

func createMockService() *RateService {
	mockRepo := new(mockRepository)
	return NewRateService(mockRepo)

}
