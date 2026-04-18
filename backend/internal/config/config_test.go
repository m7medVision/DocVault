package config

import "testing"

func TestLoadRejectsUnapprovedEmbeddingModel(t *testing.T) {
	t.Setenv("EMBEDDING_MODEL", "openai/text-embedding-3-small")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for unapproved embedding model")
	}
}

func TestLoadAcceptsOpenRouterEmbeddingModel(t *testing.T) {
	t.Setenv("EMBEDDING_MODEL", "openai/text-embedding-3-large")

	_, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
}

func TestLoadRequiresOpenRouterAPIKeyOutsideDevelopment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_AUDIENCE", "docvault-api")
	t.Setenv("OPENROUTER_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when OPENROUTER_API_KEY is missing in production")
	}
}

func TestLoadNormalizesLegacyRabbitMQURL(t *testing.T) {
	t.Setenv("EMBEDDING_MODEL", defaultEmbeddingModel)

	t.Setenv("RABBITMQ_URL", "amqp://docvault:changeme@localhost:5672//docvault")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := "amqp://docvault:changeme@localhost:5672/%2Fdocvault"
	if cfg.Queue.URL != want {
		t.Fatalf("Load() queue URL = %q, want %q", cfg.Queue.URL, want)
	}
}
