package sqlxutils

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func sqlErr(err error, query string, args ...interface{}) error {
	return errors.Wrapf(err, `run query "%s" with args %+v`, query, args)
}

func namedQuery(query string, arg interface{}) (nq string, args []interface{}, err error) {
	nq, args, err = sqlx.Named(query, arg)
	if err != nil {
		return "", nil, sqlErr(err, query, args...)
	}
	return nq, args, nil
}

func Exec(ctx context.Context, db sqlx.ExecerContext, query string, args ...interface{}) (sql.Result, error) {
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return res, sqlErr(err, query, args...)
	}

	return res, nil
}

func NamedExec(ctx context.Context, db sqlx.ExtContext, query string, arg interface{}) (sql.Result, error) {
	nq, args, err := namedQuery(query, arg)
	if err != nil {
		return nil, err
	}

	return Exec(ctx, db, db.Rebind(nq), args...)
}

func Select(ctx context.Context, db sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) error {
	if err := sqlx.SelectContext(ctx, db, dest, query, args...); err != nil {
		return sqlErr(err, query, args...)
	}

	return nil
}

func NamedSelect(ctx context.Context, db sqlx.ExtContext, dest interface{}, query string, arg interface{}) error {
	nq, args, err := namedQuery(query, arg)
	if err != nil {
		return err
	}

	return Select(ctx, db, dest, db.Rebind(nq), args...)
}

func Get(ctx context.Context, db sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) error {
	if err := sqlx.GetContext(ctx, db, dest, query, args...); err != nil {
		return sqlErr(err, query, args...)
	}

	return nil
}

func NamedGet(ctx context.Context, db sqlx.ExtContext, dest interface{}, query string, arg interface{}) error {
	nq, args, err := namedQuery(query, arg)
	if err != nil {
		return err
	}

	return Get(ctx, db, dest, db.Rebind(nq), args...)
}

type txFunc func(tx *sqlx.Tx) error

type txRunner interface {
	BeginTxx(context.Context, *sql.TxOptions) (*sqlx.Tx, error)
}

func RunTx(ctx context.Context, db txRunner, level sql.IsolationLevel, f txFunc) (err error) {
	var tx *sqlx.Tx

	tx, err = db.BeginTxx(ctx, &sql.TxOptions{Isolation: level})
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}

	defer func() {
		if err != nil {
			err = multierr.Combine(err, tx.Rollback())
		} else {
			err = tx.Commit()
		}
	}()

	return f(tx)
}
