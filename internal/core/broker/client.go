package broker

import (
	"context"
	"log/slog"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/redis/go-redis/v9"
)

const (
	QueueBuild        = "queue:build"
	QueueNotification = "queue:notification"
)

// BuildJob defines the structure for a build job payload.
type BuildJob struct {
	BuildID       string `json:"build_id"`
	UserID        string `json:"user_id"`
	BlueprintJSON string `json:"blueprint_json"`
}

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

// Enqueue adds a job to the specified queue.
func (c *Client) Enqueue(ctx context.Context, queueName string, jobData []byte) error {
	return c.rdb.LPush(ctx, queueName, jobData).Err()
}

// Dequeue removes and returns a job from the specified queue.
// This is a blocking operation.
func (c *Client) Dequeue(ctx context.Context, queueName string) ([]byte, error) {
	// BRPop will block until a value is available or the context is cancelled.
	// A timeout of 0 means it will block indefinitely.
	result, err := c.rdb.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		return nil, err
	}
	// result is a []string where result[0] is the queue name and result[1] is the value.
	return []byte(result[1]), nil
}

// EnqueueBuildJob enqueues a new website build job into the 'build_queue'.
func (c *Client) EnqueueBuildJob(ctx context.Context, jobData []byte) error {
	return c.rdb.LPush(ctx, "build_queue", jobData).Err()
}

// SubscribeToProgress subscribes to a specific channel for progress updates.
func (c *Client) SubscribeToProgress(ctx context.Context, channelName string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channelName)
}
