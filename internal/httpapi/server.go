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
	return newServerWithDependencies(
		cfg,
		auth.NewDemoService(cfg),
		service.NewChatService(llm.NewRouter(cfg), postgresStore),
		service.NewIngestService(postgresStore),
		service.NewEvalService(postgresStore),
	)
}

func newServerWithDependencies(
	cfg config.Config,
	demoService auth.DemoService,
	chatService service.ChatService,
	ingestService service.IngestService,
	evalService service.EvalService,
) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(
		mux,
		demoService,
		chatService,
		ingestService,
		evalService,
	)

	handler := middleware.CORS(cfg.CORSAllowOrigins, mux)
	handler = middleware.RequestObservability(handler)
	return handler
}
