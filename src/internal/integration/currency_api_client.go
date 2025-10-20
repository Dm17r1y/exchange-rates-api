package integration

import (
	"encoding/json"
	"exchange-rates-service/src/config"
	"fmt"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
)

type CurrencyApiClient struct {
	client *http.Client
}

func NewCurrencyApiClient(config *config.Config) ExchangeRateApiClient {
	return &CurrencyApiClient{
		client: &http.Client{
			Timeout: config.HttpClientTimeout,
		},
	}
}

const currencyApiBaseUrl = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies"

func (c *CurrencyApiClient) GetRate(from string, to string) (decimal.Decimal, error) {
	fromLower := strings.ToLower(from)
	toLower := strings.ToLower(to)

	fullUrl := fmt.Sprintf("%s/%s.json", currencyApiBaseUrl, fromLower)
	resp, err := c.client.Get(fullUrl)
	if err != nil {
		return decimal.Decimal{}, err
	}
	defer resp.Body.Close()

	responseMap := make(map[string]any)
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	err = dec.Decode(&responseMap)
	if err != nil {
		return decimal.Decimal{}, err
	}

	fromRates, ok := responseMap[fromLower]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("expected %s in response, got %v", fromLower, responseMap)
	}

	fromRatesMap, ok := fromRates.(map[string]any)
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("cannot deserialize to map: %v", fromRates)
	}

	rate, ok := fromRatesMap[toLower]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("rate %s not found in %v", toLower, fromRatesMap)
	}

	rateNumber, ok := rate.(json.Number)
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("rate %v is not a number", rate)
	}

	return decimal.NewFromString(rateNumber.String())
}
