package app

import (
	"context"
	"fmt"

	"github.com/as9840935/url-shortener/internal/analytics"
	"github.com/as9840935/url-shortener/internal/cache"
	"github.com/as9840935/url-shortener/internal/config"
	"github.com/as9840935/url-shortener/internal/database"
	"github.com/as9840935/url-shortener/internal/ratelimit"
	"github.com/as9840935/url-shortener/internal/store"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	Config            config.Config
	DB                *pgxpool.Pool
	Redis             *redis.Client
	Store             store.Storage
	AnalyticsProducer *analytics.Producer
	RateLimiter       ratelimit.Limiter
	Handlers          Handlers
	Validate          *validator.Validate
}

func New(ctx context.Context) (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	app := &Application{
		Config:            cfg,
		DB:                dbPool,
		Redis:             redisClient,
		Store:             store.NewStorage(dbPool),
		AnalyticsProducer: analytics.NewProducer(redisClient, cfg.ClickStreamName),
		RateLimiter:       ratelimit.NewRedisLimiter(redisClient),
		Validate:          validator.New(validator.WithRequiredStructEnabled()),
	}

	app.Handlers = RegisterHandlers(app)

	return app, nil
}

func (app *Application) Close() {
	if app.Redis != nil {
		_ = app.Redis.Close()
	}

	if app.DB != nil {
		app.DB.Close()
	}
}
