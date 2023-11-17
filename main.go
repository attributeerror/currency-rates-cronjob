//go:build linux && amd64 && cgo
// +build linux,amd64,cgo

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/attributeerror/currency-rates-cronjob/database"
	"github.com/attributeerror/currency-rates-cronjob/repositories"
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

	fixerIoKey, err := loadEnvVar("FIXERIO_KEY", WithIsRequired(true))
	if err != nil {
		if (errors.Is(err, EnvVarNotFoundError{})) {
			panic(err)
		}

		panic(fmt.Errorf("an error occurred whilst trying to load environment variable %s: %w", "FIXERIO_KEY", err))
	}

	fixerIoRepository, err := repositories.InitFixerIoRepository("http://data.fixer.io/api", *fixerIoKey, nil)
	if err != nil {
		panic(fmt.Errorf("error whilst initialising fixer.io repository: %w", err))
	}

	if err := run(database, fixerIoRepository); err != nil {
		panic(fmt.Errorf("error during runtime: %w", err))
	}
}

func initDatabase() (*database.TursoDatabase, error) {
	primaryUrl, err := loadEnvVar("TURSO_URL", WithIsRequired(true))
	if err != nil {
		return nil, err
	}
	authToken, err := loadEnvVar("TURSO_AUTH_TOKEN", WithIsRequired(true))
	if err != nil {
		return nil, err
	}

	dbName, _ := loadEnvVar("TURSO_DB_NAME", WithIsRequired(false), WithDefaultValue("currency-rates"))

	database, err := database.InitTursoDatabase(*primaryUrl, *authToken, *dbName, 1*time.Minute)
	if err != nil {
		return nil, err
	}

	return database, err
}

func run(db database.Database, fixerIoRepository repositories.CurrencyRatesRepository) (err error) {
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

	success, err := db.SetCurrencyRates(latestResponse.Rates)
	if err != nil {
		return err
	}

	stopwatch.Stop()

	if success {
		fmt.Printf("database records updated in %dms\n", stopwatch.Milliseconds())
	} else {
		fmt.Println("unknown error occurred whilst updating database records...")
	}

	stopwatch = stopwatch.Start()

	if tursoDb, ok := db.(*database.TursoDatabase); ok {
		err = tursoDb.SyncEmbeddedReplica()

		if err != nil {
			return err
		}
	}

	stopwatch.Stop()
	fmt.Printf("synced embedded replica in %dms", stopwatch.Milliseconds())

	stopwatch = stopwatch.Start()

	currencyRates, err := db.GetCurrencyRates()
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

func loadEnvVar(envVar string, options ...EnvVarOption) (*string, error) {
	opts := &EnvVarOptions{}
	for _, option := range options {
		option(opts)
	}

	if value, exists := os.LookupEnv(envVar); exists {
		return &value, nil
	} else if opts.Required {
		return nil, &EnvVarNotFoundError{
			envVar,
		}
	} else if opts.DefaultValue != "" {
		return &opts.DefaultValue, nil
	}

	return nil, nil
}
