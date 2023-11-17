package fixerio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type (
	FixerIoResponse struct {
		// Success is whether this request was successful
		Success bool `json:"success,omitempty"`
		// CollectedSeconds is the unix timestamp (in seconds) of when the rates were collected
		CollectedSeconds int64 `json:"timestamp,omitempty"`
		// Base is the currency code of the base conversion rate for the listed rates - likely always EUR
		Base string `json:"base,omitempty"`
		// Date is the human-readable date in DD-MM-YYYY format for whne the rates were collected.
		Date string `json:"date,omitempty"`
		// Rates is the key-value map, mapped by currency code, of the current rates
		Rates map[string]float64 `json:"rates,omitempty"`
	}

	FixerIoRepository interface {
		GetLatestRates() (*FixerIoResponse, error)
		GetLatestRatesWithCurrencyList(currencies []string) (*FixerIoResponse, error)
	}

	fixerIoRespository struct {
		baseUrl    string
		apiKey     string
		httpClient *http.Client
	}
)

func isStatusOk(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func InitFixerIoRepository(baseUrl, apiKey string, httpClient *http.Client) (*fixerIoRespository, error) {
	if baseUrl == "" {
		return nil, errors.New("'baseUrl' must not be null or empty")
	}
	if apiKey == "" {
		return nil, errors.New("'apiKey' must not be null or empty")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &fixerIoRespository{
		baseUrl:    baseUrl,
		apiKey:     apiKey,
		httpClient: httpClient,
	}, nil
}

func (fxr *fixerIoRespository) GetLatestRates() (*FixerIoResponse, error) {
	fixerIoResponse := &FixerIoResponse{}

	reqUrl := fmt.Sprintf("%s/latest?access_key=%s&base=EUR", fxr.baseUrl, fxr.apiKey)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error whilst initialising new request: %w", err)
	}

	res, err := fxr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error whilst fetching data from Fixer.io: %w", err)
	}
	defer res.Body.Close()

	resBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error whilst reading response body: %w", err)
	}

	if !isStatusOk(res.StatusCode) {
		return nil, fmt.Errorf("response returned non-2xx status code %s: %s", res.Status, string(resBodyBytes))
	}

	err = json.Unmarshal(resBodyBytes, fixerIoResponse)
	if err != nil {
		return nil, fmt.Errorf("error whilst unmarshalling JSON: %w", err)
	}

	return fixerIoResponse, nil
}

func (fxr *fixerIoRespository) GetLatestRatesWithCurrencyList(currencies []string) (*FixerIoResponse, error) {
	fixerIoResponse := &FixerIoResponse{}

	currenciesString := strings.Join(currencies, ", ")
	reqUrl := fmt.Sprintf("%s/latest?access_key=%s&base=EUR&symbols=%s", fxr.baseUrl, fxr.apiKey, currenciesString)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error whilst initialising new request: %w", err)
	}

	res, err := fxr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error whilst fetching data from Fixer.io: %w", err)
	}
	defer res.Body.Close()

	resBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error whilst reading response body: %w", err)
	}

	if !isStatusOk(res.StatusCode) {
		return nil, fmt.Errorf("response returned non-2xx status code %s: %s", res.Status, string(resBodyBytes))
	}

	err = json.Unmarshal(resBodyBytes, fixerIoResponse)
	if err != nil {
		return nil, fmt.Errorf("error whilst unmarshalling JSON: %w", err)
	}

	return fixerIoResponse, nil
}
