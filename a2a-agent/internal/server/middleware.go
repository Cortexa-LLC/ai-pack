package server

import (
	"net/http"
	"strings"
)

// CORSMiddleware restricts cross-origin requests to localhost origins only.
// All other origins receive 403. OPTIONS preflight requests are handled directly.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := origin == "" || isLocalhostOrigin(origin)

		if r.Method == http.MethodOptions {
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if !allowed {
			http.Error(w, "forbidden: cross-origin requests not allowed", http.StatusForbidden)
			return
		}

		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		next.ServeHTTP(w, r)
	})
}

func isLocalhostOrigin(origin string) bool {
	origin = strings.ToLower(origin)
	return strings.HasPrefix(origin, "http://localhost") ||
		strings.HasPrefix(origin, "https://localhost") ||
		strings.HasPrefix(origin, "http://127.0.0.1") ||
		strings.HasPrefix(origin, "https://127.0.0.1") ||
		strings.HasPrefix(origin, "http://[::1]") ||
		strings.HasPrefix(origin, "https://[::1]")
}

// APIKeyMiddleware enforces X-Api-Key header authentication when apiKey is non-empty.
// If apiKey is empty the middleware is a no-op (dev/localhost convenience).
// The /health endpoint is always exempt.
func APIKeyMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" || r.URL.Path == "/health" || r.URL.Path == "/.well-known/agent.json" {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-Api-Key")
		if key == "" {
			// Also accept Authorization: Bearer <key>
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if key != apiKey {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
