package analytics

import (
	"context"
	"time"

	"github.com/as9840935/url-shortener/internal/metrics"
	"github.com/redis/go-redis/v9"
)

type Producer struct {
	client     *redis.Client
	streamName string
}

func NewProducer(client *redis.Client, streamName string) *Producer {
	return &Producer{
		client:     client,
		streamName: streamName,
	}
}

func (p *Producer) TrackClick(ctx context.Context, event ClickEvent) error {
	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"code":       event.Code,
			"ip":         event.IP,
			"user_agent": event.UserAgent,
			"referer":    event.Referer,
			"clicked_at": event.ClickedAt.Format(time.RFC3339Nano),
		},
	}).Err()
	if err != nil {
		metrics.AnalyticsProducerErrorsTotal.Inc()
		return err
	}

	metrics.AnalyticsEventsProducedTotal.Inc()
	return nil
}
