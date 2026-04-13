package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"multi-model-rag-platform/internal/auth"
	"multi-model-rag-platform/internal/config"
	"multi-model-rag-platform/internal/llm"
	"multi-model-rag-platform/internal/rag"
	"multi-model-rag-platform/internal/service"
)

func TestRootReturnsExpectedPayload(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body rootResponse
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Service != "Multi-Model RAG API" {
		t.Fatalf("expected service name, got %q", body.Service)
	}
	if body.Status != "ok" {
		t.Fatalf("expected status ok, got %q", body.Status)
	}
	if len(body.Endpoints) == 0 || body.Endpoints[0] != "/health" {
		t.Fatalf("expected endpoints to include /health, got %#v", body.Endpoints)
	}
	if body.RequestID == "" {
		t.Fatal("expected request_id to be populated")
	}
	if res.Header().Get("x-request-id") == "" {
		t.Fatal("expected x-request-id header to be populated")
	}
}

func TestHealthReturnsExpectedPayload(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("x-request-id", "req-123")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body healthResponse
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected status ok, got %q", body.Status)
	}
	if body.Service != "api" {
		t.Fatalf("expected service api, got %q", body.Service)
	}
	if body.RequestID != "req-123" {
		t.Fatalf("expected request_id req-123, got %q", body.RequestID)
	}
	if got := res.Header().Get("x-request-id"); got != "req-123" {
		t.Fatalf("expected x-request-id header req-123, got %q", got)
	}
}

func TestOptionsPreflightReturnsNoContent(t *testing.T) {
	server := NewServer(config.Config{
		CORSAllowOrigins: []string{"http://localhost:5173"},
		DatabaseURL:      "postgresql://postgres:postgres@localhost:5432/multimodel_rag",
	})
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header, got %q", got)
	}
}

