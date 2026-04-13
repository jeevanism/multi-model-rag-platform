package service

import (
	"context"
	"database/sql"
	"testing"
)

type fakeIngestRepo struct {
	documentID      int64
	chunkID         int64
	insertedChunks  []string
	insertedVectors []string
}

func (f *fakeIngestRepo) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	return fn(nil)
}

func (f *fakeIngestRepo) InsertDocument(ctx context.Context, tx *sql.Tx, title string, content string) (int64, error) {
	return f.documentID, nil
}

func (f *fakeIngestRepo) InsertChunk(ctx context.Context, tx *sql.Tx, documentID int64, chunkIndex int, content string) (int64, error) {
	f.insertedChunks = append(f.insertedChunks, content)
	return f.chunkID + int64(chunkIndex), nil
}

func (f *fakeIngestRepo) InsertEmbedding(
	ctx context.Context,
	tx *sql.Tx,
	chunkID int64,
	vectorLiteral string,
	provider string,
	model string,
) error {
	f.insertedVectors = append(f.insertedVectors, vectorLiteral)
	return nil
}

func TestIngestTextReturnsSummary(t *testing.T) {
	repo := &fakeIngestRepo{documentID: 1, chunkID: 10}
	service := NewIngestService(repo)

	result, err := service.IngestText(context.Background(), IngestParams{
		Title:        "Doc",
		Content:      "hello world this is a document",
		ChunkSize:    20,
		ChunkOverlap: 5,
	}, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.DocumentID != 1 {
		t.Fatalf("expected document id 1, got %d", result.DocumentID)
	}
	if result.ChunkCount == 0 {
		t.Fatal("expected chunks to be generated")
	}
	if result.EmbeddingCount != result.ChunkCount {
		t.Fatalf("expected embedding count to match chunk count, got %d vs %d", result.EmbeddingCount, result.ChunkCount)
	}
	if result.EmbeddingProvider != "stub" {
		t.Fatalf("expected stub provider, got %q", result.EmbeddingProvider)
	}
	if result.EmbeddingModel != "stub-embedding-v1" {
		t.Fatalf("expected stub model, got %q", result.EmbeddingModel)
	}
}

func TestIngestTextReturnsErrorWhenNoChunksGenerated(t *testing.T) {
	repo := &fakeIngestRepo{}
	service := NewIngestService(repo)

	_, err := service.IngestText(context.Background(), IngestParams{
		Title:        "Doc",
		Content:      "   \n\t",
		ChunkSize:    20,
		ChunkOverlap: 5,
	}, true)
	if err == nil {
		t.Fatal("expected error")
	}
}
