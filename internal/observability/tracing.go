package observability

import (
	"context"
	"time"
)

func Trace(ctx context.Context, span string, fields map[string]any, fn func(context.Context) error) error {
	start := time.Now()
	startFields := cloneFields(fields)
	startFields["span"] = span
	startFields["request_id"] = RequestIDFromContext(ctx)
	LogEvent("span.start", startFields)

	err := fn(ctx)

	endFields := map[string]any{
		"span":        span,
		"request_id":  RequestIDFromContext(ctx),
		"duration_ms": float64(time.Since(start).Microseconds()) / 1000.0,
	}
	LogEvent("span.end", endFields)
	return err
}

func cloneFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}
