package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/server"
)

// stub handler that always returns 200
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// ---------------------------------------------------------------------------
// CORSMiddleware
// ---------------------------------------------------------------------------

func TestCORSMiddleware_NoOrigin(t *testing.T) {
	h := server.CORSMiddleware(okHandler)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestCORSMiddleware_LocalhostOriginAllowed(t *testing.T) {
	origins := []string{
		"http://localhost:3000",
		"https://localhost",
		"http://127.0.0.1:8080",
		"https://127.0.0.1",
		"http://[::1]:5173",
		"https://[::1]",
	}
	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			h := server.CORSMiddleware(okHandler)
			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header.Set("Origin", origin)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d for origin %q", rr.Code, origin)
			}
			if got := rr.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("expected ACAO=%q, got %q", origin, got)
			}
		})
	}
}

func TestCORSMiddleware_ExternalOriginForbidden(t *testing.T) {
	origins := []string{
		"http://evil.com",
		"https://attacker.example",
		"http://192.168.1.1",
	}
	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			h := server.CORSMiddleware(okHandler)
			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header.Set("Origin", origin)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d for origin %q", rr.Code, origin)
			}
		})
	}
}

func TestCORSMiddleware_PreflightLocalhost(t *testing.T) {
	h := server.CORSMiddleware(okHandler)
	req := httptest.NewRequest(http.MethodOptions, "/foo", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("unexpected ACAO: %q", got)
	}
}

func TestCORSMiddleware_PreflightExternalOrigin(t *testing.T) {
	h := server.CORSMiddleware(okHandler)
	req := httptest.NewRequest(http.MethodOptions, "/foo", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 (no ACAO header), got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no ACAO header for disallowed origin, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// APIKeyMiddleware
// ---------------------------------------------------------------------------

func TestAPIKeyMiddleware_NoKeyConfigured(t *testing.T) {
	// When apiKey is empty the middleware is a no-op
	h := server.APIKeyMiddleware("", okHandler)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_HealthExempt(t *testing.T) {
	h := server.APIKeyMiddleware("secret", okHandler)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// No key provided
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("/health should be exempt, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_MissingKey(t *testing.T) {
	h := server.APIKeyMiddleware("secret", okHandler)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_WrongKey(t *testing.T) {
	h := server.APIKeyMiddleware("secret", okHandler)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("X-Api-Key", "wrong")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_CorrectXApiKey(t *testing.T) {
	h := server.APIKeyMiddleware("secret", okHandler)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("X-Api-Key", "secret")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_BearerToken(t *testing.T) {
	h := server.APIKeyMiddleware("secret", okHandler)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
