package psql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionKey struct{}

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

func TxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(transactionKey{}).(pgx.Tx)
	return tx
}

func QuerierFromContext(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}

	return pool
}
