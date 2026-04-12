package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"multi-model-rag-platform/internal/config"
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
