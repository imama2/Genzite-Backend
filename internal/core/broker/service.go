package broker

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// BrokerService defines the interface for a message broker.
type BrokerService interface {
	Enqueue(ctx context.Context, queueName string, jobData []byte) error
	Dequeue(ctx context.Context, queueName string) ([]byte, error)
	SubscribeToProgress(ctx context.Context, channelName string) *redis.PubSub
}
