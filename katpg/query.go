package katpg

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

func GetOptionalRow[T any](ctx context.Context, db *pgxpool.Pool, sql string, args ...any) (*T, error) {
	rows, _ := db.Query(ctx, sql, args...)
	record, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[T])
	if IsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func GetRows[T any](ctx context.Context, db *pgxpool.Pool, sql string, args ...any) ([]*T, error) {
	rows, _ := db.Query(ctx, sql, args...)
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[T])
}
