package httpapi

import (
	"net/http"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/config"
	"multi-model-rag-platform/internal/middleware"
)

func NewServer(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, auth.NewDemoService(cfg))

	handler := middleware.CORS(cfg.CORSAllowOrigins, mux)
	handler = middleware.RequestObservability(handler)
	return handler
}
