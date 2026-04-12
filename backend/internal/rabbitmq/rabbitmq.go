// Package rabbitmq provides RabbitMQ connection initialization.
package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn  *amqp.Connection
	queue string
}

// NewConnection creates a new RabbitMQ connection.
func NewConnection(ctx context.Context, url string) (*amqp.Connection, error) {
	if url == "" {
		return nil, fmt.Errorf("RabbitMQ URL is required")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Wait for connection to be ready
	for i := 0; i < 10; i++ {
		if conn.IsClosed() {
			conn.Close()
			return nil, fmt.Errorf("connection closed during initialization")
		}
		time.Sleep(100 * time.Millisecond)
	}

	slog.Info("RabbitMQ connection established")

	return conn, nil
}

// NewChannel creates a new RabbitMQ channel.
func NewChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	slog.Info("RabbitMQ channel created")

	return ch, nil
}

func NewPublisher(conn *amqp.Connection, queue string) *Publisher {
	return &Publisher{conn: conn, queue: queue}
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(
		p.queue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", p.queue),
		},
	); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := ch.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
