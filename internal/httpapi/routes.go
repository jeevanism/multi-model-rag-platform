package httpapi

import (
	"net/http"

	"multi-model-rag-platform/internal/auth"
)

func registerRoutes(mux *http.ServeMux, demoService auth.DemoService) {
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /auth/demo-status", handleDemoStatus(demoService))
	mux.HandleFunc("POST /auth/demo-unlock", handleDemoUnlock(demoService))
	mux.HandleFunc("POST /auth/demo-lock", handleDemoLock(demoService))
}
