package pgxv5

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ QueryEngine = (pgx.Tx)(nil)
	_ QueryEngine = (*pgxpool.Conn)(nil)
)

// QueryEngineProvider - something that gives us QueryEngine
type QueryEngineProvider interface {
	// returns pgx QueryEngine API
	GetQueryEngine(ctx context.Context) QueryEngine
}

// QueryEngine is a common database query interface.
type QueryEngine interface {
	PgxCommonAPI
	PgxExtendedAPI
}

// PgxCommonAPI - pgx common API (like database/sql)
type PgxCommonAPI interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

// PgxExtendedAPI - pgx special API
type PgxExtendedAPI interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}
