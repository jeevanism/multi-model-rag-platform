package llm

import (
	"fmt"
	"strings"
	"time"

	"multi-model-rag-platform/internal/config"
)

type Provider interface {
	Generate(prompt string) (Response, error)
}

type UnsupportedProviderError struct {
	Provider string
}

func (e UnsupportedProviderError) Error() string {
	return fmt.Sprintf(
		"Unsupported provider '%s'. Supported values: gemini, openai, grok.",
		e.Provider,
	)
}

type Router struct {
	config config.Config
}

func NewRouter(cfg config.Config) Router {
	return Router{config: cfg}
}

func (r Router) GetProvider(provider string, model *string, forceStub bool) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "gemini":
		return newStubProvider("gemini", defaultModel(model, "gemini-2.5-flash"), forceStub), nil
	case "openai":
		return newStubProvider("openai", defaultModel(model, "gpt-4.1-mini"), forceStub), nil
	case "grok":
		return newStubProvider("grok", defaultModel(model, "grok-3-mini"), forceStub), nil
	default:
		return nil, UnsupportedProviderError{Provider: provider}
	}
}

func defaultModel(model *string, fallback string) string {
	if model == nil || *model == "" {
		return fallback
	}
	return *model
}

type stubProvider struct {
	provider string
	model    string
}

func newStubProvider(provider string, model string, forceStub bool) Provider {
	return stubProvider{
		provider: provider,
		model:    model,
	}
}

func (p stubProvider) Generate(prompt string) (Response, error) {
	start := time.Now()
	return Response{
		Answer:    fmt.Sprintf("[stub:%s] %s", p.provider, prompt),
		Provider:  p.provider,
		Model:     p.model,
		LatencyMS: int(time.Since(start).Milliseconds()),
	}, nil
}
