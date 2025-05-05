package katpg

import (
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mobiletoly/gokatana/katapp"
)

func ToAppError(err error, title string) *katapp.Err {
	if IsNoRows(err) {
		return katapp.NewErr(katapp.ErrNotFound, title+": not found")
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		switch pgerr.Code {
		case "23505":
			return katapp.NewErr(katapp.ErrDuplicate, title+": duplicate data")
		}
	}
	return katapp.NewErr(katapp.ErrInternal, title+": unknown error")
}

func IsNoRows(err error) bool {
	return err != nil && errors.Is(err, pgx.ErrNoRows)
}
