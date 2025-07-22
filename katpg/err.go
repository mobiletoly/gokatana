package katpg

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mobiletoly/gokatana/katapp"
)

func PgToAppError(err error, title string) *katapp.Err {
	if IsNoRows(err) {
		return katapp.NewErr(katapp.ErrNotFound, title+": not found")
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		switch pgerr.Code {
		case "23505":
			return katapp.NewErr(katapp.ErrDuplicate, title+": duplicate data")
		case "23503":
			return katapp.NewErr(katapp.ErrNotFound, title+": referenced record not found")
		}
	}
	return katapp.NewErr(katapp.ErrInternal, title+": unknown error")
}

func PgToAppErrorContext(ctx context.Context, err error, title string) *katapp.Err {
	katapp.Logger(ctx).Error("katpg.LoggedPgToAppError", "title", title, "error", err)
	return PgToAppError(err, title)
}

func IsNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}
