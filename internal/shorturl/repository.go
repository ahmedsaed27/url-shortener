package shorturl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const queryTimeout = 5 * time.Second

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, code string, originalURL string, urlHash string) (ShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const query = `
		INSERT INTO short_urls (code, original_url, url_hash)
		VALUES ($1, $2, $3)
		RETURNING id, code, original_url, url_hash, user_id, expires_at, is_active, created_at, updated_at
	`

	var shortURL ShortURL

	err := r.db.QueryRow(ctx, query, code, originalURL, urlHash).Scan(
		&shortURL.ID,
		&shortURL.Code,
		&shortURL.OriginalURL,
		&shortURL.URLHash,
		&shortURL.UserID,
		&shortURL.ExpiresAt,
		&shortURL.IsActive,
		&shortURL.CreatedAt,
		&shortURL.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ShortURL{}, ErrDuplicateCode
		}

		return ShortURL{}, fmt.Errorf("create short url: %w", err)
	}

	return shortURL, nil
}

func (r *Repository) FindByCode(ctx context.Context, code string) (ShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const query = `
		SELECT id, code, original_url, url_hash, user_id, expires_at, is_active, created_at, updated_at
		FROM short_urls
		WHERE code = $1
		LIMIT 1
	`

	var shortURL ShortURL

	err := r.db.QueryRow(ctx, query, code).Scan(
		&shortURL.ID,
		&shortURL.Code,
		&shortURL.OriginalURL,
		&shortURL.URLHash,
		&shortURL.UserID,
		&shortURL.ExpiresAt,
		&shortURL.IsActive,
		&shortURL.CreatedAt,
		&shortURL.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ShortURL{}, ErrShortURLNotFound
		}

		return ShortURL{}, fmt.Errorf("find short url by code: %w", err)
	}

	return shortURL, nil
}
