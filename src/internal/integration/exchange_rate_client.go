package integration

import (
	"encoding/json"
	"errors"
	"exchange-rates-service/src/config"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type ExchangeRateApiClient interface {
	GetRate(from string, to string) (decimal.Decimal, error)
}

type ExchangeRateClient struct {
	config *config.Config
	client *http.Client
}

type ExchangeRateApiResponse struct {
	Success   bool                       `json:"success"`
	Timestamp uint64                     `json:"timestamp"`
	Base      string                     `json:"base"`
	Rates     map[string]decimal.Decimal `json:"rates"`
}

func NewExchangeRateApiClient(config *config.Config) ExchangeRateApiClient {
	return &ExchangeRateClient{
		config: config,
		client:  &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

const apiBaseUrl = "https://api.exchangeratesapi.io"

func (c *ExchangeRateClient) GetRate(from string, to string) (decimal.Decimal, error) {
	apiKey := c.config.ExchangeIoApiKey
	fullUrl := fmt.Sprintf("%s/v1/latest?access_key=%s&base=%s&symbols=%s", apiBaseUrl, apiKey, from, to)

	resp, err := c.client.Get(fullUrl)
	if err != nil {
		return decimal.Decimal{}, err
	}
	defer resp.Body.Close()

	response := ExchangeRateApiResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return decimal.Decimal{}, err
	}

	if !response.Success {
		return decimal.Decimal{}, errors.New("error executing request")
	}

	rate, ok := response.Rates[to]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("response not returned for %s rate", to)
	}

	return rate, nil
}
