package httpapi

import (
	"net/http"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/config"
	"multi-model-rag-platform/internal/llm"
	"multi-model-rag-platform/internal/middleware"
	"multi-model-rag-platform/internal/service"
)

func NewServer(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(
		mux,
		auth.NewDemoService(cfg),
		service.NewChatService(llm.NewRouter(cfg)),
	)

	handler := middleware.CORS(cfg.CORSAllowOrigins, mux)
	handler = middleware.RequestObservability(handler)
	return handler
}
