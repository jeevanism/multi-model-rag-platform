package httpapi

import (
	"net/http"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/service"
)

func registerRoutes(
	mux *http.ServeMux,
	demoService auth.DemoService,
	chatService service.ChatService,
) {
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /chat", handleChat(demoService, chatService))
	mux.HandleFunc("POST /chat/stream", handleChatStream(demoService, chatService))
	mux.HandleFunc("GET /auth/demo-status", handleDemoStatus(demoService))
	mux.HandleFunc("POST /auth/demo-unlock", handleDemoUnlock(demoService))
	mux.HandleFunc("POST /auth/demo-lock", handleDemoLock(demoService))
}
