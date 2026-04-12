package httpapi

import (
	"strings"
	"testing"
)

func TestDecodeChatRequestAppliesDefaults(t *testing.T) {
	req, err := decodeChatRequest(strings.NewReader(`{"message":"hello","provider":"gemini"}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if req.TopK != 3 {
		t.Fatalf("expected top_k default 3, got %d", req.TopK)
	}
	if req.RAG {
		t.Fatal("expected rag default false")
	}
	if req.Debug {
		t.Fatal("expected debug default false")
	}
}

func TestDecodeChatRequestRejectsMissingMessage(t *testing.T) {
	_, err := decodeChatRequest(strings.NewReader(`{"provider":"gemini"}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "message" {
		t.Fatalf("expected field message, got %q", validationErr.Field)
	}
}

func TestDecodeChatRequestRejectsOutOfRangeTopK(t *testing.T) {
	_, err := decodeChatRequest(strings.NewReader(`{"message":"hello","provider":"gemini","top_k":11}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Field != "top_k" {
		t.Fatalf("expected field top_k, got %q", validationErr.Field)
	}
}

func TestDecodeChatRequestRejectsUnknownField(t *testing.T) {
	_, err := decodeChatRequest(strings.NewReader(`{"message":"hello","provider":"gemini","unexpected":true}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Field != "unexpected" {
		t.Fatalf("expected unexpected field error, got %q", validationErr.Field)
	}
}

func TestDecodeIngestTextRequestAppliesDefaults(t *testing.T) {
	req, err := decodeIngestTextRequest(strings.NewReader(`{"title":"Doc","content":"hello world"}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if req.ChunkSize != 400 {
		t.Fatalf("expected chunk_size default 400, got %d", req.ChunkSize)
	}
	if req.ChunkOverlap != 40 {
		t.Fatalf("expected chunk_overlap default 40, got %d", req.ChunkOverlap)
	}
}

func TestDecodeIngestTextRequestRejectsMissingContent(t *testing.T) {
	_, err := decodeIngestTextRequest(strings.NewReader(`{"title":"Doc"}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Field != "content" {
		t.Fatalf("expected field content, got %q", validationErr.Field)
	}
}

func TestDecodeIngestTextRequestRejectsOutOfRangeChunkSize(t *testing.T) {
	_, err := decodeIngestTextRequest(strings.NewReader(`{"title":"Doc","content":"hello","chunk_size":4001}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Field != "chunk_size" {
		t.Fatalf("expected field chunk_size, got %q", validationErr.Field)
	}
}

func TestDecodeIngestTextRequestRejectsNegativeChunkOverlap(t *testing.T) {
	_, err := decodeIngestTextRequest(strings.NewReader(`{"title":"Doc","content":"hello","chunk_overlap":-1}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Field != "chunk_overlap" {
		t.Fatalf("expected field chunk_overlap, got %q", validationErr.Field)
	}
}

func TestDecodeDemoUnlockRequestRejectsBlankPassword(t *testing.T) {
	_, err := decodeDemoUnlockRequest(strings.NewReader(`{"password":"   "}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Field != "password" {
		t.Fatalf("expected field password, got %q", validationErr.Field)
	}
}

func TestDecodeJSONRejectsMultipleObjects(t *testing.T) {
	_, err := decodeChatRequest(strings.NewReader(`{"message":"hello","provider":"gemini"}{"message":"hello","provider":"gemini"}`))
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(ValidationError)
	if validationErr.Message != "request body must contain a single JSON object" {
		t.Fatalf("unexpected validation error: %+v", validationErr)
	}
}
