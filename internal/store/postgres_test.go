package store

import (
	"testing"

	"multi-model-rag-platform/internal/config"
)

func TestOpenPostgresReturnsErrorWhenDatabaseURLMissing(t *testing.T) {
	_, err := OpenPostgres(config.Config{})
	if err == nil {
		t.Fatal("expected error for missing database URL")
	}
}

func TestOpenPostgresReturnsStoreForConfiguredURL(t *testing.T) {
	store, err := OpenPostgres(config.Config{
		DatabaseURL: "postgresql://postgres:postgres@localhost:5432/multimodel_rag",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if store.DB() == nil {
		t.Fatal("expected database handle to be initialized")
	}
}