func TestDemoStatusReportsLockedByDefault(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(http.MethodGet, "/auth/demo-status", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body DemoUnlockStatusResponse
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Unlocked {
		t.Fatal("expected unlocked false")
	}
}

func TestDemoUnlockReturnsUnauthorizedForInvalidPassword(t *testing.T) {
	server := NewServer(config.Config{
		DemoRealModePassword: "secret",
		DemoUnlockCookieName: "mmrag_demo_unlock",
		DatabaseURL:          "postgresql://postgres:postgres@localhost:5432/multimodel_rag",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/demo-unlock", strings.NewReader(`{"password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestDemoUnlockSetsCookieOnSuccess(t *testing.T) {
	server := NewServer(config.Config{
		DemoRealModePassword: "secret",
		DemoUnlockCookieName: "mmrag_demo_unlock",
		DatabaseURL:          "postgresql://postgres:postgres@localhost:5432/multimodel_rag",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/demo-unlock", strings.NewReader(`{"password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if got := res.Header().Get("Set-Cookie"); got == "" {
		t.Fatal("expected unlock cookie to be set")
	}
}

func TestDemoLockClearsCookie(t *testing.T) {
	server := NewServer(config.Config{
		DemoRealModePassword: "secret",
		DemoUnlockCookieName: "mmrag_demo_unlock",
		DatabaseURL:          "postgresql://postgres:postgres@localhost:5432/multimodel_rag",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/demo-lock", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if got := res.Header().Get("Set-Cookie"); !strings.Contains(got, "Max-Age=0") && !strings.Contains(got, "Max-Age=-1") {
		t.Fatalf("expected clearing cookie, got %q", got)
	}
}

func TestChatReturnsRequiredFieldsForStubProvider(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(
		http.MethodPost,
		"/chat",
		strings.NewReader(`{"message":"hello","provider":"gemini"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body ChatResponse
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Answer != "[stub:gemini] hello" {
		t.Fatalf("unexpected answer %q", body.Answer)
	}
	if body.Provider != "gemini" {
		t.Fatalf("expected provider gemini, got %q", body.Provider)
	}
	if body.Model != "gemini-2.5-flash" {
		t.Fatalf("expected default model, got %q", body.Model)
	}
	if body.Citations == nil {
		t.Fatal("expected citations slice to be present")
	}
	if body.RAGUsed {
		t.Fatal("expected rag_used false")
	}
}

func TestChatReturnsBadRequestForUnsupportedProvider(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(
		http.MethodPost,
		"/chat",
		strings.NewReader(`{"message":"hello","provider":"anthropic"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "Unsupported provider") {
		t.Fatalf("expected unsupported provider error, got %q", res.Body.String())
	}
}

func TestChatValidatesMissingMessage(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(
		http.MethodPost,
		"/chat",
		strings.NewReader(`{"provider":"gemini"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}

func TestChatRAGReturnsCitationsAndDebugChunks(t *testing.T) {
	cfg := config.Config{
		DatabaseURL: "postgresql://postgres:postgres@localhost:5432/multimodel_rag",
	}
	server := newServerWithDependencies(
		cfg,
		auth.NewDemoService(cfg),
		service.NewChatService(llm.NewRouter(cfg), fakeHTTPChatRepo{
			chunks: []rag.RetrievedChunk{
				{
					DocumentID: 1,
					ChunkID:    10,
					ChunkIndex: 0,
					Title:      "Test Doc",
					Content:    "Important context.",
					Score:      0.123,
				},
			},
		}),
		service.NewIngestService(nil),
	)
	req := httptest.NewRequest(
		http.MethodPost,
		"/chat",
		strings.NewReader(`{"message":"what is in the doc?","provider":"gemini","rag":true,"top_k":2,"debug":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body ChatResponse
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.RAGUsed {
		t.Fatal("expected rag_used true")
	}
	if len(body.Citations) != 1 || body.Citations[0] != "[source:Test Doc#chunk=0]" {
		t.Fatalf("unexpected citations %#v", body.Citations)
	}
	if len(body.RetrievedChunks) != 1 || body.RetrievedChunks[0].Title != "Test Doc" {
		t.Fatalf("unexpected retrieved chunks %#v", body.RetrievedChunks)
	}
	if !strings.Contains(body.Answer, "[source:Test Doc#chunk=0]") {
		t.Fatalf("expected citation in answer, got %q", body.Answer)
	}
}

func TestChatStreamReturnsSSEEvents(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/stream",
		strings.NewReader(`{"message":"hello world","provider":"gemini"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("expected text/event-stream content type, got %q", got)
	}
	body := res.Body.String()
	if !strings.Contains(body, "event: start") {
		t.Fatalf("expected start event, got %q", body)
	}
	if !strings.Contains(body, "event: token") {
		t.Fatalf("expected token event, got %q", body)
	}
	if !strings.Contains(body, `"text":"[stub:gemini]"`) && !strings.Contains(body, `"text": "[stub:gemini]"`) {
		t.Fatalf("expected first token payload, got %q", body)
	}
	if !strings.Contains(body, `"text":"hello"`) && !strings.Contains(body, `"text": "hello"`) {
		t.Fatalf("expected hello token payload, got %q", body)
	}
	if !strings.Contains(body, `"text":"world"`) && !strings.Contains(body, `"text": "world"`) {
		t.Fatalf("expected world token payload, got %q", body)
	}
	if !strings.Contains(body, "event: end") {
		t.Fatalf("expected end event, got %q", body)
	}
}

func TestChatStreamReturnsBadRequestForUnsupportedProvider(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/stream",
		strings.NewReader(`{"message":"hello","provider":"anthropic"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}

type fakeHTTPChatRepo struct {
	chunks []rag.RetrievedChunk
}

func (f fakeHTTPChatRepo) RetrieveChunks(query string, topK int) ([]rag.RetrievedChunk, error) {
	return f.chunks, nil
}

func TestIngestTextValidatesRequiredFields(t *testing.T) {
	server := NewServer(config.Load())
	req := httptest.NewRequest(
		http.MethodPost,
		"/ingest/text",
		strings.NewReader(`{"title":"Doc"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}
