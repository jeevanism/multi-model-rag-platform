package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/llm"
	"multi-model-rag-platform/internal/observability"
	"multi-model-rag-platform/internal/rag"
	"multi-model-rag-platform/internal/service"
	"multi-model-rag-platform/internal/sse"
)

type rootResponse struct {
	Service   string   `json:"service"`
	Status    string   `json:"status"`
	Endpoints []string `json:"endpoints"`
	RequestID string   `json:"request_id"`
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	RequestID string `json:"request_id"`
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, rootResponse{
		Service:   "Multi-Model RAG API",
		Status:    "ok",
		Endpoints: []string{"/health", "/chat", "/chat/stream", "/evals/runs"},
		RequestID: observability.RequestIDFromContext(r.Context()),
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   "api",
		RequestID: observability.RequestIDFromContext(r.Context()),
	})
}

func handleDemoStatus(demoService auth.DemoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, DemoUnlockStatusResponse{
			Unlocked:      demoService.IsUnlocked(r),
			UnlockEnabled: demoService.UnlockEnabled(),
		})
	}
}

func handleDemoUnlock(demoService auth.DemoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeDemoUnlockRequest(r.Body)
		if err != nil {
			writeValidationError(w, err)
			return
		}
		if !demoService.ValidatePassword(req.Password) {
			writeError(w, http.StatusUnauthorized, "Invalid demo password")
			return
		}

		demoService.SetUnlockCookie(w)
		writeJSON(w, http.StatusOK, DemoUnlockStatusResponse{
			Unlocked:      true,
			UnlockEnabled: true,
		})
	}
}

func handleDemoLock(demoService auth.DemoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		demoService.ClearUnlockCookie(w)
		writeJSON(w, http.StatusOK, DemoUnlockStatusResponse{
			Unlocked:      false,
			UnlockEnabled: demoService.UnlockEnabled(),
		})
	}
}

func handleChat(
	demoService auth.DemoService,
	chatService service.ChatService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeChatRequest(r.Body)
		if err != nil {
			writeValidationError(w, err)
			return
		}

		result, err := chatService.GenerateChatResponse(service.ChatParams{
			Message:  req.Message,
			Provider: req.Provider,
			Model:    req.Model,
			RAG:      req.RAG,
			TopK:     req.TopK,
			Debug:    req.Debug,
		}, !demoService.IsUnlocked(r))
		if err != nil {
			var unsupported llm.UnsupportedProviderError
			switch {
			case errors.As(err, &unsupported):
				writeError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrRAGRepositoryRequired):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		writeJSON(w, http.StatusOK, ChatResponse{
			Answer:          result.Response.Answer,
			Provider:        result.Response.Provider,
			Model:           result.Response.Model,
			LatencyMS:       result.Response.LatencyMS,
			TokensIn:        result.Response.TokensIn,
			TokensOut:       result.Response.TokensOut,
			CostUSD:         result.Response.CostUSD,
			Citations:       result.Citations,
			RAGUsed:         result.RAGUsed,
			RetrievedChunks: toRetrievedChunkPreviews(result.RetrievedChunks, req.Debug),
		})
	}
}

func handleChatStream(
	demoService auth.DemoService,
	chatService service.ChatService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeChatRequest(r.Body)
		if err != nil {
			writeValidationError(w, err)
			return
		}

		result, err := chatService.GenerateChatResponse(service.ChatParams{
			Message:  req.Message,
			Provider: req.Provider,
			Model:    req.Model,
			RAG:      req.RAG,
			TopK:     req.TopK,
			Debug:    req.Debug,
		}, !demoService.IsUnlocked(r))
		if err != nil {
			var unsupported llm.UnsupportedProviderError
			switch {
			case errors.As(err, &unsupported):
				writeError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrRAGRepositoryRequired):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)

		if err := sse.WriteEvent(w, "start", map[string]string{
			"provider": result.Response.Provider,
			"model":    result.Response.Model,
		}); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
		}

		for _, token := range strings.Fields(result.Response.Answer) {
			if err := sse.WriteEvent(w, "token", map[string]string{"text": token}); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}

		if err := sse.WriteEvent(w, "end", map[string]any{
			"answer":     result.Response.Answer,
			"latency_ms": result.Response.LatencyMS,
			"tokens_in":  result.Response.TokensIn,
			"tokens_out": result.Response.TokensOut,
			"cost_usd":   result.Response.CostUSD,
		}); err != nil {
			_, _ = fmt.Fprint(w, "")
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
}

func toRetrievedChunkPreviews(
	chunks []rag.RetrievedChunk,
	debug bool,
) []RetrievedChunkPreview {
	if !debug || len(chunks) == 0 {
		return nil
	}

	previews := make([]RetrievedChunkPreview, 0, len(chunks))
	for _, chunk := range chunks {
		previews = append(previews, RetrievedChunkPreview{
			DocumentID: chunk.DocumentID,
			ChunkID:    chunk.ChunkID,
			ChunkIndex: chunk.ChunkIndex,
			Title:      chunk.Title,
			Content:    chunk.Content,
			Score:      chunk.Score,
		})
	}
	return previews
}

func handleIngestText(
	demoService auth.DemoService,
	ingestService service.IngestService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeIngestTextRequest(r.Body)
		if err != nil {
			writeValidationError(w, err)
			return
		}

		result, err := ingestService.IngestText(r.Context(), service.IngestParams{
			Title:        req.Title,
			Content:      req.Content,
			ChunkSize:    req.ChunkSize,
			ChunkOverlap: req.ChunkOverlap,
		}, !demoService.IsUnlocked(r))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, IngestTextResponse{
			DocumentID:        int(result.DocumentID),
			ChunkCount:        result.ChunkCount,
			EmbeddingCount:    result.EmbeddingCount,
			EmbeddingProvider: result.EmbeddingProvider,
			EmbeddingModel:    result.EmbeddingModel,
		})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeValidationError(w http.ResponseWriter, err error) {
	validationErr, ok := err.(ValidationError)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "Invalid request")
		return
	}
	writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
		"detail": []ValidationError{validationErr},
	})
}

func writeError(w http.ResponseWriter, statusCode int, detail string) {
	writeJSON(w, statusCode, map[string]string{"detail": detail})
}
