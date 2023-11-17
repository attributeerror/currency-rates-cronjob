package database

import (
	"context"
	"database/sql"

	"github.com/libsql/go-libsql"
)

type Database interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(context context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(context context.Context, query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(context context.Context, query string, args ...any) (sql.Result, error)
	GetCurrencyRates() (map[string]float64, error)
	SetCurrencyRates(currencyRates map[string]float64) (bool, error)
	Close() error
}

type (
	TursoDatabase struct {
		Db             *sql.DB
		connector      *libsql.Connector
		localDirectory string
		tableName      string
	}
)
