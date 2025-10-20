package integration

import (
	"encoding/json"
	"errors"
	"exchange-rates-service/src/config"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type ExchangeRateApiClient interface {
	GetRate(from string, to string) (decimal.Decimal, error)
}

type ExchangeRateApiIoClient struct {
	config *config.Config
	client *http.Client
}

type ExchangeRateApiIoResponse struct {
	Success   bool                       `json:"success"`
	Timestamp uint64                     `json:"timestamp"`
	Base      string                     `json:"base"`
	Rates     map[string]decimal.Decimal `json:"rates"`
}

func NewExchangeRateApiIoClient(config *config.Config) ExchangeRateApiClient {
	return &ExchangeRateApiIoClient{
		config: config,
		client: &http.Client{
			Timeout: config.HttpClientTimeout,
		},
	}
}

const exchangeRatesApiIoBaseUrl = "https://api.exchangeratesapi.io"

func (c *ExchangeRateApiIoClient) GetRate(from string, to string) (decimal.Decimal, error) {
	apiKey := c.config.ExchangeIoApiKey
	fullUrl := fmt.Sprintf("%s/v1/latest?access_key=%s&base=%s&symbols=%s", exchangeRatesApiIoBaseUrl, apiKey, from, to)

	resp, err := c.client.Get(fullUrl)
	if err != nil {
		return decimal.Decimal{}, err
	}
	defer resp.Body.Close()

	response := ExchangeRateApiIoResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return decimal.Decimal{}, err
	}

	if !response.Success {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return decimal.Decimal{}, err
		}

		return decimal.Decimal{}, errors.New(string(bodyBytes))
	}

	rate, ok := response.Rates[to]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("response not returned for %s rate", to)
	}

	return rate, nil
}
