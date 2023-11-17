package repositories

import "net/http"

type (
	CurrencyRatesRepository interface {
		GetLatestRates() (*FixerIoResponse, error)
		GetLatestRatesWithCurrencyList(currencies []string) (*FixerIoResponse, error)
	}

	fixerIoRespository struct {
		baseUrl    string
		apiKey     string
		httpClient *http.Client
	}
)
