package katpg

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kattest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDBLink_DoMigration_MultipleServices_SeparateSchemas(t *testing.T) {
	ctx := kattest.AppTestContext()

	// Configure migration directories for each service
	service1MigrationsStep1 := "internal/dbmgr_test/migrations_service1_step1"
	service1MigrationsStep2 := "internal/dbmgr_test/migrations_service1_step2"
	service2MigrationsStep1 := "internal/dbmgr_test/migrations_service2_step1"
	service2MigrationsStep2 := "internal/dbmgr_test/migrations_service2_step2"

	pc := RunPostgresTestContainer(ctx, t, nil, []string{})
	t.Cleanup(func() {
		pc.Terminate(ctx, t)
	})
	dbPool := pc.BuildPgxPool(ctx, t)

	runMigration := func(service string, migrationPath string, schema string, validateFunc func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T)) {
		dbLink := &DBLink{
			Pool: dbPool,
			cfg: &katapp.DatabaseConfig{
				Host:     pc.Host,
				Port:     pc.Port,
				Name:     pc.Name,
				User:     pc.User,
				Password: pc.Password,
				Sslmode:  "disable",
				Migrations: []katapp.DatabaseMigrationConfig{
					{
						Service: service,
						Schema:  schema,
						Path:    migrationPath,
					},
				},
			},
		}

		err := dbLink.DoMigration(ctx)
		require.NoError(t, err)
		validateFunc(ctx, dbPool, t)
	}

	runMigration("test_service_1", service1MigrationsStep1, "service1_sample", func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T) {
		// Verify table creation for service1
		var tableExists bool
		err := dbPool.QueryRow(ctx, `
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = 'service1_sample' AND table_name = 'contact'
            )`).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists)

		// Verify that 'email' column does NOT exist for service1
		var emailColumnExists bool
		err = dbPool.QueryRow(ctx, `
            SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_schema = 'service1_sample' AND table_name = 'contact' AND column_name = 'email'
            )`).Scan(&emailColumnExists)
		require.NoError(t, err)
		assert.False(t, emailColumnExists, "'email' column should not exist for service1 after the first migration")
	})

	runMigration("test_service_2", service2MigrationsStep1, "service2_sample", func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T) {
		// Verify table creation for service2
		var tableExists bool
		err := dbPool.QueryRow(ctx, `
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_schema = 'service2_sample' AND table_name = 'orders'
            )`).Scan(&tableExists)
		require.NoError(t, err)
		assert.True(t, tableExists)

		// Verify that 'order_date' column does NOT exist for service2
		var orderDateColumnExists bool
		err = dbPool.QueryRow(ctx, `
            SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_schema = 'service2_sample' AND table_name = 'orders' AND column_name = 'order_date'
            )`).Scan(&orderDateColumnExists)
		require.NoError(t, err)
		assert.False(t, orderDateColumnExists, "'order_date' column should not exist for service2 after the first migration")
	})

	runMigration("test_service_1", service1MigrationsStep2, "service1_sample", func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T) {
		// Verify that 'email' column exists for service1
		var emailColumnExists bool
		err := dbPool.QueryRow(ctx, `
            SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_schema = 'service1_sample' AND table_name = 'contact' AND column_name = 'email'
            )`).Scan(&emailColumnExists)
		require.NoError(t, err)
		assert.True(t, emailColumnExists, "'email' column should exist for service1 after the second migration")
	})

	runMigration("test_service_1", service1MigrationsStep2, "service1_sample", func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T) {
		// Attempt to apply the same migration again should not cause any error
	})

	runMigration("test_service_2", service2MigrationsStep2, "service2_sample", func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T) {
		// Verify that 'order_date' column exists for service2
		var orderDateColumnExists bool
		err := dbPool.QueryRow(ctx, `
            SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_schema = 'service2_sample' AND table_name = 'orders' AND column_name = 'order_date'
            )`).Scan(&orderDateColumnExists)
		require.NoError(t, err)
		assert.True(t, orderDateColumnExists, "'order_date' column should exist for service2 after the second migration")
	})

	runMigration("test_service_2", service2MigrationsStep2, "service2_sample", func(ctx context.Context, dbPool *pgxpool.Pool, t *testing.T) {
		// Attempt to apply the same migration again should not cause any error
	})
}
