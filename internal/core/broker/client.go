package broker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/imama2/Genzite-Backend/internal/core/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(context.Context, []byte) error

type Client struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
	logger    *slog.Logger
}

func NewClient(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	if strings.TrimSpace(cfg.RabbitMQURL) == "" {
		return nil, errors.New("RABBITMQ_URL is required")
	}

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	queueName := strings.TrimSpace(cfg.RabbitMQQueue)
	if queueName == "" {
		queueName = "notifications.email"
	}

	if _, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := channel.Qos(5, 0, false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Client{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
		logger:    logger,
	}, nil
}

func (c *Client) Publish(ctx context.Context, body []byte) error {
	if c == nil || c.channel == nil {
		return errors.New("broker not initialized")
	}

	return c.channel.PublishWithContext(
		ctx,
		"",
		c.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

func (c *Client) Consume(ctx context.Context, handler HandlerFunc) error {
	if c == nil || c.channel == nil {
		return errors.New("broker not initialized")
	}
	if handler == nil {
		return errors.New("handler required")
	}

	deliveries, err := c.channel.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			if err := handler(ctx, delivery.Body); err != nil {
				if nackErr := delivery.Nack(false, false); nackErr != nil {
					c.logger.Error("failed to nack message", "error", nackErr)
				}
				continue
			}
			if err := delivery.Ack(false); err != nil {
				c.logger.Error("failed to ack message", "error", err)
			}
		}
	}
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
