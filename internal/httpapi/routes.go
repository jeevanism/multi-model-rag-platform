package httpapi

import "net/http"

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("GET /health", handleHealth)
}
