package katpg

import (
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

func GetOptionalRow[T any](ctx context.Context, tx pgx.Tx, sql string, args ...any) (*T, error) {
	rows, _ := tx.Query(ctx, sql, args...)
	record, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[T])
	if IsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func GetRowsTx[T any](ctx context.Context, tx pgx.Tx, sql string, args ...any) ([]*T, error) {
	rows, _ := tx.Query(ctx, sql, args...)
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[T])
}
