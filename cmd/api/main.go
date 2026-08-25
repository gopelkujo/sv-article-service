// Package main is the entrypoint for the article REST API service.
// It loads configuration, wires dependencies, and starts the HTTP server.
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

	"github.com/gopelkujo/sv-article-service/internal/config"
	"github.com/gopelkujo/sv-article-service/internal/handler"
	"github.com/gopelkujo/sv-article-service/internal/repository/mysql"
	"github.com/gopelkujo/sv-article-service/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(logger); err != nil {
		logger.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := mysql.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	repo := mysql.NewArticleRepository(db)
	articleService := service.NewArticleService(repo)
	articleHandler := handler.NewArticleHandler(articleService, logger)
	healthHandler := handler.NewHealthHandler(db)
	router := handler.NewRouter(articleHandler, healthHandler, logger, cfg.CORSAllowedOrigins)

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting",
			"addr", cfg.Addr(),
			"env", cfg.AppEnv,
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("http server stopped gracefully")

	// Drain ListenAndServe result if shutdown was signal-driven.
	select {
	case <-errCh:
	case <-time.After(time.Second):
	}

	return nil
}
