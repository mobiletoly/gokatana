package katpg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/internal"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

type PostgresTestContainer struct {
	Name      string
	User      string
	Password  string
	Host      string
	Port      int
	Container *postgres.PostgresContainer
}

// RunPostgresTestContainer starts a postgres Container for testing and register
// a cleanup function to terminate the Container after the test is complete.
func RunPostgresTestContainer(ctx context.Context, t *testing.T, migrateDir *string, scripts []string) *PostgresTestContainer {

	defaultLogger := log.New(os.Stderr, "", log.LstdFlags)

	pc := PostgresTestContainer{
		Name:     "testdb",
		User:     "postgres",
		Password: "postgres",
	}

	customizer := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(pc.Name),
		postgres.WithUsername(pc.User),
		postgres.WithPassword(pc.Password),
		testcontainers.WithLogger(defaultLogger),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5 * time.Second)),
	}

	if migrateDir != nil {
		files, err := internal.ListFilesInDirSortedByFilename(*migrateDir)
		if err != nil {
			t.Fatalf("failed to list files in directory %s: %v", *migrateDir, err)
		}
		if len(files) > 0 {
			customizer = append(customizer, postgres.WithInitScripts(files...))
		}
	}
	if len(scripts) > 0 {
		customizer = append(customizer, postgres.WithInitScripts(scripts...))
	}

	var err error
	pc.Container, err = postgres.Run(ctx, "docker.io/postgres:16-alpine", customizer...)
	if err != nil {
		t.Fatal(err)
	}

	pc.Host, err = pc.Container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := pc.Container.MappedPort(context.Background(), "5432")
	if err != nil {
		t.Fatal(err)
	}
	pc.Port, _ = strconv.Atoi(port.Port())

	t.Logf("PostgresTestContainer started: %+v", pc)

	return &pc
}

func (pc *PostgresTestContainer) Terminate(ctx context.Context, t *testing.T) {
	if err := pc.Container.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate Container: %s", err)
	}
}

func (pc *PostgresTestContainer) ApplyToConfig(cfg *katapp.DatabaseConfig) {
	cfg.Name = pc.Name
	cfg.User = pc.User
	cfg.Password = pc.Password
	cfg.Host = pc.Host
	cfg.Port = pc.Port
}

func (pc *PostgresTestContainer) BuildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", pc.User, pc.Password, pc.Host, pc.Port, pc.Name)
}

func (pc *PostgresTestContainer) BuildPgxPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	poolCfg, err := pgxpool.ParseConfig(pc.BuildDSN())
	require.NoError(t, err, "failed to parse pool config")
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	require.NoError(t, err, "failed to create pgx pool")
	return pool
}
