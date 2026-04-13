package httpapi

import (
	"net/http"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/config"
	"multi-model-rag-platform/internal/llm"
	"multi-model-rag-platform/internal/middleware"
	"multi-model-rag-platform/internal/service"
	"multi-model-rag-platform/internal/store"
)

func NewServer(cfg config.Config) http.Handler {
	postgresStore, err := store.OpenPostgres(cfg)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	registerRoutes(
		mux,
		auth.NewDemoService(cfg),
		service.NewChatService(llm.NewRouter(cfg)),
		service.NewIngestService(postgresStore),
	)

	handler := middleware.CORS(cfg.CORSAllowOrigins, mux)
	handler = middleware.RequestObservability(handler)
	return handler
}
