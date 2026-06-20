package store

import (
	"github.com/as9840935/url-shortener/internal/shorturl"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	ShortURLs *shorturl.Repository
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		ShortURLs: shorturl.NewRepository(db),
	}
}
