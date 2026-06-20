package shorturl

import "time"

type ShortURL struct {
	ID          int64
	Code        string
	OriginalURL string
	URLHash     *string
	UserID      *int64
	ExpiresAt   *time.Time
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
