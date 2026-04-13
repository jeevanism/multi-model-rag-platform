package config

import (
	"os"
	"testing"
)

func TestNormalizeDatabaseURLStripsPsycopgDriverPrefix(t *testing.T) {
	raw := "postgresql+psycopg://postgres:postgres@localhost:5432/multimodel_rag"

	got := normalizeDatabaseURL(raw)

	if got != "postgresql://postgres:postgres@localhost:5432/multimodel_rag" {
		t.Fatalf("unexpected normalized URL %q", got)
	}
}

func TestDatabaseURLPrefersPSQLDatabaseURL(t *testing.T) {
	t.Setenv("PSQL_DATABASE_URL", "postgresql://preferred")
	t.Setenv("DATABASE_URL", "postgresql+psycopg://fallback")

	got := databaseURL()

	if got != "postgresql://preferred" {
		t.Fatalf("expected PSQL_DATABASE_URL to win, got %q", got)
	}
}

func TestLoadSetsDatabaseURL(t *testing.T) {
	originalDatabaseURL, hadDatabaseURL := os.LookupEnv("DATABASE_URL")
	originalPSQLDatabaseURL, hadPSQLDatabaseURL := os.LookupEnv("PSQL_DATABASE_URL")
	t.Cleanup(func() {
		if hadDatabaseURL {
			_ = os.Setenv("DATABASE_URL", originalDatabaseURL)
		} else {
			_ = os.Unsetenv("DATABASE_URL")
		}
		if hadPSQLDatabaseURL {
			_ = os.Setenv("PSQL_DATABASE_URL", originalPSQLDatabaseURL)
		} else {
			_ = os.Unsetenv("PSQL_DATABASE_URL")
		}
	})

	_ = os.Unsetenv("PSQL_DATABASE_URL")
	_ = os.Setenv("DATABASE_URL", "postgresql+psycopg://postgres:postgres@localhost:5432/app")

	cfg := Load()

	if cfg.DatabaseURL != "postgresql://postgres:postgres@localhost:5432/app" {
		t.Fatalf("unexpected config database URL %q", cfg.DatabaseURL)
	}
}
