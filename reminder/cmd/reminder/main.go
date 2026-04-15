// Package main is the entry point for the DocVault Worker Service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/docvault/reminder/internal/application"
	"github.com/docvault/reminder/internal/config"
	"github.com/docvault/reminder/internal/database"
	"github.com/docvault/reminder/internal/queue"
	"github.com/docvault/reminder/internal/reminder"
	"github.com/docvault/reminder/internal/sentry"
	"github.com/docvault/reminder/internal/telemetry"
	"github.com/docvault/reminder/internal/transport"
)

func main() {
	loadRepoEnv()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	telemetryCtx, telemetryCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer telemetryCancel()

	tp, err := telemetry.Init(telemetryCtx, cfg, "docvault-worker")
	if err != nil {
		slog.Warn("failed to initialize telemetry, continuing without", "error", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown telemetry", "error", err)
		}
		if !sentry.Flush(shutdownCtx) {
			slog.Error("failed to flush Sentry")
		}
	}()

	if err := sentry.Init(telemetryCtx, cfg.SentryDSN, cfg.Environment, "docvault-worker"); err != nil {
		slog.Warn("failed to initialize Sentry, continuing without", "error", err)
	}

	_ = tp

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "dlq":
			runDLQCommand(cfg)
			return
		case "worker":
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Usage: worker [command]")
			fmt.Println("Commands:")
			fmt.Println("  dlq list      - List messages in dead-letter queue")
			fmt.Println("  dlq count     - Count messages in dead-letter queue")
			fmt.Println("  dlq retry N   - Retry message at index N")
			fmt.Println("  dlq retry-all - Retry all messages in DLQ")
			fmt.Println("  dlq purge     - Purge all messages in DLQ")
			fmt.Println("  worker        - Run the worker (default)")
			os.Exit(1)
		}
	}

	runWorker(cfg)
}

func loadRepoEnv() {
	root, err := findRepoRoot()
	if err != nil {
		return
	}

	_ = godotenv.Load(filepath.Join(root, ".env"))
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "justfile")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found")
		}
		dir = parent
	}
}

func runDLQCommand(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	inspector, err := queue.NewDLQInspector(cfg)
	if err != nil {
		slog.Error("failed to create DLQ inspector", "error", err)
		os.Exit(1)
	}
	defer inspector.Close()

	if len(os.Args) < 3 {
		fmt.Println("Usage: worker dlq [list|count|retry N|retry-all|purge]")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		messages, err := inspector.List(ctx, 20)
		if err != nil {
			slog.Error("failed to list DLQ messages", "error", err)
			os.Exit(1)
		}
		if len(messages) == 0 {
			fmt.Println("No messages in DLQ")
			return
		}
		fmt.Printf("Found %d messages in DLQ:\n", len(messages))
		for i, msg := range messages {
			jobID := "<unavailable>"
			documentID := "<unavailable>"
			if msg.OriginalMessage != nil {
				if msg.OriginalMessage.JobID != "" {
					jobID = msg.OriginalMessage.JobID
				}
				if msg.OriginalMessage.DocumentID != "" {
					documentID = msg.OriginalMessage.DocumentID
				}
			}
			fmt.Printf("\n[%d] JobID: %s, DocumentID: %s\n", i, jobID, documentID)
			fmt.Printf("    Error: %s\n", msg.Error)
			fmt.Printf("    FailedAt: %s\n", msg.FailedAt.Format(time.RFC3339))
			if msg.OriginalQueue != "" {
				fmt.Printf("    OriginalQueue: %s\n", msg.OriginalQueue)
			}
			if msg.OriginalMessage != nil {
				fmt.Printf("    RetryCount: %d\n", msg.OriginalMessage.RetryCount)
			}
			if msg.OriginalBody != "" {
				fmt.Printf("    OriginalBody: %s\n", msg.OriginalBody)
			}
		}

	case "count":
		count, err := inspector.Count(ctx)
		if err != nil {
			slog.Error("failed to count DLQ messages", "error", err)
			os.Exit(1)
		}
		fmt.Printf("DLQ contains %d messages\n", count)

	case "retry":
		if len(os.Args) < 4 {
			fmt.Println("Usage: worker dlq retry N")
			os.Exit(1)
		}
		index, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid index: %s\n", os.Args[3])
			os.Exit(1)
		}
		if err := inspector.Retry(ctx, index); err != nil {
			slog.Error("failed to retry message", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Message at index %d retried successfully\n", index)

	case "retry-all":
		count, err := inspector.RetryAll(ctx)
		if err != nil {
			slog.Error("failed to retry all messages", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Retried %d messages\n", count)

	case "purge":
		count, err := inspector.Purge(ctx)
		if err != nil {
			slog.Error("failed to purge DLQ", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Purged %d messages\n", count)

	default:
		fmt.Printf("Unknown DLQ command: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func runWorker(cfg *config.Config) {
	logger := slog.Default()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.NewReminderStore(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.InitSchema(ctx); err != nil {
		slog.Error("failed to initialize reminder schema", "error", err)
		os.Exit(1)
	}

	extractor := reminder.NewReminderExtractor(cfg.NotifyDaysBefore)
	var dateExtractor reminder.DateExtractor
	if cfg.ReminderExtractionEnabled {
		dateExtractor = reminder.NewOpenRouterDateExtractor(
			cfg.OpenRouterAPIKey,
			cfg.ReminderExtractionModel,
			cfg.ReminderExtractionMaxChars,
			nil,
		)
	}
	handler := application.NewReminderJobHandler(extractor, dateExtractor, db)

	conn, err := transport.NewRabbitMQConnection(transport.RabbitMQConfig{
		URL:      cfg.RabbitMQURL,
		Queue:    cfg.ReminderQueue,
		DLQ:      cfg.DeadLetterQueue,
		Prefetch: cfg.PrefetchCount,
	})
	if err != nil {
		slog.Error("failed to create RabbitMQ connection", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	consumer := transport.NewQueueConsumer(conn, handler.Handle, cfg.MaxRetryAttempts)
	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start consumer", "error", err)
		os.Exit(1)
	}

	logger.Info("worker service started",
		"environment", cfg.Environment,
		"queue", cfg.ReminderQueue,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down worker service...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	_ = shutdownCtx

	if err := consumer.Stop(); err != nil {
		logger.Error("failed to stop consumer", "error", err)
	}

	logger.Info("worker service stopped")
}
