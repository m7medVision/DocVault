// Package queue provides RabbitMQ consumer for reminder jobs.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/docvault/reminder/internal/config"
)

// DLQMessage represents a message in the dead-letter queue.
type DLQMessage struct {
	OriginalMessage *Message  `json:"original_message"`
	Error           string    `json:"error"`
	FailedAt        time.Time `json:"failed_at"`
}

// DLQInspector provides utilities for inspecting and retrying DLQ messages.
type DLQInspector struct {
	cfg     *config.Config
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *slog.Logger
}

// NewDLQInspector creates a new DLQ inspector.
func NewDLQInspector(cfg *config.Config) (*DLQInspector, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare DLQ to ensure it exists
	_, err = ch.QueueDeclare(
		cfg.DeadLetterQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	return &DLQInspector{
		cfg:     cfg,
		conn:    conn,
		channel: ch,
		logger:  slog.Default(),
	}, nil
}

// List returns all messages in the DLQ without consuming them.
func (d *DLQInspector) List(ctx context.Context, limit int) ([]DLQMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var messages []DLQMessage

	for i := 0; i < limit; i++ {
		// Use Get to peek at messages without consuming
		msg, ok, err := d.channel.Get(d.cfg.DeadLetterQueue, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get message from DLQ: %w", err)
		}

		if !ok {
			// No more messages
			break
		}

		var dlqMsg DLQMessage
		if err := json.Unmarshal(msg.Body, &dlqMsg); err != nil {
			d.logger.Error("failed to unmarshal DLQ message", "error", err)
			// Reject and continue
			msg.Reject(false)
			continue
		}

		messages = append(messages, dlqMsg)

		// Reject to put it back (we're just peeking)
		msg.Reject(true)
	}

	return messages, nil
}

// Count returns the number of messages in the DLQ.
func (d *DLQInspector) Count(ctx context.Context) (int, error) {
	queue, err := d.channel.QueueInspect(d.cfg.DeadLetterQueue)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect DLQ: %w", err)
	}
	return queue.Messages, nil
}

// Retry message retries a single DLQ message by moving it back to the main queue.
func (d *DLQInspector) Retry(ctx context.Context, messageIndex int) error {
	// First, find and track the message
	for i := 0; ; i++ {
		msg, ok, err := d.channel.Get(d.cfg.DeadLetterQueue, false)
		if err != nil {
			return fmt.Errorf("failed to get message from DLQ: %w", err)
		}

		if !ok {
			return fmt.Errorf("message not found at index %d", messageIndex)
		}

		if i == messageIndex {
			// Found the message, reset retry count and republish
			var dlqMsg DLQMessage
			if err := json.Unmarshal(msg.Body, &dlqMsg); err != nil {
				msg.Reject(false)
				return fmt.Errorf("failed to unmarshal DLQ message: %w", err)
			}

			// Reset retry count
			if dlqMsg.OriginalMessage != nil {
				dlqMsg.OriginalMessage.RetryCount = 0
			}

			// Republish to main queue
			body, err := json.Marshal(dlqMsg.OriginalMessage)
			if err != nil {
				msg.Reject(false)
				return fmt.Errorf("failed to marshal message: %w", err)
			}

			err = d.channel.PublishWithContext(ctx,
				"",
				d.cfg.ReminderQueue,
				false,
				false,
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "application/json",
					Body:         body,
				},
			)
			if err != nil {
				msg.Reject(false)
				return fmt.Errorf("failed to republish message: %w", err)
			}

			// Acknowledge the original DLQ message
			msg.Ack(false)

			d.logger.Info("message retried successfully",
				"job_id", dlqMsg.OriginalMessage.JobID,
				"document_id", dlqMsg.OriginalMessage.DocumentID,
			)

			return nil
		}

		// Not our message, put it back
		msg.Reject(true)
	}
}

// RetryAll retries all messages in the DLQ.
func (d *DLQInspector) RetryAll(ctx context.Context) (int, error) {
	count := 0
	for {
		msg, ok, err := d.channel.Get(d.cfg.DeadLetterQueue, false)
		if err != nil {
			return count, fmt.Errorf("failed to get message from DLQ: %w", err)
		}

		if !ok {
			// No more messages
			break
		}

		var dlqMsg DLQMessage
		if err := json.Unmarshal(msg.Body, &dlqMsg); err != nil {
			d.logger.Error("failed to unmarshal DLQ message", "error", err)
			msg.Reject(false) // Discard malformed
			continue
		}

		// Reset retry count
		if dlqMsg.OriginalMessage != nil {
			dlqMsg.OriginalMessage.RetryCount = 0
		}

		// Republish to main queue
		body, err := json.Marshal(dlqMsg.OriginalMessage)
		if err != nil {
			d.logger.Error("failed to marshal message", "error", err)
			msg.Reject(false)
			continue
		}

		err = d.channel.PublishWithContext(ctx,
			"",
			d.cfg.ReminderQueue,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         body,
			},
		)
		if err != nil {
			d.logger.Error("failed to republish message", "error", err)
			msg.Reject(false)
			continue
		}

		msg.Ack(false)
		count++
	}

	d.logger.Info("all DLQ messages retried", "count", count)
	return count, nil
}

// Purge discards all messages in the DLQ.
func (d *DLQInspector) Purge(ctx context.Context) (int, error) {
	count, err := d.channel.QueuePurge(d.cfg.DeadLetterQueue, false)
	if err != nil {
		return 0, fmt.Errorf("failed to purge DLQ: %w", err)
	}

	d.logger.Info("DLQ purged", "count", count)
	return int(count), nil
}

// Close closes the DLQ inspector connection.
func (d *DLQInspector) Close() error {
	if err := d.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := d.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
