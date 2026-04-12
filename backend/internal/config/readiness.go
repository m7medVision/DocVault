package config

import (
	"context"
	"fmt"
)

type Dependencies struct {
	DB       Pingable
	Redis    PingableRedis
	RabbitMQ PingableRabbitMQ
}

type Pingable interface {
	Ping(context.Context) error
}

type PingableRedis interface {
	Ping(ctx context.Context) error
}

type PingableRabbitMQ interface {
	Ping() error
}

func CheckReadiness(ctx context.Context, cfg *Config, deps *Dependencies) error {
	if err := deps.DB.Ping(ctx); err != nil {
		return fmt.Errorf("db: %w", err)
	}
	if err := deps.Redis.Ping(ctx); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	if err := deps.RabbitMQ.Ping(); err != nil {
		return fmt.Errorf("rabbitmq: %w", err)
	}
	return nil
}
