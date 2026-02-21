package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/abramyaan/subscription-service/internal/config"
	"github.com/abramyaan/subscription-service/internal/handler"
	"github.com/abramyaan/subscription-service/internal/repository"
	"github.com/abramyaan/subscription-service/internal/service"
	"github.com/abramyaan/subscription-service/pkg/logger"
	"github.com/abramyaan/subscription-service/pkg/validator"
)

// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.New(cfg.Log.Level)
	appLogger.Info("Starting subscription service",
		slog.String("port", cfg.Server.Port),
		slog.String("log_level", cfg.Log.Level),
	)

	ctx := context.Background()
	pool, err := connectDB(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	appLogger.Info("Database connection established")

	if err := runMigrations(pool, appLogger); err != nil {
		appLogger.Error("Failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	v := validator.New()
	subscriptionRepo := repository.NewSubscriptionRepository(pool, appLogger)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, appLogger)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService, v, appLogger)

	router := handler.NewRouter(subscriptionHandler)

	srv := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		appLogger.Info("Server starting", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	appLogger.Info("Server stopped gracefully")
}
func connectDB(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
    logger.Debug("connecting to database", slog.String("dsn", "host="+cfg.Database.Host))

    poolConfig, err := pgxpool.ParseConfig(cfg.Database.GetDSN())
    if err != nil {
        return nil, fmt.Errorf("failed to parse database config: %w", err)
    }

    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5
    poolConfig.MaxConnLifetime = time.Hour
    poolConfig.MaxConnIdleTime = 30 * time.Minute

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    logger.Info("database connection established successfully")
    return pool, nil
}

func runMigrations(pool *pgxpool.Pool, logger *slog.Logger) error {
    logger.Info("Starting database migrations...")

    migrationSQL, err := os.ReadFile("migrations/001_init.sql")
    if err != nil {
        return fmt.Errorf("failed to read migration file: %w", err)
    }

    ctx := context.Background()
    if _, err := pool.Exec(ctx, string(migrationSQL)); err != nil {
        return fmt.Errorf("failed to execute migration: %w", err)
    }

    logger.Info("Migrations applied successfully")
    return nil
}
