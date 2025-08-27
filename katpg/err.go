package katpg

import (
	"context"
	"errors"
	"strings"

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
			if strings.Contains(pgerr.Message, "update") || strings.Contains(pgerr.Message, "delete") {
				if strings.Contains(pgerr.Detail, "is not present") {
					return katapp.NewErr(katapp.ErrNotFound, title+": referenced record not found")
				} else {
					return katapp.NewErr(katapp.ErrConflict, ": record is in use by other records")
				}
			} else {
				return katapp.NewErr(katapp.ErrNotFound, title+": referenced record not found")
			}
		default:
			return katapp.NewErr(katapp.ErrInternal, title+": unknown error")
		}
	}
	return katapp.NewErr(katapp.ErrInternal, title+": unknown error")
}

func PgToAppErrorContext(ctx context.Context, err error, title string) *katapp.Err {
	var pgerr *pgconn.PgError
	var details string
	if errors.As(err, &pgerr) {
		details = pgerr.Detail
	}
	katapp.Logger(ctx).Error("katpg.LoggedPgToAppError", "title", title, "error", err, "details", details)
	return PgToAppError(err, title)
}

func IsNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}
