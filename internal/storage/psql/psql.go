package psql

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"
	"walletApp/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed *.sql
var FS embed.FS

func InitDB(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres.New (parse connection config): %w", err)
	}

	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	//poolConfig.ConnConfig.RuntimeParams["application_name"] = "walletApp"
	//poolConfig.ConnConfig.RuntimeParams["default_transaction_isolation"] = "read committed"

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres.New (create pool): %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return pool, nil
}

func RunMigrations(ctx context.Context, cfg config.Database) (err error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.MigrationTimeout)
	defer cancel()

	connConfig, err := pgx.ParseConfig(cfg.DSN)
	if err != nil {
		return fmt.Errorf("migrations.Up (parse connection config): %w", err)
	}
	connConfig.ConnectTimeout = cfg.ConnectTimeout
	connConfig.RuntimeParams["application_name"] = "wallet-migrations"

	db := stdlib.OpenDB(*connConfig)
	db.SetMaxOpenConns(1)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("migrations.Up (close database): %w", closeErr))
		}
	}()

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("migrations.Up (create advisory locker): %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, FS, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("migrations.Up (create provider): %w", err)
	}

	if _, err = provider.Up(ctx); err != nil {
		return fmt.Errorf("migrations.Up (apply migrations): %w", err)
	}

	return nil
}
