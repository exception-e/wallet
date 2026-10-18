package container

import (
	"context"
	"fmt"
	"testing"
	"time"
	"walletApp/internal/config"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const PostgresImage = "postgres:18.4-alpine"

// StartPostgres returns connection settings for a fresh PostgreSQL instance.
// The container is removed by tb.Cleanup, including when setup fails.
func StartPostgres(tb testing.TB, maxConns int32) config.Database {
	tb.Helper()
	ctx, cancel := context.WithTimeout(tb.Context(), 2*time.Minute)
	defer cancel()

	// Leave spare connections for migrations and diagnostics.
	maxConnections := int64(maxConns) + 10
	instance, err := postgres.Run(ctx, PostgresImage,
		postgres.WithDatabase("wallet"),
		postgres.WithUsername("wallet"),
		postgres.WithPassword("wallet"),
		// The postgres module disables fsync by default. Keep durable commits here.
		testcontainers.WithCmd("postgres",
			"-c", "fsync=on",
			"-c", "synchronous_commit=on",
			"-c", "full_page_writes=on",
			"-c", fmt.Sprintf("max_connections=%d", maxConnections),
		),
		postgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(tb, instance)
	if err != nil {
		tb.Fatal(fmt.Errorf("container.StartPostgres (start container): %w", err))
	}

	dsn, err := instance.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tb.Fatal(fmt.Errorf("container.StartPostgres (connection string): %w", err))
	}

	return config.Database{
		DSN:              dsn,
		MaxConns:         maxConns,
		MinConns:         maxConns,
		ConnectTimeout:   10 * time.Second,
		MaxConnLifetime:  time.Hour,
		MaxConnIdleTime:  time.Minute,
		MigrationTimeout: time.Minute,
	}
}
