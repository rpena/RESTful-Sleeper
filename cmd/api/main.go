package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rpena/RESTful-Sleeper/internal/cache"
	"github.com/rpena/RESTful-Sleeper/internal/config"
	"github.com/rpena/RESTful-Sleeper/internal/handler"
	"github.com/rpena/RESTful-Sleeper/internal/sleeper"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	redisCache := cache.New(cfg.RedisURL)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := redisCache.Ping(ctx); err != nil {
		logger.Error("redis unavailable", "error", err)
		os.Exit(1)
	}

	client := sleeper.NewClient(cfg.SleeperBaseURL, cfg.RequestTimeout)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.New(client, redisCache, cfg.CacheTTL),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
