package broker

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// BrokerService defines the interface for a message broker.
type BrokerService interface {
	EnqueueBuildJob(ctx context.Context, jobData []byte) error
	SubscribeToProgress(ctx context.Context, channelName string) *redis.PubSub
}
