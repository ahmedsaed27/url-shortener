package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                   string `validate:"required"`
	HTTPPort                 int    `validate:"required,min=1"`
	AppBaseURL               string `validate:"required,url"`
	ShortCodeLength          int    `validate:"required,min=4,max=20"`
	DatabaseURL              string `validate:"required"`
	RedisAddr                string `validate:"required"`
	RedisPassword            string
	RedisDB                  int           `validate:"min=0"`
	ShortURLCacheTTL         time.Duration `validate:"gt=0"`
	ShortURLNegativeCacheTTL time.Duration `validate:"gt=0"`
	ClickStreamName          string        `validate:"required"`
	ClickStreamGroup         string        `validate:"required"`
	ClickStreamConsumer      string        `validate:"required"`
	ClickStreamBatchSize     int           `validate:"min=1"`
	ClickStreamBlockTime     time.Duration `validate:"gt=0"`
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	httpPort, err := getEnvAsInt("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	redisDB, err := getEnvAsInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	shortCodeLength, err := getEnvAsInt("SHORT_CODE_LENGTH", 7)
	if err != nil {
		return Config{}, err
	}

	shortURLCacheTTL, err := getEnvAsDuration("SHORT_URL_CACHE_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	shortURLNegativeCacheTTL, err := getEnvAsDuration("SHORT_URL_NEGATIVE_CACHE_TTL", time.Minute)
	if err != nil {
		return Config{}, err
	}

	clickStreamBatchSize, err := getEnvAsInt("CLICK_STREAM_BATCH_SIZE", 50)
	if err != nil {
		return Config{}, err
	}

	clickStreamBlockTime, err := getEnvAsDuration("CLICK_STREAM_BLOCK_TIME", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:                   getEnv("APP_ENV", "local"),
		HTTPPort:                 httpPort,
		AppBaseURL:               getEnv("APP_BASE_URL", "http://localhost:8080"),
		ShortCodeLength:          shortCodeLength,
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		RedisAddr:                getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:            getEnv("REDIS_PASSWORD", ""),
		RedisDB:                  redisDB,
		ShortURLCacheTTL:         shortURLCacheTTL,
		ShortURLNegativeCacheTTL: shortURLNegativeCacheTTL,
		ClickStreamName:          getEnv("CLICK_STREAM_NAME", "url_clicks"),
		ClickStreamGroup:         getEnv("CLICK_STREAM_GROUP", "url_click_workers"),
		ClickStreamConsumer:      getEnv("CLICK_STREAM_CONSUMER", "worker-1"),
		ClickStreamBatchSize:     clickStreamBatchSize,
		ClickStreamBlockTime:     clickStreamBlockTime,
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return parsedValue, nil
}

func getEnvAsDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	parsedValue, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return parsedValue, nil
}
