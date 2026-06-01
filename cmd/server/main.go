package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/donamo/todo-backend/db/migrations"
	"github.com/donamo/todo-backend/internal/app"
	"github.com/donamo/todo-backend/internal/config"
	"github.com/donamo/todo-backend/internal/logging"
)

func main() {
	logging.Init()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn("load .env failed, using environment variables", "err", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	db, err := openDB()
	if err != nil {
		slog.Error("init database", "err", err)
		os.Exit(1)
	}
	if db != nil && cfg.SessionSecret == "" {
		slog.Error("SESSION_SECRET is required when database-backed auth is enabled")
		closeDB(db)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.New(cfg, db),
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", "port", cfg.Port)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err == nil {
			closeDB(db)
			return
		}
		slog.Error("server stopped", "err", err)
		closeDB(db)
		os.Exit(1)
	case <-ctx.Done():
		slog.Info("shutdown requested")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	closeDB(db)
	if err := <-serverErr; err != nil {
		slog.Error("server stopped during shutdown", "err", err)
		os.Exit(1)
	}
	slog.Info("server shutdown complete")
}

func openDB() (*sql.DB, error) {
	dbURL := config.String("DATABASE_URL", "")
	if dbURL == "" {
		slog.Warn("DATABASE_URL is not set, domain GraphQL operations are disabled")
		return nil, nil
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := goose.Up(db, "."); err != nil {
		_ = db.Close()
		return nil, err
	}
	slog.Info("database ready")
	return db, nil
}

func closeDB(db *sql.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		slog.Error("database close failed", "err", err)
		return
	}
	slog.Info("database connection closed")
}
