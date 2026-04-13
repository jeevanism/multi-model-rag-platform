package service

import (
	"context"
	"database/sql"
	"errors"

	"multi-model-rag-platform/internal/rag"
)

var ErrNoChunksGenerated = errors.New("No chunks generated from content")

type IngestParams struct {
	Title        string
	Content      string
	ChunkSize    int
	ChunkOverlap int
}

type IngestResult struct {
	DocumentID        int64
	ChunkCount        int
	EmbeddingCount    int
	EmbeddingProvider string
	EmbeddingModel    string
}

type IngestRepository interface {
	WithTx(ctx context.Context, fn func(*sql.Tx) error) error
	InsertDocument(ctx context.Context, tx *sql.Tx, title string, content string) (int64, error)
	InsertChunk(ctx context.Context, tx *sql.Tx, documentID int64, chunkIndex int, content string) (int64, error)
	InsertEmbedding(
		ctx context.Context,
		tx *sql.Tx,
		chunkID int64,
		vectorLiteral string,
		provider string,
		model string,
	) error
}

type IngestService struct {
	repo IngestRepository
}

func NewIngestService(repo IngestRepository) IngestService {
	return IngestService{repo: repo}
}

func (s IngestService) IngestText(
	ctx context.Context,
	req IngestParams,
	forceStub bool,
) (IngestResult, error) {
	chunks, err := rag.ChunkText(req.Content, req.ChunkSize, req.ChunkOverlap)
	if err != nil {
		return IngestResult{}, err
	}
	if len(chunks) == 0 {
		return IngestResult{}, ErrNoChunksGenerated
	}

	var result IngestResult
	err = s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		documentID, err := s.repo.InsertDocument(ctx, tx, req.Title, req.Content)
		if err != nil {
			return err
		}

		embeddingCount := 0
		firstProvider := ""
		firstModel := ""

		for chunkIndex, chunk := range chunks {
			chunkID, err := s.repo.InsertChunk(ctx, tx, documentID, chunkIndex, chunk)
			if err != nil {
				return err
			}

			embedding, err := rag.EmbedText(chunk, rag.EmbeddingDim)
			if err != nil {
				return err
			}

			if err := s.repo.InsertEmbedding(
				ctx,
				tx,
				chunkID,
				rag.ToPGVectorLiteral(embedding.Vector),
				embedding.Provider,
				embedding.Model,
			); err != nil {
				return err
			}

			if chunkIndex == 0 {
				firstProvider = embedding.Provider
				firstModel = embedding.Model
			}
			embeddingCount++
		}

		result = IngestResult{
			DocumentID:        documentID,
			ChunkCount:        len(chunks),
			EmbeddingCount:    embeddingCount,
			EmbeddingProvider: firstProvider,
			EmbeddingModel:    firstModel,
		}
		return nil
	})
	if err != nil {
		return IngestResult{}, err
	}

	_ = forceStub
	return result, nil
}
