package httpapi

import (
	"encoding/json"
	"net/http"

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

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
