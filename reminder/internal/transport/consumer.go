package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, delivery amqp.Delivery) error

type QueueConsumer struct {
	conn    *RabbitMQConnection
	handler MessageHandler
	retries int
	done    chan struct{}
	wg      sync.WaitGroup
	logger  *slog.Logger
}

func NewQueueConsumer(conn *RabbitMQConnection, handler MessageHandler, retries int) *QueueConsumer {
	return &QueueConsumer{
		conn:    conn,
		handler: handler,
		retries: retries,
		done:    make(chan struct{}),
		logger:  slog.Default(),
	}
}

func (c *QueueConsumer) Start(ctx context.Context) error {
	msgs, err := c.conn.Channel().Consume(
		c.conn.config.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
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
				c.handleMessage(ctx, msg)
			}
		}
	}()

	c.logger.Info("consumer started", "queue", c.conn.config.Queue)
	return nil
}

func (c *QueueConsumer) handleMessage(ctx context.Context, delivery amqp.Delivery) {
	var msg QueueMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		c.logger.Error("failed to unmarshal message", "error", err)
		delivery.Reject(false)
		return
	}

	c.logger.Info("processing message",
		"job_id", msg.JobID,
		"document_id", msg.DocumentID,
		"priority", msg.Priority,
	)

	if err := c.handler(ctx, delivery); err != nil {
		c.logger.Error("handler failed",
			"job_id", msg.JobID,
			"error", err,
			"retry_count", msg.RetryCount,
		)

		if msg.RetryCount >= c.retries {
			c.logger.Error("max retries exceeded, sending to DLQ",
				"job_id", msg.JobID,
				"retry_count", msg.RetryCount,
			)
			c.sendToDLQ(&msg, err.Error())
			delivery.Reject(false)
			return
		}

		backoff := 5 * time.Minute * time.Duration(1<<msg.RetryCount)
		c.logger.Info("scheduling retry with backoff",
			"job_id", msg.JobID,
			"retry_count", msg.RetryCount,
			"backoff", backoff,
		)

		msg.RetryCount++
		c.requeueWithDelay(&msg, backoff)
		delivery.Reject(false)
		return
	}

	delivery.Ack(false)
	c.logger.Info("message processed successfully", "job_id", msg.JobID)
}

func (c *QueueConsumer) requeue(msg *QueueMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error("failed to marshal message for requeue", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.conn.Channel().PublishWithContext(ctx,
		"",
		c.conn.config.Queue,
		false,
		false,
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

func (c *QueueConsumer) requeueWithDelay(msg *QueueMessage, delay time.Duration) {
	go func() {
		time.Sleep(delay)

		body, err := json.Marshal(msg)
		if err != nil {
			c.logger.Error("failed to marshal message for delayed requeue", "error", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = c.conn.Channel().PublishWithContext(ctx,
			"",
			c.conn.config.Queue,
			false,
			false,
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

func (c *QueueConsumer) sendToDLQ(msg *QueueMessage, errorMsg string) {
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

	err = c.conn.Channel().PublishWithContext(ctx,
		"",
		c.conn.config.DLQ,
		false,
		false,
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

func (c *QueueConsumer) Stop() error {
	close(c.done)
	c.wg.Wait()
	return c.conn.Close()
}

type Delivery struct {
	amqp.Delivery
}

func (d *Delivery) Unmarshal(v interface{}) error {
	return json.Unmarshal(d.Body, v)
}

type QueueMessage struct {
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
