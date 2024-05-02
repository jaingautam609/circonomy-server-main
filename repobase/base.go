package repobase

import (
	"circonomy-server/dbutil"
	"circonomy-server/utils"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Base interface {
	Select(dest interface{}, query string, args ...interface{}) error
	SelectWithContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	GetWithContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (res sql.Result, err error)
	QueryRowX(query string, args ...interface{}) *sqlx.Row
	ExecC(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error)
	ExecMustAffect(targetAffectedCount int, query string, args ...interface{}) error
	ExecCMustAffect(ctx context.Context, targetAffectedCount int, query string, args ...interface{}) error
	ExecErrorOnly(query string, args ...interface{}) error
	ExecCErrorOnly(ctx context.Context, query string, args ...interface{}) error
	DB() *sqlx.DB
	CopyWithTX(tx *sqlx.Tx) Base
	IsTransaction() bool
}

var ErrWrongNumberOfAffectedRows = fmt.Errorf("incorrect number of affected rows")

type base struct {
	_dbInstance   *sqlx.DB
	db            dbutil.Sqlxer
	queryPrefixer func(ctx context.Context, query string, args ...interface{}) string
}

func NewBase(db *sqlx.DB) Base {
	return &base{
		_dbInstance: db,
		db:          db,
	}
}

func NewBaseWithPrefixer(db *sqlx.DB, queryPrefixer func(ctx context.Context, query string, args ...interface{}) string) Base {
	return &base{
		_dbInstance:   db,
		db:            db,
		queryPrefixer: queryPrefixer,
	}
}

func (b *base) DB() *sqlx.DB {
	return b._dbInstance
}

func (b *base) CopyWithTX(tx *sqlx.Tx) Base {
	repoCopy := *b
	repoCopy.db = tx
	return &repoCopy
}

func (b *base) IsTransaction() bool {
	_, ok := b.db.(*sqlx.Tx)
	return ok
}

func (b *base) Select(dest interface{}, query string, args ...interface{}) error {
	err := b.db.Select(dest, query, args...)
	// returns nil if the error is nil
	return utils.SQLErrorLogger(err, query, args...)
}

func (b *base) SelectWithContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if b.queryPrefixer != nil {
		query = b.queryPrefixer(ctx, query, args...)
	}
	err := b.db.SelectContext(ctx, dest, query, args...)
	// returns nil if the error is nil
	return utils.SQLErrorLogger(err, query, args...)
}

func (b *base) Get(dest interface{}, query string, args ...interface{}) error {
	err := b.db.Get(dest, query, args...)
	// returns nil if the error is nil
	return utils.SQLErrorLogger(err, query, args...)
}

func (b *base) GetWithContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if b.queryPrefixer != nil {
		query = b.queryPrefixer(ctx, query, args...)
	}
	err := b.db.GetContext(ctx, dest, query, args...)
	// returns nil if the error is nil
	return utils.SQLErrorLogger(err, query, args...)
}

func (b *base) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	res, err = b.db.Exec(query, args...)
	// returns nil if the error is nil
	err = utils.SQLErrorLogger(err, query, args...)
	return
}

func (b *base) QueryRowX(query string, args ...interface{}) *sqlx.Row {
	return b.db.QueryRowx(query, args...)
}

func (b *base) ExecC(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	if b.queryPrefixer != nil {
		query = b.queryPrefixer(ctx, query, args...)
	}
	res, err = b.db.ExecContext(ctx, query, args...)
	// returns nil if the error is nil
	err = utils.SQLErrorLogger(err, query, args...)
	return
}

func (b *base) ExecMustAffect(targetAffectedCount int, query string, args ...interface{}) error {
	res, err := b.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if int(rowsAffected) != targetAffectedCount {
		return fmt.Errorf("got %d, expected %d %w", rowsAffected, targetAffectedCount, ErrWrongNumberOfAffectedRows)
	}
	return nil
}

func (b *base) ExecCMustAffect(ctx context.Context, targetAffectedCount int, query string, args ...interface{}) error {
	res, err := b.ExecC(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if int(rowsAffected) != targetAffectedCount {
		return fmt.Errorf("got %d, expected %d %w", rowsAffected, targetAffectedCount, ErrWrongNumberOfAffectedRows)
	}
	return nil
}

func (b *base) ExecErrorOnly(query string, args ...interface{}) error {
	_, err := b.Exec(query, args...)
	return err
}

func (b *base) ExecCErrorOnly(ctx context.Context, query string, args ...interface{}) error {
	_, err := b.ExecC(ctx, query, args...)
	return err
}
