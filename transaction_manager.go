package transaction_manager

import "context"

// TransactionOption - any transaction option
type TransactionOption func(x any)

// TransactionManager - transaction manager interface
type TransactionManager interface {
	RunTransaction(ctx context.Context, fn func(txCtx context.Context) error, opts ...TransactionOption) error
}
