// Package queue provides RabbitMQ consumer for reminder jobs.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/docvault/reminder/internal/config"
)

// Message represents a message from the queue.
type Message struct {
	JobID        string            `json:"job_id"`
	DocumentID   string            `json:"document_id"`
	TenantID     string            `json:"tenant_id"`
	OrgID        string            `json:"org_id"`
	SourceText   string            `json:"source_text"`
	DocumentType *string           `json:"document_type,omitempty"`
	ExpiryDate   *string           `json:"expiry_date,omitempty"`
	Issuer       *string           `json:"issuer,omitempty"`
	Priority     string            `json:"priority"`
	CreatedAt    string            `json:"created_at"`
	RetryCount   int               `json:"retry_count"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Handler is a function that processes a message.
type Handler func(ctx context.Context, msg *Message) error

// Consumer consumes messages from RabbitMQ.
type Consumer struct {
	cfg     *config.Config
	conn    *amqp.Connection
	channel *amqp.Channel
	handler Handler
	logger  *slog.Logger
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewConsumer creates a new RabbitMQ consumer.
func NewConsumer(cfg *config.Config, handler Handler) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	if err := ch.Qos(cfg.PrefetchCount, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare queues
	queues := []string{cfg.ReminderQueue, cfg.DeadLetterQueue}
	for _, queue := range queues {
		_, err := ch.QueueDeclare(
			queue,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare queue %s: %w", queue, err)
		}
	}

	return &Consumer{
		cfg:     cfg,
		conn:    conn,
		channel: ch,
		handler: handler,
		logger:  slog.Default(),
		done:    make(chan struct{}),
	}, nil
}

// Start begins consuming messages.
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.cfg.ReminderQueue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("context cancelled, stopping consumer")
				return
			case <-c.done:
				c.logger.Info("consumer stopped")
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Info("channel closed")
					return
				}
				c.processMessage(ctx, msg)
			}
		}
	}()

	c.logger.Info("consumer started", "queue", c.cfg.ReminderQueue)
	return nil
}

// processMessage handles a single message.
func (c *Consumer) processMessage(ctx context.Context, delivery amqp.Delivery) {
	var msg Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		c.logger.Error("failed to unmarshal message", "error", err)
		// Reject and don't requeue malformed messages
		delivery.Reject(false)
		return
	}

	c.logger.Info("processing message",
		"job_id", msg.JobID,
		"document_id", msg.DocumentID,
		"priority", msg.Priority,
	)

	if err := c.handler(ctx, &msg); err != nil {
		c.logger.Error("handler failed",
			"job_id", msg.JobID,
			"error", err,
			"retry_count", msg.RetryCount,
		)

		// Check retry count
		if msg.RetryCount >= c.cfg.MaxRetryAttempts {
			c.logger.Error("max retries exceeded, sending to DLQ",
				"job_id", msg.JobID,
				"retry_count", msg.RetryCount,
			)
			c.sendToDLQ(&msg, err.Error())
			delivery.Reject(false)
			return
		}

		// Calculate backoff with exponential increase
		backoff := c.cfg.RetryBackoffDuration * time.Duration(1<<msg.RetryCount)
		c.logger.Info("scheduling retry with backoff",
			"job_id", msg.JobID,
			"retry_count", msg.RetryCount,
			"backoff", backoff,
		)

		// Requeue with incremented retry count and delay
		msg.RetryCount++
		c.requeueWithDelay(&msg, backoff)
		delivery.Reject(true)
		return
	}

	delivery.Ack(false)
	c.logger.Info("message processed successfully", "job_id", msg.JobID)
}

// requeue publishes a message back to the queue with updated retry count.
func (c *Consumer) requeue(msg *Message) {
	body, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error("failed to marshal message for requeue", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.channel.PublishWithContext(ctx,
		"",                  // exchange
		c.cfg.ReminderQueue, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		c.logger.Error("failed to requeue message", "error", err)
	}
}

// requeueWithDelay publishes a message back to the queue with a delay before retry.
func (c *Consumer) requeueWithDelay(msg *Message, delay time.Duration) {
	go func() {
		// Wait for backoff duration
		time.Sleep(delay)

		body, err := json.Marshal(msg)
		if err != nil {
			c.logger.Error("failed to marshal message for delayed requeue", "error", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = c.channel.PublishWithContext(ctx,
			"",                  // exchange
			c.cfg.ReminderQueue, // routing key
			false,               // mandatory
			false,               // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         body,
			},
		)
		if err != nil {
			c.logger.Error("failed to delayed requeue message", "error", err)
		} else {
			c.logger.Info("message requeued with delay",
				"job_id", msg.JobID,
				"retry_count", msg.RetryCount,
				"delay", delay,
			)
		}
	}()
}

// sendToDLQ sends a failed message to the dead-letter queue.
func (c *Consumer) sendToDLQ(msg *Message, errorMsg string) {
	body, err := json.Marshal(map[string]interface{}{
		"original_message": msg,
		"error":            errorMsg,
		"failed_at":        time.Now().UTC(),
	})
	if err != nil {
		c.logger.Error("failed to marshal DLQ message", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.channel.PublishWithContext(ctx,
		"",                    // exchange
		c.cfg.DeadLetterQueue, // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		c.logger.Error("failed to send to DLQ", "error", err)
	}
}

// Stop gracefully stops the consumer.
func (c *Consumer) Stop() error {
	close(c.done)
	c.wg.Wait()

	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}
