package pgxv5

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moguchev/transaction_manager"
)

var (
	_ transaction_manager.TransactionManager = (*TransactionManager)(nil)
)

// TransactionManager - transaction manager: allows you to perform
// the functions of different repositories included in one database
// within a transaction.
type TransactionManager struct {
	pool *pgxpool.Pool
}

// New returns transaction_manager.TransactionManager interface implementation
func New(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// WithIsoLevel - set transaction IsoLevel
func WithIsoLevel(lvl pgx.TxIsoLevel) transaction_manager.TransactionOption {
	return func(x any) {
		if opts, ok := x.(*pgx.TxOptions); ok {
			opts.IsoLevel = lvl
		}
	}
}

// WithAccessMode - set transaction AccessMode
func WithAccessMode(mode pgx.TxAccessMode) transaction_manager.TransactionOption {
	return func(x any) {
		if opts, ok := x.(*pgx.TxOptions); ok {
			opts.AccessMode = mode
		}
	}
}

// WithDeferrableMode - set transaction DeferrableMode
func WithDeferrableMode(mode pgx.TxDeferrableMode) transaction_manager.TransactionOption {
	return func(x any) {
		if opts, ok := x.(*pgx.TxOptions); ok {
			opts.DeferrableMode = mode
		}
	}
}

// RunReadCommitted - execute function fn in runTransaction
func (m *TransactionManager) RunTransaction(
	ctx context.Context,
	fn func(txCtx context.Context) error,
	opts ...transaction_manager.TransactionOption,
) error {
	var txOptions pgx.TxOptions
	for _, opt := range opts {
		opt(&txOptions)
	}

	return m.runTransaction(ctx, txOptions, fn)
}

type key string

const txKey key = "tx"

func (m *TransactionManager) runTransaction(ctx context.Context, txOpts pgx.TxOptions, fn func(txCtx context.Context) error) (err error) {
	// If it's nested Transaction, skip initiating a new one and return func(ctx context.Context) error
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	// Begin transaction
	tx, err := m.pool.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("transaction_manager: can't begin transaction: %v", err)
	}

	// Set transaction to context
	txCtx := context.WithValue(ctx, txKey, tx)

	// Set up a defer function for rolling back the runTransaction.
	defer func() {
		// recover from panic
		if r := recover(); r != nil {
			err = fmt.Errorf("transaction_manager: panic recovered: %v", r)
		}

		// if func(ctx context.Context) error didn't return error - commit
		if err == nil {
			// if commit returns error -> rollback
			err = tx.Commit(ctx)
			if err != nil {
				err = fmt.Errorf("transaction_manager: commit failed: %v", err)
			}
		}

		// rollback on any error
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = fmt.Errorf("transaction_manager: rollback failed: %v", errRollback)
			}
		}
	}()

	// Execute the code inside the runTransaction. If the function
	// fails, return the error and the defer function will roll back or commit otherwise.
	return fn(txCtx)
}

// GetQueryEngine provides QueryEngine
func (m *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	// Transaction always runs on node with NodeRoleWrite role
	if tx, ok := ctx.Value(txKey).(QueryEngine); ok {
		return tx
	}

	return m.pool
}
