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
		}, !demoService.IsUnlocked(r))
		if err != nil {
			var unsupported llm.UnsupportedProviderError
			switch {
			case errors.As(err, &unsupported):
				writeError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrRAGNotImplemented):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		writeJSON(w, http.StatusOK, ChatResponse{
			Answer:          result.Answer,
			Provider:        result.Provider,
			Model:           result.Model,
			LatencyMS:       result.LatencyMS,
			TokensIn:        result.TokensIn,
			TokensOut:       result.TokensOut,
			CostUSD:         result.CostUSD,
			Citations:       []string{},
			RAGUsed:         false,
			RetrievedChunks: nil,
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
		}, !demoService.IsUnlocked(r))
		if err != nil {
			var unsupported llm.UnsupportedProviderError
			switch {
			case errors.As(err, &unsupported):
				writeError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrRAGNotImplemented):
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
			"provider": result.Provider,
			"model":    result.Model,
		}); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
		}

		for _, token := range strings.Fields(result.Answer) {
			if err := sse.WriteEvent(w, "token", map[string]string{"text": token}); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}

		if err := sse.WriteEvent(w, "end", map[string]any{
			"answer":     result.Answer,
			"latency_ms": result.LatencyMS,
			"tokens_in":  result.TokensIn,
			"tokens_out": result.TokensOut,
			"cost_usd":   result.CostUSD,
		}); err != nil {
			_, _ = fmt.Fprint(w, "")
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
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
