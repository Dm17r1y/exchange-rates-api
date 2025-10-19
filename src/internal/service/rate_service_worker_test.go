package service

import (
	"errors"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/model"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockApiClient struct {
	mock.Mock
}

func (m *mockApiClient) GetRate(from string, to string) (decimal.Decimal, error) {
	args := m.Called(from, to)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func TestExecuteUpdate_ShouldSetErrorWhenReturnedErrorFromApi(t *testing.T) {
	mockRepo, mockClient, worker := createMocks()

	rateUpdate := model.ExchangeRateUpdateDbo{
		Id:           "update-id-1",
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Status:       model.StatusUpdating,
	}

	mockRepo.On("GetRatesForUpdate", 10).Return([]model.ExchangeRateUpdateDbo{rateUpdate}, nil)
	mockClient.On("GetRate", rateUpdate.FromCurrency, rateUpdate.ToCurrency).
		Return(decimal.Decimal{}, errors.New("api error"))
	mockRepo.On("SetUpdateError", rateUpdate.Id).Return(nil)

	count, err := worker.ExecuteUpdate()

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestExecuteUpdate_ShouldReturnErrorWhenWeHaveProblemWithRepository(t *testing.T) {
	mockRepo, mockClient, worker := createMocks()

	rateUpdate := model.ExchangeRateUpdateDbo{
		Id:           "update-id-1",
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Status:       model.StatusUpdating,
	}

	rate := decimal.NewFromFloat(1.18)
	repositoryError := errors.New("database error")

	mockRepo.On("GetRatesForUpdate", 10).Return([]model.ExchangeRateUpdateDbo{rateUpdate}, nil)
	mockClient.On("GetRate", rateUpdate.FromCurrency, rateUpdate.ToCurrency).Return(rate, nil)
	mockRepo.On("UpdateRate", rateUpdate.Id, rateUpdate.FromCurrency, rateUpdate.ToCurrency, rate).Return(repositoryError)

	count, err := worker.ExecuteUpdate()

	assert.Error(t, err)
	assert.Equal(t, repositoryError, err)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestExecuteUpdate_ShouldReturnUpdateCountBeforeError(t *testing.T) {
	mockRepo, mockClient, worker := createMocks()

	update1 := model.ExchangeRateUpdateDbo{
		Id:           "update-id-1",
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Status:       model.StatusUpdating,
	}

	update2 := model.ExchangeRateUpdateDbo{
		Id:           "update-id-2",
		FromCurrency: "EUR",
		ToCurrency:   "MXN",
		Status:       model.StatusUpdating,
	}

	update3 := model.ExchangeRateUpdateDbo{
		Id:           "update-id-3",
		FromCurrency: "MXN",
		ToCurrency:   "USD",
		Status:       model.StatusUpdating,
	}

	mockRepo.On("GetRatesForUpdate", 10).Return([]model.ExchangeRateUpdateDbo{update1, update2, update3}, nil)

	rate1 := decimal.NewFromFloat(1.18)
	mockClient.On("GetRate", update1.FromCurrency, update1.ToCurrency).Return(rate1, nil)
	mockRepo.On("UpdateRate", update1.Id, update1.FromCurrency, update1.ToCurrency, rate1).Return(nil)

	rate2 := decimal.NewFromFloat(1.35)
	mockClient.On("GetRate", update2.FromCurrency, update2.ToCurrency).Return(rate2, nil)
	mockRepo.On("UpdateRate", update2.Id, update2.FromCurrency, update2.ToCurrency, rate2).Return(nil)

	rate3 := decimal.NewFromFloat(155.23)
	repositoryError := errors.New("database error")
	mockClient.On("GetRate", update3.FromCurrency, update3.ToCurrency).Return(rate3, nil)
	mockRepo.On("UpdateRate", update3.Id, update3.FromCurrency, update3.ToCurrency, rate3).Return(repositoryError)

	count, err := worker.ExecuteUpdate()

	assert.Error(t, err)
	assert.Equal(t, repositoryError, err)
	assert.Equal(t, 2, count)
	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func createMocks() (*mockRepository, *mockApiClient, *RateServiceWorker) {
	mockRepo := new(mockRepository)
	mockClient := new(mockApiClient)
	config := &config.Config{WorkerFetchSize: 10}
	worker := &RateServiceWorker{
		config:     config,
		repository: mockRepo,
		client:     mockClient,
	}

	return mockRepo, mockClient, worker

}
