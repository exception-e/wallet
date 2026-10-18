package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const FileName = "config.env"

type Config struct {
	HTTP     HTTP
	Database Database
}

type HTTP struct {
	Address           string        `env:"HTTP_ADDRESS"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT"`
	ShutdownTimeout   time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT"`
}

type Database struct {
	DSN              string        `env:"POSTGRES_DSN"`
	MaxConns         int32         `env:"POSTGRES_MAX_CONNS"`
	MinConns         int32         `env:"POSTGRES_MIN_CONNS,notEmpty"`
	ConnectTimeout   time.Duration `env:"POSTGRES_CONNECT_TIMEOUT"`
	MaxConnLifetime  time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME"`
	MaxConnIdleTime  time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME"`
	MigrationTimeout time.Duration `env:"POSTGRES_MIGRATION_TIMEOUT"`
}

func Load(path string) (Config, error) {
	values, err := godotenv.Read(path)
	if err != nil {
		return Config{}, fmt.Errorf("config.Load (read path=%q): %w", path, err)
	}

	var cfg Config
	err = env.ParseWithOptions(&cfg, env.Options{
		Environment:     values,
		RequiredIfNoDef: true,
	})
	if err != nil {
		return Config{}, fmt.Errorf("config.Load (parse): %w", err)
	}
	return cfg, nil
}
