# Golang Transaction Manager [![Go Report Card](https://goreportcard.com/badge/github.com/moguchev/transaction_manager?style=flat-square)](https://goreportcard.com/report/github.com/moguchev/transaction_manager) [![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/moguchev/transaction_manager) [![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/golang-standards/project-layout/releases/latest)


Transaction manager is an abstraction to coordinate database transaction boundaries.

Easiest way to get the perfect repository.

## Supported implementations

- [pgx/v5](https://github.com/jackc/pgx)

## Installation

```sh
go get github.com/moguchev/transaction_manager
```

## Usage

### Examples

- [pgx/v5](./examples/postgres/pgxv5/main.go)

Below is an example how to start usage.

1. Implement a Transaction Manager at your Repository Layer

```go
import pgxv5_transaction_manager "github.com/moguchev/transaction_manager/postgres/pgxv5"

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
```

2. Use Transaction Manager in your code

```go
import "github.com/moguchev/transaction_manager"

// repoInterface - your dependency
type repoInterface interface {
	Foo(ctx context.Context, v int) error
	Bar(ctx context.Context, v int) error
}


// example - your business logic
type example struct {
	transactionManager transaction_manager.TransactionManager // <-- add this one

	repo repoInterface
}

// Create - some usecase...
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
        // additional options for out transaction without any influence on our code
        // (easy to mock)
		pgxv5_transaction_manager.WithAccessMode(pgx.ReadWrite),
		pgxv5_transaction_manager.WithIsoLevel(pgx.ReadCommitted),
		pgxv5_transaction_manager.WithDeferrableMode(pgx.Deferrable),
	)
}
```

3. Initialize Transaction Manager and use dependency injection to your objects
```go
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
```

## Other projects

- [avito-tech/go-transaction-manager](https://github.com/avito-tech/go-transaction-manager)