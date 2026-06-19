package transport

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/docvault/reminder/internal/config"
)

type RabbitMQConfig struct {
	URL      string
	Queue    string
	DLQ      string
	Retry    string
	Prefetch int
}

type RabbitMQConnection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  RabbitMQConfig
}

func NewRabbitMQConnection(cfg RabbitMQConfig) (*RabbitMQConnection, error) {
	if cfg.Retry == "" {
		cfg.Retry = cfg.Queue + ".retry"
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	queues := []string{cfg.Queue, cfg.DLQ, cfg.Retry}
	for _, queue := range queues {
		var args amqp.Table
		switch queue {
		case cfg.Queue:
			// Main queue dead-letters to the DLQ on permanent failure.
			args = amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": cfg.DLQ,
			}
		case cfg.Retry:
			// Delayed-retry queue: it has no consumer. Messages published here
			// carry a per-message TTL and are dead-lettered back to the main
			// queue when the TTL expires, giving non-blocking exponential
			// backoff without ever stalling the consumer goroutine.
			args = amqp.Table{
				"x-dead-letter-exchange":    "",
				"x-dead-letter-routing-key": cfg.Queue,
			}
		}

		_, err := ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			args,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", queue, err)
		}
	}

	return &RabbitMQConnection{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

func (c *RabbitMQConnection) Channel() *amqp.Channel {
	return c.channel
}

func (c *RabbitMQConnection) Config() RabbitMQConfig {
	return c.config
}

func (c *RabbitMQConnection) Close() error {
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

func NewRabbitMQConnectionFromConfig(cfg *config.Config) *RabbitMQConfig {
	return &RabbitMQConfig{
		URL:      cfg.RabbitMQURL,
		Queue:    cfg.ReminderQueue,
		DLQ:      cfg.DeadLetterQueue,
		Retry:    cfg.ReminderQueue + ".retry",
		Prefetch: cfg.PrefetchCount,
	}
}
