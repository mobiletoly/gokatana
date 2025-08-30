package katpg

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/katapp"
)

type DBLink struct {
	*pgxpool.Pool
	cfg *katapp.DatabaseConfig
}

// MustConnect connects to PostgreSQL database and returns DBLink reference with established connection
func MustConnect(ctx context.Context, cfg *katapp.DatabaseConfig) *DBLink {
	db, err := Connect(ctx, cfg)
	if err != nil {
		katapp.Logger(ctx).Fatalf("failed to establish mandatory db connection: %v", err)
	}
	return db
}

// Connect connects to PostgreSQL database and returns DBLink reference with established connection
func Connect(ctx context.Context, cfg *katapp.DatabaseConfig) (*DBLink, error) {
	logger := katapp.Logger(ctx)
	escapedUser := url.QueryEscape(cfg.User)
	escapedPassword := url.QueryEscape(cfg.Password)
	escapedName := url.QueryEscape(cfg.Name)

	parameters := url.Values{}
	for k, v := range cfg.Parameters {
		parameters.Add(k, v)
	}

	URL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=%d&%s",
		escapedUser, escapedPassword, cfg.Host, cfg.Port, escapedName, cfg.Sslmode, cfg.ConnectTimeout,
		parameters.Encode())

	logger.Infof("establishing connection to database "+
		"postgres://%s:<password>@%s:%d/%s?sslmode=%s&connect_timeout=%d",
		escapedUser, cfg.Host, cfg.Port, escapedName, cfg.Sslmode, cfg.ConnectTimeout)

	pcfg, err := pgxpool.ParseConfig(URL)
	if err != nil {
		logger.ErrorContext(ctx, "failed to parse pool config", "error", err)
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}
	pcfg.MaxConns = int32(cfg.Pool.MaxConns)
	pcfg.MinConns = int32(cfg.Pool.MinConns)
	pcfg.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	pcfg.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	pcfg.HealthCheckPeriod = cfg.Pool.HealthPeriod

	initCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(initCtx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(initCtx, 4*time.Second)
	defer pingCancel()
	err = pool.Ping(pingCtx)
	if err != nil {
		pool.Close()
		pool = nil
		return nil, logger.LogNewError("error probing database connection: %w", err)
	}

	logger.Infof("connection to database was successfully established")

	return &DBLink{Pool: pool, cfg: cfg}, nil
}

func (db *DBLink) Close() {
	db.Pool.Close()
}

func (db *DBLink) MustRunScript(ctx context.Context, filename string) {
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read script file %s: %w", filename, err)
		}
		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute script %s: %w", filename, err)
		}
		return nil
	})
	if err != nil {
		log.Panicln("failed to run scripts:", err)
	}
}
