package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type TxFunc func(tx pgx.Tx) error

func (db *DB) WithTransaction(ctx context.Context, fn TxFunc) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) WithSerializableTx(ctx context.Context, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
