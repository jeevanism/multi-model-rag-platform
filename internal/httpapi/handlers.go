package httpapi

import (
	"encoding/json"
	"net/http"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/observability"
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
