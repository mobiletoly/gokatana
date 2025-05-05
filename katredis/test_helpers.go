package katredis

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

type RedisTestContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

func RunRedisTestContainer(ctx context.Context, t *testing.T) *RedisTestContainer {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7.0-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(10 * time.Second),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get Redis container host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("failed to get Redis container port: %v", err)
	}

	return &RedisTestContainer{
		Container: redisContainer,
		Host:      host,
		Port:      port.Port(),
	}
}

func (rc *RedisTestContainer) Address() string {
	return fmt.Sprintf("%s:%s", rc.Host, rc.Port)
}

func (rc *RedisTestContainer) Terminate(ctx context.Context, t *testing.T) {
	t.Helper()
	if err := rc.Container.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate Redis container: %v", err)
	}
}
