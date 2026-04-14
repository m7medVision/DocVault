// Package rabbitmq provides RabbitMQ connection initialization.
package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var amqpDial = amqp.Dial

type Publisher struct {
	conn  *amqp.Connection
	url   string
	queue string
	mu    sync.Mutex
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

func NewPublisher(conn *amqp.Connection, url string, queue string) *Publisher {
	return &Publisher{conn: conn, url: url, queue: queue}
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch, err := p.openChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Ensure DLQ exists first
	dlqName := fmt.Sprintf("%s.dlq", p.queue)
	if _, err := ch.QueueDeclare(
		dlqName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Declare main queue with DLQ configuration
	if _, err := ch.QueueDeclare(
		p.queue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlqName,
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

func (p *Publisher) openChannel() (*amqp.Channel, error) {
	if err := p.ensureConnection(); err != nil {
		return nil, err
	}

	ch, err := p.conn.Channel()
	if err == nil {
		return ch, nil
	}

	p.closeConnection()
	if err := p.ensureConnection(); err != nil {
		return nil, err
	}

	ch, err = p.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return ch, nil
}

func (p *Publisher) ensureConnection() error {
	if p.conn != nil && !p.conn.IsClosed() {
		return nil
	}
	if p.url == "" {
		return fmt.Errorf("rabbit connection is closed")
	}

	conn, err := amqpDial(p.url)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	p.conn = conn
	return nil
}

func (p *Publisher) closeConnection() {
	if p.conn == nil {
		return
	}
	_ = p.conn.Close()
	p.conn = nil
}
