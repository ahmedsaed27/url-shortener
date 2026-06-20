package health

import (
	"context"
	"net/http"
	"time"

	"github.com/as9840935/url-shortener/internal/response"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewHandler(db *pgxpool.Pool, redisClient *redis.Client) *Handler {
	return &Handler{
		db:    db,
		redis: redisClient,
	}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	statusCode := http.StatusOK
	result := map[string]string{
		"status":   "ok",
		"postgres": "ok",
		"redis":    "ok",
	}

	if err := h.db.Ping(ctx); err != nil {
		statusCode = http.StatusServiceUnavailable
		result["status"] = "error"
		result["postgres"] = "down"
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		statusCode = http.StatusServiceUnavailable
		result["status"] = "error"
		result["redis"] = "down"
	}

	response.JSON(w, statusCode, result)
}
