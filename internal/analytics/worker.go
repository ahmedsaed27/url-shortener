package analytics

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClickRepositoryContract interface {
	InsertClicks(ctx context.Context, events []ClickEvent) error
}

type Worker struct {
	client       *redis.Client
	repo         ClickRepositoryContract
	streamName   string
	groupName    string
	consumerName string
	batchSize    int64
	blockTime    time.Duration
}

func NewWorker(client *redis.Client, repo ClickRepositoryContract, streamName, groupName, consumerName string, batchSize int, blockTime time.Duration) *Worker {
	return &Worker{
		client:       client,
		repo:         repo,
		streamName:   streamName,
		groupName:    groupName,
		consumerName: consumerName,
		batchSize:    int64(batchSize),
		blockTime:    blockTime,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.ensureConsumerGroup(ctx); err != nil {
		return err
	}

	for {
		if ctx.Err() != nil {
			return nil
		}

		streams, err := w.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    w.groupName,
			Consumer: w.consumerName,
			Streams:  []string{w.streamName, ">"},
			Count:    w.batchSize,
			Block:    w.blockTime,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if errors.Is(err, redis.Nil) {
				continue
			}
			return fmt.Errorf("read click events: %w", err)
		}

		events := make([]ClickEvent, 0, w.batchSize)
		messageIDs := make([]string, 0, w.batchSize)
		invalidMessageIDs := make([]string, 0)

		for _, stream := range streams {
			for _, message := range stream.Messages {
				event, err := clickEventFromMessage(message)
				if err != nil {
					log.Printf("decode click event %s: %v", message.ID, err)
					invalidMessageIDs = append(invalidMessageIDs, message.ID)
					continue
				}

				events = append(events, event)
				messageIDs = append(messageIDs, message.ID)
			}
		}

		if len(invalidMessageIDs) > 0 {
			if err := w.client.XAck(ctx, w.streamName, w.groupName, invalidMessageIDs...).Err(); err != nil {
				log.Printf("acknowledge invalid click events: %v", err)
			}
		}

		if len(events) == 0 {
			continue
		}

		if err := w.repo.InsertClicks(ctx, events); err != nil {
			log.Printf("insert %d click events: %v", len(events), err)
			continue
		}

		if err := w.client.XAck(ctx, w.streamName, w.groupName, messageIDs...).Err(); err != nil {
			log.Printf("acknowledge %d click events: %v", len(messageIDs), err)
		}
	}
}

func (w *Worker) ensureConsumerGroup(ctx context.Context) error {
	err := w.client.XGroupCreateMkStream(ctx, w.streamName, w.groupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create click consumer group: %w", err)
	}

	return nil
}

func clickEventFromMessage(message redis.XMessage) (ClickEvent, error) {
	code, err := messageValue(message, "code")
	if err != nil {
		return ClickEvent{}, err
	}
	clickedAtValue, err := messageValue(message, "clicked_at")
	if err != nil {
		return ClickEvent{}, err
	}
	clickedAt, err := time.Parse(time.RFC3339Nano, clickedAtValue)
	if err != nil {
		return ClickEvent{}, fmt.Errorf("parse clicked_at: %w", err)
	}
	ip, err := messageValue(message, "ip")
	if err != nil {
		return ClickEvent{}, err
	}
	userAgent, err := messageValue(message, "user_agent")
	if err != nil {
		return ClickEvent{}, err
	}
	referer, err := messageValue(message, "referer")
	if err != nil {
		return ClickEvent{}, err
	}

	return ClickEvent{Code: code, IP: ip, UserAgent: userAgent, Referer: referer, ClickedAt: clickedAt}, nil
}

func messageValue(message redis.XMessage, key string) (string, error) {
	value, ok := message.Values[key]
	if !ok {
		return "", fmt.Errorf("missing %s", key)
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}

	return stringValue, nil
}
