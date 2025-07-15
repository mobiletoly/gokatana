package katsqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mobiletoly/gokatana/internal"
	"github.com/mobiletoly/gokatana/katapp"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const migrationTable = "_kat_migration"

var validMigrationFilePattern = regexp.MustCompile(`^(?:[A-Za-z]+_)?\d{4}_.+\.sql$`)

type migrationRecord struct {
	Service       string
	UpdatedAt     time.Time
	LatestVersion string
}

// DoMigration performs all migrations for the registered configurations.
func (db *DBLink) DoMigration(ctx context.Context) error {
	mgrTime := time.Now()
	for _, mgr := range db.cfg.Migrations {
		logger := katapp.Logger(ctx).WithGroup("katsqlite.DoMigration").
			With("service", mgr.Service)
		logger.Infof("performing SQLite migration")
		files, err := internal.ListFilesInDirSortedByFilename(mgr.Path)
		if err != nil {
			return logger.LogNewError("failed to list migration files in directory %s: %w", mgr.Path, err)
		}

		tx, err := db.DB.BeginTx(ctx, nil)
		if err != nil {
			return logger.LogNewError("failed to begin transaction: %w", err)
		}

		err = func(tx *sql.Tx) error {
			if err := createMigrationTableIfNotExist(ctx, tx); err != nil {
				return err
			}
			mgrRec, err := loadMigrationRecord(ctx, tx, mgr.Service)
			if err != nil {
				return err
			}

			var startFilePrefix string
			if mgrRec == nil {
				startFilePrefix = ""
			} else {
				startFilePrefix = mgrRec.LatestVersion
			}

			var latestVersion string
			for _, file := range files {
				_, f := filepath.Split(file)
				if !validMigrationFilePattern.MatchString(f) {
					logger.Warnf("skipping incorrectly named migration file: %s", file)
					continue
				}
				if f[:4] <= startFilePrefix {
					continue
				}
				content, err := os.ReadFile(file)
				if err != nil {
					return logger.LogNewError("failed to read migration file (%s): %w", file, err)
				}
				if _, err = tx.ExecContext(ctx, string(content)); err != nil {
					return logger.LogNewError("failed to execute migration file (%s): %w", file, err)
				}
				logger.Infof("successfully applied migration file: %s", file)
				latestVersion = f[:4]
			}

			if latestVersion != "" {
				if err := upsertMigrationRecord(ctx, tx, mgr.Service, latestVersion, mgrTime); err != nil {
					return err
				}
			}
			return nil
		}(tx)

		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Errorf("failed to rollback transaction: %v", rollbackErr)
			}
			return fmt.Errorf("migration failed: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return logger.LogNewError("failed to commit transaction: %w", err)
		}

		logger.Infof("migration completed successfully for service: %s", mgr.Service)
	}
	return nil
}

// MustDoMigration ensures migration is performed or exits on failure.
func (db *DBLink) MustDoMigration(ctx context.Context) {
	if err := db.DoMigration(ctx); err != nil {
		katapp.Logger(ctx).Fatalf("migration failed: %v", err)
	}
}

func createMigrationTableIfNotExist(ctx context.Context, tx *sql.Tx) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	service TEXT NOT NULL UNIQUE,
	last_version TEXT NOT NULL
);`, migrationTable)

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return katapp.Logger(ctx).LogNewError("failed to create migration table: %w", err)
	}
	return nil
}

func loadMigrationRecord(ctx context.Context, tx *sql.Tx, service string) (*migrationRecord, error) {
	query := fmt.Sprintf(`
SELECT service, updated_at, last_version
FROM %s
WHERE service = ?
LIMIT 1
`, migrationTable)

	row := tx.QueryRowContext(ctx, query, service)

	var record migrationRecord
	err := row.Scan(&record.Service, &record.UpdatedAt, &record.LatestVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, katapp.Logger(ctx).LogNewError("failed to find migration record for service=%s: %w", service, err)
	}
	return &record, nil
}

func upsertMigrationRecord(ctx context.Context, tx *sql.Tx, service string, version string, time time.Time) error {
	query := fmt.Sprintf(`
INSERT INTO %s (service, updated_at, last_version)
VALUES (?, ?, ?)
ON CONFLICT(service) DO UPDATE SET
    updated_at = excluded.updated_at,
    last_version = excluded.last_version
`, migrationTable)

	_, err := tx.ExecContext(ctx, query, service, time, version)
	if err != nil {
		return katapp.Logger(ctx).LogNewError("failed to upsert migration record for service=%s: %w", service, err)
	}
	return nil
}
