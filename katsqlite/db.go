package katsqlite

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
)

type DBLink struct {
	*sql.DB
	cfg *katapp.DatabaseConfig
}

// Connect connects to SQLite database and returns DBLink reference with established connection
func Connect(ctx context.Context, cfg *katapp.DatabaseConfig) (*DBLink, error) {
	logger := katapp.Logger(ctx).WithGroup("katsqlite.Connect").With("file", cfg.File)
	logger.Infof("opening database file")
	db, err := sql.Open("sqlite3", cfg.File)
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to database", "error", err)
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	err = db.Ping()
	if err != nil {
		logger.ErrorContext(ctx, "error probing database connection", "error", err)
		return nil, fmt.Errorf("error probing database connection: %w", err)
	}
	logger.Info("connection to database was successfully established")

	return &DBLink{DB: db, cfg: cfg}, nil
}

func (db *DBLink) Close() {
	err := db.DB.Close()
	if err != nil {
		slog.Warn("error closing database connection", "error", err)
	}
}
