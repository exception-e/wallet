package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"walletApp/internal/api/rest"
	"walletApp/internal/config"
	"walletApp/internal/logger"
	"walletApp/internal/service"
	"walletApp/internal/storage"
	"walletApp/internal/storage/psql"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	logHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)
	log := logger.New(logHandler)

	if err := godotenv.Load(); err != nil {
		log.Info("No .env, use provided environment variables")
	}

	cfg, err := config.Load(config.FileName)
	if err != nil {
		log.Error("Failed to config.Load", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := psql.InitDB(ctx, cfg.Database)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing DB connection")
		pool.Close()
	}()

	if err := psql.RunMigrations(ctx, cfg.Database); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	walletStorage := storage.New(pool)
	service := service.New(walletStorage)
	router := chi.NewRouter()
	rest.New(service).SetupRouter(router)

	srv := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		log.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("server failed", "error", err)
				os.Exit(1)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	}
	log.Info("server stopped gracefully")
}
