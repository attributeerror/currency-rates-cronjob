package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/libsql/go-libsql"
)

func InitTursoDatabase(primaryUrl, authToken, dbName string, syncInterval time.Duration) (*TursoDatabase, error) {
	// ensure the 'TursoDatabase' struct matches the 'Database' interface at compile time
	// this doesn't use up any extra memory as I'm using the discard (_) operator
	// this means the variable gets disposed as soon as its created
	var _ Database = (*TursoDatabase)(nil)

	dbInst := &TursoDatabase{
		tableName: "currency_rates",
	}

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	dbInst.localDirectory = dir

	connector, err := libsql.NewEmbeddedReplicaConnectorWithAutoSync(dir+"/"+dbName+".db", primaryUrl, authToken, syncInterval)
	if err != nil {
		return nil, fmt.Errorf("error whilst connecting to database: %w", err)
	}

	dbInst.connector = connector
	dbInst.Db = sql.OpenDB(connector)

	return dbInst, nil
}

func (db *TursoDatabase) Query(query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *TursoDatabase) QueryContext(context context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.Db.QueryContext(context, query, args...)
}

func (db *TursoDatabase) QueryRow(query string, args ...any) (*sql.Rows, error) {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *TursoDatabase) QueryRowContext(context context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.Db.QueryContext(context, query, args...)
}

func (db *TursoDatabase) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *TursoDatabase) ExecContext(context context.Context, query string, args ...any) (sql.Result, error) {
	return db.Db.ExecContext(context, query, args...)
}

func (db *TursoDatabase) SyncEmbeddedReplica() error {
	return db.connector.Sync()
}

func (db *TursoDatabase) GetCurrencyRates() (map[string]float64, error) {
	currencyRates := make(map[string]float64)

	sqlStmt := fmt.Sprintf("SELECT code, printf('%%.5f', to_euro_rate) FROM %s", db.tableName)
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error whilst querying 'currency_rates' table: %w", err)
	}

	defer func() {
		if closeError := rows.Close(); closeError != nil {
			fmt.Println("error closing rows", closeError)
			if err == nil {
				err = closeError
			}
		}
	}()

	for rows.Next() {
		var code string
		var toEuroRate string
		err = rows.Scan(&code, &toEuroRate)
		if err != nil {
			return nil, err
		}

		euroFloat, err := strconv.ParseFloat(toEuroRate, 64)
		if err != nil {
			return nil, fmt.Errorf("error whilst parsing float: %w", err)
		}
		currencyRates[code] = euroFloat
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return currencyRates, nil
}

func (db *TursoDatabase) SetCurrencyRates(currencyRates map[string]float64) (bool, error) {
	for code, eurRate := range currencyRates {
		eurRate = (eurRate * 100) / 100

		sqlStmt := fmt.Sprintf("SELECT 1 FROM %s WHERE code = '%s'", db.tableName, code)
		var exists int32
		err := db.Db.QueryRow(sqlStmt).Scan(&exists)
		if err != nil {
			if err != sql.ErrNoRows {
				return false, fmt.Errorf("error whilst querying %s for rows with code %s: %w", db.tableName, code, err)
			} else {
				// currency code record not found, insert new one
				sqlStmt = fmt.Sprintf("INSERT INTO %s (code, to_euro_rate) VALUES ('%s', %s)", db.tableName, code, fmt.Sprintf("%.6f", eurRate))
				_, err := db.Exec(sqlStmt)
				if err != nil {
					return false, fmt.Errorf("error whilst executing INSERT statement: %w", err)
				}
			}
		} else {
			// currency code record can be found, update existing
			sqlStmt = fmt.Sprintf("UPDATE %s SET to_euro_rate = %s, last_update_date = CURRENT_TIMESTAMP WHERE code = '%s'", db.tableName, fmt.Sprintf("%.6f", eurRate), code)
			_, err := db.Exec(sqlStmt)
			if err != nil {
				return false, fmt.Errorf("error whilst executing UPDATE statement: %w", err)
			}
		}
	}

	return true, nil
}

func (db *TursoDatabase) Close() error {
	if sqlCloseError := db.Db.Close(); sqlCloseError != nil {
		return fmt.Errorf("error whilst closing database: %w", sqlCloseError)
	}

	if connectorCloseError := db.connector.Close(); connectorCloseError != nil {
		return fmt.Errorf("error whilst closing connector: %w", connectorCloseError)
	}

	if deleteTempDirError := os.RemoveAll(db.localDirectory); deleteTempDirError != nil {
		return fmt.Errorf("error whilst deleting temporary files: %w", deleteTempDirError)
	}

	return nil
}
