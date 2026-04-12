package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return value
}

func NewRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(bytes[:])
}
