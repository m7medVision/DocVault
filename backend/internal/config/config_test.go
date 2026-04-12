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
	t.Setenv("EMBEDDING_MODEL", "mistralai/mistral-embed-2312")

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

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when OPENROUTER_API_KEY is missing in production")
	}
}
