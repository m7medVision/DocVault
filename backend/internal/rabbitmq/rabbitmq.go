// Package rabbitmq provides RabbitMQ connection initialization.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	documentapp "github.com/docvault/backend/internal/document/app"
	amqp "github.com/rabbitmq/amqp091-go"
)

var amqpDial = amqp.Dial

type Publisher struct {
	conn  *amqp.Connection
	url   string
	queue string
	mu    sync.Mutex
	ch    *amqp.Channel
}

type OCRDispatcher struct {
	publisher *Publisher
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

func NewOCRDispatcher(conn *amqp.Connection, url string, queue string) *OCRDispatcher {
	return &OCRDispatcher{publisher: NewPublisher(conn, url, queue)}
}

func (d *OCRDispatcher) DispatchOCR(ctx context.Context, job documentapp.OCRJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal OCR job: %w", err)
	}

	return d.publisher.Publish(ctx, body)
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Reuse a single long-lived channel and declare the queue topology only
	// once per channel lifetime. Channels are not goroutine-safe, but the mutex
	// above serializes every publish, so one cached channel is sufficient and
	// avoids opening/closing a channel plus redeclaring both queues on every
	// message.
	if err := p.ensureChannel(); err != nil {
		return err
	}

	if err := p.ch.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
		// The channel may be in a bad state after a publish error; drop it so
		// the next call reopens and re-declares the topology on a fresh channel.
		p.resetChannel()
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// ensureChannel lazily opens the cached channel and declares the queue topology
// once. Callers must hold p.mu.
func (p *Publisher) ensureChannel() error {
	if p.ch != nil && !p.ch.IsClosed() {
		return nil
	}
	p.resetChannel()

	ch, err := p.openChannel()
	if err != nil {
		return err
	}
	if err := declareTopology(ch, p.queue); err != nil {
		_ = ch.Close()
		return err
	}
	p.ch = ch
	return nil
}

// resetChannel closes and clears the cached channel. Callers must hold p.mu.
func (p *Publisher) resetChannel() {
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
}

// declareTopology declares the DLQ and the main queue (wired to the DLQ). It is
// idempotent and only needs to run once per channel.
func declareTopology(ch *amqp.Channel, queue string) error {
	dlqName := fmt.Sprintf("%s.dlq", queue)
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

	if _, err := ch.QueueDeclare(
		queue,
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
