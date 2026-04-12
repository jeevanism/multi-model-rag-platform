package middleware

import (
	"net/http"
	"time"

	"multi-model-rag-platform/internal/observability"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func RequestObservability(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("x-request-id")
		if requestID == "" {
			requestID = observability.NewRequestID()
		}

		ctx := observability.WithRequestID(r.Context(), requestID)
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		observability.LogEvent("request.start", map[string]any{
			"request_id": requestID,
			"method":     r.Method,
			"path":       r.URL.Path,
		})

		defer func() {
			recorder.Header().Set("x-request-id", requestID)
			observability.LogEvent("request.end", map[string]any{
				"request_id":  requestID,
				"method":      r.Method,
				"path":        r.URL.Path,
				"status_code": recorder.statusCode,
				"duration_ms": float64(time.Since(start).Microseconds()) / 1000.0,
			})
		}()

		next.ServeHTTP(recorder, r.WithContext(ctx))
	})
}
