package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moguchev/transaction_manager"
	pgxv5_transaction_manager "github.com/moguchev/transaction_manager/postgres/pgxv5"
)

// ===================================================================================

// repo - repository
type repo struct {
	driver pgxv5_transaction_manager.QueryEngineProvider // <-- add this one
}

func (r *repo) Foo(ctx context.Context, v int) error {
	db := r.driver.GetQueryEngine(ctx) // <-- instead of pgx.Pool/pgx.Tx direct call

	_, err := db.Exec(ctx, "INSERT INTO test_table_1 (id) VALUES ($1)", v)

	return err
}

func (r *repo) Bar(ctx context.Context, v int) error {
	db := r.driver.GetQueryEngine(ctx) // <-- instead of pgx.Pool/pgx.Tx direct call

	_, err := db.Exec(ctx, "INSERT INTO test_table_2 (id) VALUES ($1)", v)

	return err
}

// ===================================================================================

type repoInterface interface {
	Foo(ctx context.Context, v int) error
	Bar(ctx context.Context, v int) error
}

// example - any business logic
type example struct {
	transactionManager transaction_manager.TransactionManager // <-- add this one

	repo repoInterface
}

// Create - some usecase
func (ex *example) Create(ctx context.Context, id int) error {
	// ...

	transaction := func(txCtx context.Context) error { // Begin of Transaction Scope
		// Foo & Bar will be called in one transaction
		// (if repo supports transaction manager)

		if err := ex.repo.Foo(txCtx, id); err != nil {
			return fmt.Errorf("foo: %w", err)
		}

		if err := ex.repo.Bar(txCtx, id); err != nil {
			return fmt.Errorf("foo: %w", err)
		}

		return nil
	} // End of Transaction Scope

	return ex.transactionManager.RunTransaction(ctx, transaction,
		pgxv5_transaction_manager.WithAccessMode(pgx.ReadWrite),
		pgxv5_transaction_manager.WithIsoLevel(pgx.ReadCommitted),
		pgxv5_transaction_manager.WithDeferrableMode(pgx.Deferrable),
	)
}

// ===================================================================================

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	txManager := pgxv5_transaction_manager.New(pool)

	repo := &repo{
		driver: txManager,
	}

	ex := example{
		repo:               repo,
		transactionManager: txManager,
	}

	if err := ex.Create(ctx, 1); err != nil {
		log.Fatal(err)
	}
	log.Println("OK")
}
