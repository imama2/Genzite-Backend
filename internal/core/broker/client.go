package broker

import (
	"context"
	"log/slog"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/redis/go-redis/v9"
)

// Client implements the BrokerService interface using Redis.
type Client struct {
	rdb    *redis.Client
	logger *slog.Logger
}

// NewClient creates a new Redis client and returns it as a BrokerService.
func NewClient(cfg *config.Config, logger *slog.Logger) (BrokerService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// Verify connection
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Client{
		rdb:    rdb,
		logger: logger,
	}, nil
}

// EnqueueBuildJob enqueues a new website build job into the 'build_queue'.
func (c *Client) EnqueueBuildJob(ctx context.Context, jobData []byte) error {
	return c.rdb.LPush(ctx, "build_queue", jobData).Err()
}

// SubscribeToProgress subscribes to a specific channel for progress updates.
func (c *Client) SubscribeToProgress(ctx context.Context, channelName string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channelName)
}
