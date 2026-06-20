package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/as9840935/url-shortener/internal/analytics"
	"github.com/as9840935/url-shortener/internal/cache"
	"github.com/as9840935/url-shortener/internal/config"
	"github.com/as9840935/url-shortener/internal/database"
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

	worker := analytics.NewWorker(redisClient, analytics.NewRepository(dbPool), cfg.ClickStreamName, cfg.ClickStreamGroup, cfg.ClickStreamConsumer, cfg.ClickStreamBatchSize, cfg.ClickStreamBlockTime)
	return worker.Run(ctx)
}
