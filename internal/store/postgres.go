package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"multi-model-rag-platform/internal/config"
)

type PostgresStore struct {
	db *sql.DB
}

func OpenPostgres(cfg config.Config) (*PostgresStore, error) {
	if cfg.DatabaseURL == "" {
		return nil, errors.New("database URL is required")
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStore) DB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}

func (s *PostgresStore) CheckConnection(ctx context.Context) error {
	var value int
	return s.db.QueryRowContext(ctx, "SELECT 1").Scan(&value)
}

func (s *PostgresStore) WithTx(
	ctx context.Context,
	fn func(*sql.Tx) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *PostgresStore) InsertDocument(
	ctx context.Context,
	tx *sql.Tx,
	title string,
	content string,
) (int64, error) {
	var documentID int64
	err := tx.QueryRowContext(
		ctx,
		`
		INSERT INTO documents (title, content)
		VALUES ($1, $2)
		RETURNING id
		`,
		title,
		content,
	).Scan(&documentID)
	return documentID, err
}

func (s *PostgresStore) InsertChunk(
	ctx context.Context,
	tx *sql.Tx,
	documentID int64,
	chunkIndex int,
	content string,
) (int64, error) {
	var chunkID int64
	err := tx.QueryRowContext(
		ctx,
		`
		INSERT INTO chunks (document_id, chunk_index, content)
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		documentID,
		chunkIndex,
		content,
	).Scan(&chunkID)
	return chunkID, err
}

func (s *PostgresStore) InsertEmbedding(
	ctx context.Context,
	tx *sql.Tx,
	chunkID int64,
	vectorLiteral string,
	provider string,
	model string,
) error {
	_, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO embeddings (chunk_id, provider, model, embedding)
		VALUES ($1, $2, $3, CAST($4 AS vector(8)))
		`,
		chunkID,
		provider,
		model,
		vectorLiteral,
	)
	return err
}
