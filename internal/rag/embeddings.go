package rag

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	EmbeddingDim          = 8
	StubEmbeddingProvider = "stub"
	StubEmbeddingModel    = "stub-embedding-v1"
)

func EmbedText(text string, dimensions int) (EmbeddingResult, error) {
	vector, err := EmbedTextDeterministic(text, dimensions)
	if err != nil {
		return EmbeddingResult{}, err
	}
	return EmbeddingResult{
		Vector:   vector,
		Provider: StubEmbeddingProvider,
		Model:    StubEmbeddingModel,
	}, nil
}

func EmbedTextDeterministic(text string, dimensions int) ([]float64, error) {
	if dimensions <= 0 {
		return nil, errors.New("dimensions must be > 0")
	}

	digest := sha256.Sum256([]byte(text))
	values := make([]float64, 0, dimensions)
	for i := 0; i < dimensions; i++ {
		b := digest[i%len(digest)]
		value := (float64(b) / 127.5) - 1.0
		values = append(values, round6(value))
	}
	return values, nil
}

func ToPGVectorLiteral(vector []float64) string {
	parts := make([]string, 0, len(vector))
	for _, value := range vector {
		parts = append(parts, fmt.Sprintf("%.6f", value))
	}
	return "[" + stringsJoin(parts, ",") + "]"
}

func round6(value float64) float64 {
	return float64(int(value*1_000_000+signOffset(value))) / 1_000_000
}

func signOffset(value float64) float64 {
	if value >= 0 {
		return 0.5
	}
	return -0.5
}

func stringsJoin(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	joined := items[0]
	for _, item := range items[1:] {
		joined += sep + item
	}
	return joined
}
