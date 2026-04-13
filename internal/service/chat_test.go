package service

import (
	"strings"
	"testing"

	"multi-model-rag-platform/internal/config"
	"multi-model-rag-platform/internal/llm"
	"multi-model-rag-platform/internal/rag"
)

type fakeChatRepo struct {
	chunks []rag.RetrievedChunk
}

func (f fakeChatRepo) RetrieveChunks(query string, topK int) ([]rag.RetrievedChunk, error) {
	return f.chunks, nil
}

func TestGenerateChatResponseReturnsRAGCitationsAndDebugChunks(t *testing.T) {
	service := NewChatService(
		llm.NewRouter(configForTests()),
		fakeChatRepo{
			chunks: []rag.RetrievedChunk{
				{
					DocumentID: 1,
					ChunkID:    10,
					ChunkIndex: 0,
					Title:      "Test Doc",
					Content:    "Important context.",
					Score:      0.123,
				},
			},
		},
	)

	result, err := service.GenerateChatResponse(ChatParams{
		Message:  "what is in the doc?",
		Provider: "gemini",
		RAG:      true,
		TopK:     2,
		Debug:    true,
	}, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.RAGUsed {
		t.Fatal("expected rag_used true")
	}
	if len(result.Citations) != 1 || result.Citations[0] != "[source:Test Doc#chunk=0]" {
		t.Fatalf("unexpected citations %#v", result.Citations)
	}
	if len(result.RetrievedChunks) != 1 || result.RetrievedChunks[0].Title != "Test Doc" {
		t.Fatalf("unexpected retrieved chunks %#v", result.RetrievedChunks)
	}
	if result.Response.Answer == "[stub:gemini] what is in the doc?" {
		t.Fatal("expected grounded prompt to be used for generation")
	}
	if !contains(result.Response.Answer, "[source:Test Doc#chunk=0]") {
		t.Fatalf("expected citations appended to answer, got %q", result.Response.Answer)
	}
}

func TestGenerateChatResponseErrorsWhenRAGRepositoryMissing(t *testing.T) {
	service := NewChatService(llm.NewRouter(configForTests()), nil)

	_, err := service.GenerateChatResponse(ChatParams{
		Message:  "hello",
		Provider: "gemini",
		RAG:      true,
		TopK:     2,
	}, true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func configForTests() config.Config {
	return config.Config{LLMProviderMode: "stub"}
}

func contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}
