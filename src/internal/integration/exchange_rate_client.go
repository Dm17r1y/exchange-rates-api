package integration

import (
	"exchange-rates-service/src/config"

	"github.com/shopspring/decimal"
)

type ExchangeRateClient struct {
}

type ExchangeRateClientResponse struct {
	Value decimal.Decimal
}

func NewExchangeRateClient(config *config.Config) *ExchangeRateClient {
	return &ExchangeRateClient{}
}

func (c *ExchangeRateClient) GetRates(from string, to string) (ExchangeRateClientResponse, error) {
	value, err := decimal.NewFromString("19.99")
	if err != nil {
		return ExchangeRateClientResponse{}, err
	}

	return ExchangeRateClientResponse{
		Value: value,
	}, nil
}
