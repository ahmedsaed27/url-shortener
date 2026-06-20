package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/as9840935/url-shortener/internal/analytics"
	"github.com/as9840935/url-shortener/internal/cache"
	"github.com/as9840935/url-shortener/internal/config"
	"github.com/as9840935/url-shortener/internal/database"
	"github.com/as9840935/url-shortener/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	metrics.Register()
	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer dbPool.Close()
	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer redisClient.Close()

	metricsServer := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.WorkerMetricsPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/metrics" {
				http.NotFound(w, r)
				return
			}
			promhttp.Handler().ServeHTTP(w, r)
		}),
	}
	metricsServerErr := make(chan error, 1)
	go func() {
		metricsServerErr <- metricsServer.ListenAndServe()
	}()

	worker := analytics.NewWorker(redisClient, analytics.NewRepository(dbPool), cfg.ClickStreamName, cfg.ClickStreamGroup, cfg.ClickStreamConsumer, cfg.ClickStreamBatchSize, cfg.ClickStreamBlockTime)
	workerErr := make(chan error, 1)
	go func() {
		workerErr <- worker.Run(ctx)
	}()

	select {
	case err := <-workerErr:
		if err != nil {
			return err
		}
	case err := <-metricsServerErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("worker metrics server stopped: %w", err)
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown worker metrics server: %w", err)
	}

	return nil
}
