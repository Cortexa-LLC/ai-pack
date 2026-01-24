package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/config"
)

// ProxyTransport wraps http.RoundTripper to rewrite URLs for corporate proxies
type ProxyTransport struct {
	Transport http.RoundTripper
	ProxyType string // "acme", "custom"
	BaseURL   string // Full proxy base URL
}

// RoundTrip implements http.RoundTripper
func (t *ProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// DEBUG: Uncomment for troubleshooting
	// fmt.Printf("🔍 ProxyTransport called: %s %s\n", newReq.Method, newReq.URL.String())

	// Rewrite URL based on proxy type
	switch t.ProxyType {
	case "acme":
		// corporate proxy expects: https://api.acme.com/api/v1/anthropic/messages
		// SDK sends to: https://api.anthropic.com/v1/messages
		// We need to replace the base and insert /anthropic before /messages

		if strings.Contains(newReq.URL.Path, "/v1/messages") {
			// Build the correct corporate proxy URL
			// BaseURL = https://api.acme.com/api/v1/anthropic
			// SDK Path = /v1/messages
			// Result: https://api.acme.com/api/v1/anthropic/v1/messages

			newReq.URL.Scheme = "https"
			newReq.URL.Host = extractHost(t.BaseURL)

			// Extract path from base URL (e.g., /api/v1/anthropic)
			basePath := extractPath(t.BaseURL)

			// Append the SDK's path (/v1/messages) as-is
			newReq.URL.Path = basePath + newReq.URL.Path

			// corporate proxy expects Bearer token in Authorization header
			// Get the token from X-Api-Key header (set by Anthropic SDK)
			token := newReq.Header.Get("X-Api-Key")

			if token != "" {
				// Replace Anthropic's x-api-key with Bearer token format
				newReq.Header.Del("X-Api-Key")
				newReq.Header.Set("Authorization", "Bearer "+token)
			}

			// Ensure anthropic-version header is set
			if newReq.Header.Get("anthropic-version") == "" && newReq.Header.Get("Anthropic-Version") == "" {
				newReq.Header.Set("anthropic-version", "2023-06-01")
			}

			// DEBUG: Uncomment for troubleshooting
			// fmt.Printf("🔄 Rewrote URL to: %s\n", newReq.URL.String())
		}

	case "custom":
		// For custom proxies, just replace the host and scheme
		newReq.URL.Scheme = "https"
		newReq.URL.Host = extractHost(t.BaseURL)
		// Keep the path as-is
	}

	// Execute the modified request
	resp, err := t.Transport.RoundTrip(newReq)

	// DEBUG: Uncomment for troubleshooting
	// if err != nil {
	// 	fmt.Printf("❌ Proxy request failed: %v\n", err)
	// } else {
	// 	fmt.Printf("✅ Proxy response: %d %s\n", resp.StatusCode, resp.Status)
	// }

	return resp, err
}

// extractHost extracts hostname from URL
func extractHost(urlStr string) string {
	// Remove scheme
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Get host (everything before first /)
	if idx := strings.Index(urlStr, "/"); idx >= 0 {
		return urlStr[:idx]
	}
	return urlStr
}

// extractPath extracts path from URL
func extractPath(urlStr string) string {
	// Remove scheme
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Get path (everything after first /)
	if idx := strings.Index(urlStr, "/"); idx >= 0 {
		return urlStr[idx:]
	}
	return ""
}

// NewHTTPClient creates an HTTP client with proxy support
func NewHTTPClient(cfg *config.APIConfig) *http.Client {
	if cfg.Mode == "proxy" && cfg.Proxy != nil {
		client := &http.Client{
			Transport: &ProxyTransport{
				Transport: http.DefaultTransport,
				ProxyType: cfg.Proxy.Type,
				BaseURL:   cfg.Proxy.BaseURL,
			},
		}
		return client
	}

	// Direct mode - return nil so SDK uses its default
	return nil
}

// GetBaseURL returns the base URL for the Anthropic SDK
func GetBaseURL(cfg *config.APIConfig) string {
	if cfg.Mode == "proxy" && cfg.Proxy != nil {
		// For proxy mode, we still need to provide a base URL
		// But our transport will rewrite it anyway
		// Return empty to use SDK default, transport will fix it
		return ""
	}

	// Direct mode - no custom base URL
	return ""
}

// LogProxyMode logs the current proxy configuration
func LogProxyMode(cfg *config.APIConfig) string {
	if cfg.Mode == "proxy" && cfg.Proxy != nil {
		return fmt.Sprintf("Proxy mode: %s (%s)", cfg.Proxy.Type, cfg.Proxy.BaseURL)
	}
	return "Direct mode"
}
