package katpg

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/internal"
	"github.com/mobiletoly/gokatana/katapp"
)

const migrationTable = "_kat_migration"

var validMigrationFilePattern = regexp.MustCompile(`^(?:[A-Za-z]+_)?\d{4}_.+\.sql$`)

type migrationRecord struct {
	Service       string    `db:"service"`
	UpdatedAt     time.Time `db:"updated_at"`
	LatestVersion string    `db:"last_version"`
}

type migrationFile struct {
	name        string
	displayPath string
	readPath    string
	sourceFS    fs.FS
}

// DoMigration performs all migrations for the provided PostgreSQL pool.
func DoMigration(ctx context.Context, pgpool *pgxpool.Pool, migrations []katapp.DatabaseMigrationConfig) error {
	if pgpool == nil {
		return fmt.Errorf("migration failed: pgpool is nil")
	}

	mgrTime := time.Now()
	for _, mgr := range migrations {
		logger := katapp.Logger(ctx).WithGroup("katpg.DoMigration").
			With("service", mgr.Service).
			With("schema", mgr.Schema)
		logger.Infof("perform db migration")
		files, err := listMigrationFiles(mgr)
		if err != nil {
			return logger.LogNewError("failed to list migration files in directory %s: %w", mgr.Path, err)
		}
		err = pgx.BeginFunc(ctx, pgpool, func(tx pgx.Tx) error {
			if err := createMigrationTableIfNotExist(ctx, tx, mgr.Schema); err != nil {
				return err
			}
			mgrRec, err := loadMigrationRecord(ctx, tx, mgr.Schema, mgr.Service)
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
				if !validMigrationFilePattern.MatchString(file.name) {
					logger.Warnf("skipping incorrectly named migration file: %s", file.displayPath)
					continue
				}
				if file.name[:4] <= startFilePrefix {
					continue
				}
				content, err := readMigrationFile(file)
				if err != nil {
					return logger.LogNewError("failed to read migration file (%s): %w", file.displayPath, err)
				}
				_, err = tx.Exec(ctx, string(content))
				if err != nil {
					return logger.LogNewError("failed to execute migration file (%s): %w", file.displayPath, err)
				}
				logger.Infof("successfully applied migration file: %s", file.displayPath)
				latestVersion = file.name[:4]
			}
			if latestVersion != "" {
				if err := upsertMigrationRecord(ctx, tx, mgr.Schema, mgr.Service, latestVersion, mgrTime); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		logger.InfoContext(ctx, "migration completed", "service", mgr.Service)
	}
	return nil
}

// DoMigration performs all migrations for the registered configurations.
func (db *DBLink) DoMigration(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("migration failed: db is nil")
	}
	if db.cfg == nil {
		return fmt.Errorf("migration failed: db config is nil")
	}
	return DoMigration(ctx, db.Pool, db.cfg.Migrations)
}

func (db *DBLink) MustDoMigration(ctx context.Context) {
	err := db.DoMigration(ctx)
	if err != nil {
		katapp.Logger(ctx).Fatalf("migration failed: %v", err)
	}
}

func createMigrationTableIfNotExist(ctx context.Context, tx pgx.Tx, schema string) error {
	sql := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, schema)
	_, err := tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create schema for migration table: %v", err)
	}
	sql = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.%s(
	id SERIAL PRIMARY KEY,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	service TEXT NOT NULL UNIQUE,
	last_version TEXT NOT NULL
);`, schema, migrationTable)
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return katapp.Logger(ctx).LogNewError("failed to create migration table: %w", err)
	}
	return nil
}

func loadMigrationRecord(ctx context.Context, tx pgx.Tx, schema string, service string) (*migrationRecord, error) {
	rows, _ := tx.Query(ctx, fmt.Sprintf(`
SELECT service, updated_at, last_version
FROM %s.%s
WHERE service = $1
LIMIT 1
	`, schema, migrationTable), service)
	record, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[migrationRecord])
	if IsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, katapp.Logger(ctx).LogNewError("failed to find migration record by service=%s: %w", service, err)
	}
	return &record, nil
}

func upsertMigrationRecord(
	ctx context.Context,
	tx pgx.Tx,
	schema string,
	service string,
	version string,
	time time.Time,
) error {
	q := fmt.Sprintf(`
INSERT INTO %s.%s (service, updated_at, last_version)
VALUES ($1, $2, $3)
ON CONFLICT (service) DO UPDATE
SET updated_at = $2, last_version = $3
	`, schema, migrationTable)
	_, err := tx.Exec(ctx, q, service, time, version)
	if err != nil {
		return katapp.Logger(ctx).LogNewError("failed to upsert migration record for service=%s: %w", service, err)
	}
	return err
}

func listMigrationFiles(mgr katapp.DatabaseMigrationConfig) ([]migrationFile, error) {
	if mgr.FS == nil {
		files, err := internal.ListFilesInDirSortedByFilename(mgr.Path)
		if err != nil {
			return nil, err
		}
		result := make([]migrationFile, 0, len(files))
		for _, file := range files {
			_, name := filepath.Split(file)
			result = append(result, migrationFile{
				name:        name,
				displayPath: file,
				readPath:    file,
			})
		}
		return result, nil
	}

	entries, err := fs.ReadDir(mgr.FS, mgr.Path)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	result := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		readPath := path.Join(mgr.Path, entry.Name())
		result = append(result, migrationFile{
			name:        entry.Name(),
			displayPath: readPath,
			readPath:    readPath,
			sourceFS:    mgr.FS,
		})
	}
	return result, nil
}

func readMigrationFile(file migrationFile) ([]byte, error) {
	if file.sourceFS != nil {
		return fs.ReadFile(file.sourceFS, file.readPath)
	}
	return os.ReadFile(file.readPath)
}
