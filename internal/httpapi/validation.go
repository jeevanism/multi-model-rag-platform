package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	defaultChatTopK        = 3
	defaultIngestChunkSize = 400
	defaultIngestChunkOver = 40
	maxChatTopK            = 10
	maxIngestChunkSize     = 4000
	maxIngestChunkOverlap  = 1000
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func decodeChatRequest(body io.Reader) (ChatRequest, error) {
	var req ChatRequest
	if err := decodeJSON(body, &req); err != nil {
		return ChatRequest{}, err
	}
	applyChatDefaults(&req)
	if err := validateChatRequest(req); err != nil {
		return ChatRequest{}, err
	}
	return req, nil
}

func decodeIngestTextRequest(body io.Reader) (IngestTextRequest, error) {
	var req IngestTextRequest
	if err := decodeJSON(body, &req); err != nil {
		return IngestTextRequest{}, err
	}
	applyIngestTextDefaults(&req)
	if err := validateIngestTextRequest(req); err != nil {
		return IngestTextRequest{}, err
	}
	return req, nil
}

func decodeDemoUnlockRequest(body io.Reader) (DemoUnlockRequest, error) {
	var req DemoUnlockRequest
	if err := decodeJSON(body, &req); err != nil {
		return DemoUnlockRequest{}, err
	}
	if strings.TrimSpace(req.Password) == "" {
		return DemoUnlockRequest{}, ValidationError{Field: "password", Message: "is required"}
	}
	return req, nil
}

func applyChatDefaults(req *ChatRequest) {
	if req.TopK == 0 {
		req.TopK = defaultChatTopK
	}
}

func applyIngestTextDefaults(req *IngestTextRequest) {
	if req.ChunkSize == 0 {
		req.ChunkSize = defaultIngestChunkSize
	}
	if req.ChunkOverlap == 0 {
		req.ChunkOverlap = defaultIngestChunkOver
	}
}

func validateChatRequest(req ChatRequest) error {
	if strings.TrimSpace(req.Message) == "" {
		return ValidationError{Field: "message", Message: "must be a non-empty string"}
	}
	if strings.TrimSpace(req.Provider) == "" {
		return ValidationError{Field: "provider", Message: "is required"}
	}
	if req.TopK <= 0 || req.TopK > maxChatTopK {
		return ValidationError{Field: "top_k", Message: "must be between 1 and 10"}
	}
	return nil
}

func validateIngestTextRequest(req IngestTextRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return ValidationError{Field: "title", Message: "must be a non-empty string"}
	}
	if strings.TrimSpace(req.Content) == "" {
		return ValidationError{Field: "content", Message: "must be a non-empty string"}
	}
	if req.ChunkSize <= 0 || req.ChunkSize > maxIngestChunkSize {
		return ValidationError{Field: "chunk_size", Message: "must be between 1 and 4000"}
	}
	if req.ChunkOverlap < 0 || req.ChunkOverlap > maxIngestChunkOverlap {
		return ValidationError{Field: "chunk_overlap", Message: "must be between 0 and 1000"}
	}
	return nil
}

func decodeJSON(body io.Reader, target any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return ValidationError{Message: "request body is required"}
		}
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return ValidationError{Message: "invalid JSON payload"}
		case errors.As(err, &typeErr):
			field := typeErr.Field
			if field == "" {
				field = typeErr.Struct
			}
			return ValidationError{Field: field, Message: "has the wrong type"}
		default:
			if strings.HasPrefix(err.Error(), "json: unknown field ") {
				field := strings.TrimPrefix(err.Error(), "json: unknown field ")
				field = strings.Trim(field, `"`)
				return ValidationError{Field: field, Message: "is not allowed"}
			}
			return ValidationError{Message: "invalid JSON payload"}
		}
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return ValidationError{Message: "request body must contain a single JSON object"}
	}

	return nil
}
