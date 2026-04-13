package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"multi-model-rag-platform/internal/config"
	"multi-model-rag-platform/internal/rag"
	"multi-model-rag-platform/internal/service"
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

func (s *PostgresStore) RetrieveChunks(
	query string,
	topK int,
) ([]rag.RetrievedChunk, error) {
	embedding, err := rag.EmbedText(query, rag.EmbeddingDim)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		`
		SELECT
			d.id AS document_id,
			c.id AS chunk_id,
			c.chunk_index AS chunk_index,
			d.title AS title,
			c.content AS content,
			(e.embedding <-> CAST($1 AS vector(8))) AS distance
		FROM embeddings e
		JOIN chunks c ON c.id = e.chunk_id
		JOIN documents d ON d.id = c.document_id
		ORDER BY e.embedding <-> CAST($1 AS vector(8)) ASC
		LIMIT $2
		`,
		rag.ToPGVectorLiteral(embedding.Vector),
		topK,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]rag.RetrievedChunk, 0)
	for rows.Next() {
		var item rag.RetrievedChunk
		if err := rows.Scan(
			&item.DocumentID,
			&item.ChunkID,
			&item.ChunkIndex,
			&item.Title,
			&item.Content,
			&item.Score,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (s *PostgresStore) ListEvalRuns(limit int) ([]service.EvalRunSummary, error) {
	rows, err := s.db.Query(
		`
		SELECT
			id,
			dataset_name,
			provider,
			model,
			total_cases,
			passed_cases,
			avg_latency_ms,
			created_at
		FROM eval_run
		ORDER BY id DESC
		LIMIT $1
		`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]service.EvalRunSummary, 0)
	for rows.Next() {
		var item service.EvalRunSummary
		if err := rows.Scan(
			&item.ID,
			&item.DatasetName,
			&item.Provider,
			&item.Model,
			&item.TotalCases,
			&item.PassedCases,
			&item.AvgLatencyMS,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (s *PostgresStore) GetEvalRunDetail(evalRunID int) (service.EvalRunDetailResult, bool, error) {
	var run service.EvalRunSummary
	err := s.db.QueryRow(
		`
		SELECT
			id,
			dataset_name,
			provider,
			model,
			total_cases,
			passed_cases,
			avg_latency_ms,
			created_at
		FROM eval_run
		WHERE id = $1
		`,
		evalRunID,
	).Scan(
		&run.ID,
		&run.DatasetName,
		&run.Provider,
		&run.Model,
		&run.TotalCases,
		&run.PassedCases,
		&run.AvgLatencyMS,
		&run.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return service.EvalRunDetailResult{}, false, nil
	}
	if err != nil {
		return service.EvalRunDetailResult{}, false, err
	}

	rows, err := s.db.Query(
		`
		SELECT
			id,
			case_id,
			question,
			passed,
			latency_ms,
			correctness_score,
			groundedness_score,
			hallucination_score,
			citations,
			error
		FROM eval_run_case
		WHERE eval_run_id = $1
		ORDER BY id ASC
		`,
		evalRunID,
	)
	if err != nil {
		return service.EvalRunDetailResult{}, false, err
	}
	defer rows.Close()

	cases := make([]service.EvalRunCase, 0)
	for rows.Next() {
		var item service.EvalRunCase
		var citations []string
		if err := rows.Scan(
			&item.ID,
			&item.CaseID,
			&item.Question,
			&item.Passed,
			&item.LatencyMS,
			&item.CorrectnessScore,
			&item.GroundednessScore,
			&item.HallucinationScore,
			&citations,
			&item.Error,
		); err != nil {
			return service.EvalRunDetailResult{}, false, err
		}
		item.Citations = citations
		cases = append(cases, item)
	}
	if err := rows.Err(); err != nil {
		return service.EvalRunDetailResult{}, false, err
	}

	return service.EvalRunDetailResult{
		Run:   run,
		Cases: cases,
	}, true, nil
}
