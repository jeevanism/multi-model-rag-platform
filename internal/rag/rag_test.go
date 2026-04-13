package rag

import (
	"strings"
	"testing"
)

func TestChunkTextRespectsMaxChars(t *testing.T) {
	text := strings.Repeat("word ", 80)

	chunks, err := ChunkText(text, 50, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	for _, chunk := range chunks {
		if len(chunk) > 50 {
			t.Fatalf("expected chunk length <= 50, got %d", len(chunk))
		}
	}
}

func TestChunkTextReturnsEmptyForBlankInput(t *testing.T) {
	chunks, err := ChunkText("   \n\t  ", 400, 40)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected no chunks, got %#v", chunks)
	}
}

func TestChunkTextRejectsInvalidOverlap(t *testing.T) {
	_, err := ChunkText("hello", 10, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeterministicEmbeddingIsStableAndPGVectorFormatted(t *testing.T) {
	vector1, err := EmbedTextDeterministic("hello", EmbeddingDim)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	vector2, err := EmbedTextDeterministic("hello", EmbeddingDim)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	vector3, err := EmbedTextDeterministic("different", EmbeddingDim)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(vector1) != 8 {
		t.Fatalf("expected vector length 8, got %d", len(vector1))
	}
	if ToPGVectorLiteral(vector1)[0] != '[' {
		t.Fatalf("expected pgvector literal, got %q", ToPGVectorLiteral(vector1))
	}
	if !equalVectors(vector1, vector2) {
		t.Fatal("expected identical embeddings for identical text")
	}
	if equalVectors(vector1, vector3) {
		t.Fatal("expected different embeddings for different text")
	}
}

func TestEmbedTextDefaultsToStub(t *testing.T) {
	result, err := EmbedText("hello", EmbeddingDim)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Provider != StubEmbeddingProvider {
		t.Fatalf("expected stub provider, got %q", result.Provider)
	}
	if result.Model != StubEmbeddingModel {
		t.Fatalf("expected stub model, got %q", result.Model)
	}
	if len(result.Vector) != EmbeddingDim {
		t.Fatalf("expected vector length %d, got %d", EmbeddingDim, len(result.Vector))
	}
}

func TestBuildGroundedPromptIncludesCitationsAndContext(t *testing.T) {
	prompt := BuildGroundedPrompt("What is in the doc?", []RetrievedChunk{
		{Title: "Test Doc", ChunkIndex: 0, Content: "Important context."},
	})

	if !strings.Contains(prompt, "[source:Test Doc#chunk=0] Important context.") {
		t.Fatalf("expected citation and content in prompt, got %q", prompt)
	}
	if !strings.Contains(prompt, "Question:\nWhat is in the doc?") {
		t.Fatalf("expected question in prompt, got %q", prompt)
	}
}

func equalVectors(left []float64, right []float64) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
