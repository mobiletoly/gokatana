package katsqlite

import (
	"context"
	"database/sql"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kattest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDBLink_DoMigration_MultipleServices(t *testing.T) {
	ctx := kattest.AppTestContext()

	service1MigrationsStep1 := "internal/dbmgr_test/migrations_service1_step1"
	service1MigrationsStep2 := "internal/dbmgr_test/migrations_service1_step2"
	service2MigrationsStep1 := "internal/dbmgr_test/migrations_service2_step1"
	service2MigrationsStep2 := "internal/dbmgr_test/migrations_service2_step2"

	tempDBFile := "test_migration.db"
	t.Cleanup(func() {
		os.Remove(tempDBFile)
	})

	dbLink, err := Connect(ctx, &katapp.DatabaseConfig{File: tempDBFile})
	require.NoError(t, err)
	t.Cleanup(func() {
		dbLink.Close()
	})

	runMigration := func(service string, migrationPath string, validateFunc func(ctx context.Context, db *sql.DB, t *testing.T)) {
		dbLink.cfg.Migrations = []katapp.DatabaseMigrationConfig{
			{
				Service: service,
				Path:    migrationPath,
			},
		}

		err := dbLink.DoMigration(ctx)
		require.NoError(t, err)
		validateFunc(ctx, dbLink.DB, t)
	}

	runMigration("test_service_1", service1MigrationsStep1, func(ctx context.Context, db *sql.DB, t *testing.T) {
		// Verify 'contact' table creation
		var tableExists int
		err := db.QueryRowContext(ctx, `
            SELECT COUNT(*) 
            FROM sqlite_master 
            WHERE type='table' AND name='contact'
        `).Scan(&tableExists)
		require.NoError(t, err)
		assert.Equal(t, 1, tableExists)

		// Verify that 'email' column does NOT exist
		var emailColumnExists int
		err = db.QueryRowContext(ctx, `
            SELECT COUNT(*) 
            FROM pragma_table_info('contact') 
            WHERE name='email'
        `).Scan(&emailColumnExists)
		require.NoError(t, err)
		assert.Equal(t, 0, emailColumnExists, "'email' column should not exist after the first migration")
	})

	runMigration("test_service_2", service2MigrationsStep1, func(ctx context.Context, db *sql.DB, t *testing.T) {
		// Verify 'orders' table creation
		var tableExists int
		err := db.QueryRowContext(ctx, `
            SELECT COUNT(*) 
            FROM sqlite_master 
            WHERE type='table' AND name='orders'
        `).Scan(&tableExists)
		require.NoError(t, err)
		assert.Equal(t, 1, tableExists)

		// Verify that 'order_date' column does NOT exist
		var orderDateColumnExists int
		err = db.QueryRowContext(ctx, `
            SELECT COUNT(*) 
            FROM pragma_table_info('orders') 
            WHERE name='order_date'
        `).Scan(&orderDateColumnExists)
		require.NoError(t, err)
		assert.Equal(t, 0, orderDateColumnExists, "'order_date' column should not exist after the first migration")
	})

	runMigration("test_service_1", service1MigrationsStep2, func(ctx context.Context, db *sql.DB, t *testing.T) {
		// Verify that 'email' column exists
		var emailColumnExists int
		err := db.QueryRowContext(ctx, `
            SELECT COUNT(*) 
            FROM pragma_table_info('contact') 
            WHERE name='email'
        `).Scan(&emailColumnExists)
		require.NoError(t, err)
		assert.Equal(t, 1, emailColumnExists, "'email' column should exist after the second migration")
	})

	runMigration("test_service_1", service1MigrationsStep2, func(ctx context.Context, db *sql.DB, t *testing.T) {
		// Reapply migration to ensure idempotence
	})

	runMigration("test_service_2", service2MigrationsStep2, func(ctx context.Context, db *sql.DB, t *testing.T) {
		// Verify that 'order_date' column exists
		var orderDateColumnExists int
		err := db.QueryRowContext(ctx, `
            SELECT COUNT(*) 
            FROM pragma_table_info('orders') 
            WHERE name='order_date'
        `).Scan(&orderDateColumnExists)
		require.NoError(t, err)
		assert.Equal(t, 1, orderDateColumnExists, "'order_date' column should exist after the second migration")
	})

	runMigration("test_service_2", service2MigrationsStep2, func(ctx context.Context, db *sql.DB, t *testing.T) {
		// Reapply migration to ensure idempotence
	})
}
