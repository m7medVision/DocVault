package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
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

var permanentErrorMarkers = []string{
	"status 400",
	"status 401",
	"status 403",
	"status 404",
	"status 422",
	"invalid_json",
	"malformed",
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
		errorMsg := "invalid_json: " + err.Error()
		c.logger.Error("failed to unmarshal message", "error", errorMsg)
		if c.sendRawToDLQ(delivery.Body, errorMsg) {
			delivery.Ack(false)
		} else {
			delivery.Reject(true)
		}
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

		if isPermanentError(err) || msg.RetryCount >= c.retries {
			c.logger.Error("max retries exceeded, sending to DLQ",
				"job_id", msg.JobID,
				"retry_count", msg.RetryCount,
			)
			if c.sendToDLQ(&msg, err.Error()) {
				delivery.Ack(false)
			} else {
				delivery.Reject(true)
			}
			return
		}

		backoff := 5 * time.Minute * time.Duration(1<<msg.RetryCount)
		c.logger.Info("scheduling retry with backoff",
			"job_id", msg.JobID,
			"retry_count", msg.RetryCount,
			"backoff", backoff,
		)

		msg.RetryCount++
		time.Sleep(backoff)
		if err := c.requeue(&msg); err != nil {
			c.logger.Error("failed to requeue message", "error", err)
			delivery.Reject(true)
		} else {
			delivery.Ack(false)
		}
		return
	}

	delivery.Ack(false)
	c.logger.Info("message processed successfully", "job_id", msg.JobID)
}

func (c *QueueConsumer) requeue(msg *QueueMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.publish(c.conn.config.Queue, body)
}

func (c *QueueConsumer) sendToDLQ(msg *QueueMessage, errorMsg string) bool {
	body, err := json.Marshal(map[string]interface{}{
		"original_queue":   c.conn.config.Queue,
		"original_message": msg,
		"error":            errorMsg,
		"failed_at":        time.Now().UTC(),
	})
	if err != nil {
		c.logger.Error("failed to marshal DLQ message", "error", err)
		return false
	}

	if err := c.publish(c.conn.config.DLQ, body); err != nil {
		c.logger.Error("failed to send to DLQ", "error", err)
		return false
	}

	return true
}

func (c *QueueConsumer) sendRawToDLQ(body []byte, errorMsg string) bool {
	envelope, err := json.Marshal(map[string]interface{}{
		"original_queue":   c.conn.config.Queue,
		"original_message": nil,
		"original_body":    string(body),
		"error":            errorMsg,
		"failed_at":        time.Now().UTC(),
	})
	if err != nil {
		c.logger.Error("failed to marshal raw DLQ message", "error", err)
		return false
	}

	if err := c.publish(c.conn.config.DLQ, envelope); err != nil {
		c.logger.Error("failed to send raw message to DLQ", "error", err)
		return false
	}

	return true
}

func (c *QueueConsumer) publish(queue string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.conn.Channel().PublishWithContext(ctx,
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func isPermanentError(err error) bool {
	lower := strings.ToLower(err.Error())
	for _, marker := range permanentErrorMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
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
