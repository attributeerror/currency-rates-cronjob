//go:build linux && amd64 && cgo
// +build linux,amd64,cgo

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/attributeerror/currency-rates-cronjob/database"
	"github.com/attributeerror/currency-rates-cronjob/fixerio"
	"github.com/bradhe/stopwatch"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		panic(fmt.Errorf("error whilst opening .env: %w", err))
	}

	database, err := initDatabase()
	if err != nil {
		panic(fmt.Errorf("error whilst initialising database: %w", err))
	}
	defer database.Close()

	fixerIoRepository, err := fixerio.InitFixerIoRepository("http://data.fixer.io/api", os.Getenv("FIXERIO_KEY"), nil)
	if err != nil {
		panic(fmt.Errorf("error whilst initialising fixer.io repository: %w", err))
	}

	if err := run(database, fixerIoRepository); err != nil {
		panic(fmt.Errorf("error during runtime: %w", err))
	}
}

func initDatabase() (*database.TursoDatabase, error) {
	primaryUrl := os.Getenv("TURSO_URL")
	if primaryUrl == "" {
		return nil, errors.New("TURSO_URL environment variable is not set")
	}
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if authToken == "" {
		return nil, errors.New("TURSO_AUTH_TOKEN environment variable is not set")
	}

	database, err := database.InitTursoDatabase(primaryUrl, authToken, "currency-rates", 1*time.Minute)
	if err != nil {
		return nil, err
	}

	return database, err
}

func run(database database.Database, fixerIoRepository fixerio.FixerIoRepository) (err error) {
	needToUpdate := false

	if needToUpdate {
		stopwatch := stopwatch.Start()

		latestResponse, err := fixerIoRepository.GetLatestRates()
		if err != nil {
			return err
		}

		stopwatch.Stop()

		fmt.Printf("--- FIXER.IO RESPONSE (fetched in %dms) ---\n", stopwatch.Milliseconds())
		for key, val := range latestResponse.Rates {
			fmt.Println("1 EUR ->", key, "=", val, key)
		}

		stopwatch = stopwatch.Start()

		success, err := database.SetCurrencyRates(latestResponse.Rates)
		if err != nil {
			return err
		}

		stopwatch.Stop()

		if success {
			fmt.Printf("database records updated in %dms\n", stopwatch.Milliseconds())
		} else {
			fmt.Println("unknown error occurred whilst updating database records...")
		}
	}

	stopwatch := stopwatch.Start()

	currencyRates, err := database.GetCurrencyRates()
	if err != nil {
		return err
	}

	stopwatch.Stop()

	fmt.Printf("\n--- DATABASE RECORDS (fetched in %dms) ---\n", stopwatch.Milliseconds())

	for key, val := range currencyRates {
		fmt.Printf("%s -> %f\n", key, val)
	}

	return nil
}
