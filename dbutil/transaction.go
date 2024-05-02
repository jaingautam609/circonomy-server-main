package dbutil

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// A TxFn is a function that will be called with an initialized `Transaction` object
// that can be used for executing statements and queries against a database.
type TxFn func(tx *sqlx.Tx) error

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
// nolint: gocritic // no need to lint
func WithTransaction(db *sqlx.DB, fn TxFn) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and re-panic
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			_ = tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

type Sqlxer interface {
	sqlx.Queryer
	sqlx.Execer
	sqlx.ExecerContext
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
