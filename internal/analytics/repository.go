package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const queryTimeout = 5 * time.Second

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertClick(ctx context.Context, event ClickEvent) error {
	return r.InsertClicks(ctx, []ClickEvent{event})
}

func (r *Repository) InsertClicks(ctx context.Context, events []ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var query strings.Builder
	query.WriteString("INSERT INTO url_clicks (code, ip_address, user_agent, referer, clicked_at) VALUES ")

	args := make([]any, 0, len(events)*5)
	placeholders := make([]string, 0, len(events))

	for index, event := range events {
		placeholderOffset := index*5 + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", placeholderOffset, placeholderOffset+1, placeholderOffset+2, placeholderOffset+3, placeholderOffset+4))
		args = append(args, event.Code, event.IP, event.UserAgent, event.Referer, event.ClickedAt)
	}

	query.WriteString(strings.Join(placeholders, ", "))

	if _, err := r.db.Exec(ctx, query.String(), args...); err != nil {
		return fmt.Errorf("insert url clicks: %w", err)
	}

	return nil
}
