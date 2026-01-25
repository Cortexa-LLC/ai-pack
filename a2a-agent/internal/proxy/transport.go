package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
)

// ProxyTransport wraps http.RoundTripper to rewrite URLs for corporate proxies
// This only does URL rewriting - auth headers are handled separately
type ProxyTransport struct {
	Transport http.RoundTripper
	BaseURL   string // Full proxy base URL
}

// RoundTrip implements http.RoundTripper
func (t *ProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// DEBUG: Uncomment for troubleshooting
	fmt.Printf("🔍 ProxyTransport called: %s %s\n", newReq.Method, newReq.URL.String())
	fmt.Printf("   Headers: %v\n", newReq.Header)

	// Rewrite URL to use custom base URL
	// Simply replace the host/scheme with the configured base URL
	newReq.URL.Scheme = "https"
	newReq.URL.Host = extractHost(t.BaseURL)

	// If there's a path in the base URL, prepend it
	basePath := extractPath(t.BaseURL)
	if basePath != "" {
		newReq.URL.Path = basePath + newReq.URL.Path
	}

	// DEBUG: Uncomment for troubleshooting
	fmt.Printf("🔄 Rewrote URL to: %s\n", newReq.URL.String())
	fmt.Printf("   Headers: %v\n", newReq.Header)

	// Execute the modified request
	resp, err := t.Transport.RoundTrip(newReq)

	// DEBUG: Uncomment for troubleshooting
	if err != nil {
		fmt.Printf("❌ Proxy request failed: %v\n", err)
	} else {
		fmt.Printf("✅ Proxy response: %d %s\n", resp.StatusCode, resp.Status)
	}

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

// BearerTokenTransport wraps http.RoundTripper to add Bearer token authentication
type BearerTokenTransport struct {
	Transport   http.RoundTripper
	BearerToken string
	BaseURL     string // Optional: for proxy mode
}

// RoundTrip implements http.RoundTripper for Bearer token auth
func (t *BearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// DEBUG: Uncomment for troubleshooting
	fmt.Printf("🔍 BearerTokenTransport called: %s %s\n", newReq.Method, newReq.URL.String())

	// Rewrite URL if using proxy
	if t.BaseURL != "" {
		newReq.URL.Scheme = "https"
		newReq.URL.Host = extractHost(t.BaseURL)

		basePath := extractPath(t.BaseURL)
		if basePath != "" {
			newReq.URL.Path = basePath + newReq.URL.Path
		}
		fmt.Printf("🔄 Rewrote URL to: %s\n", newReq.URL.String())
	}

	// Set Bearer token (replaces any x-api-key header from SDK)
	newReq.Header.Del("x-api-key")
	newReq.Header.Del("X-Api-Key")
	newReq.Header.Set("Authorization", "Bearer "+t.BearerToken)
	fmt.Printf("   ✓ Set Bearer token authorization\n")

	// Ensure anthropic-version header is set
	if newReq.Header.Get("anthropic-version") == "" && newReq.Header.Get("Anthropic-Version") == "" {
		newReq.Header.Set("anthropic-version", "2023-06-01")
	}

	// Execute the request
	resp, err := t.Transport.RoundTrip(newReq)

	// DEBUG
	if err != nil {
		fmt.Printf("❌ Bearer token request failed: %v\n", err)
	} else {
		fmt.Printf("✅ Bearer token response: %d %s\n", resp.StatusCode, resp.Status)
	}

	return resp, err
}

// NewHTTPClient creates an HTTP client with proxy support (API key mode only)
func NewHTTPClient(cfg *config.APIConfig) *http.Client {
	if cfg.Mode == "proxy" && cfg.Proxy != nil {
		client := &http.Client{
			Transport: &ProxyTransport{
				Transport: http.DefaultTransport,
				BaseURL:   cfg.Proxy.BaseURL,
			},
		}
		return client
	}

	// Direct mode - return nil so SDK uses its default
	return nil
}

// NewBearerTokenClient creates an HTTP client with Bearer token authentication
func NewBearerTokenClient(bearerToken string, cfg *config.APIConfig) *http.Client {
	baseURL := ""
	if cfg.Mode == "proxy" && cfg.Proxy != nil {
		baseURL = cfg.Proxy.BaseURL
	}

	return &http.Client{
		Transport: &BearerTokenTransport{
			Transport:   http.DefaultTransport,
			BearerToken: bearerToken,
			BaseURL:     baseURL,
		},
	}
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
		return fmt.Sprintf("Proxy mode (base URL: %s)", cfg.Proxy.BaseURL)
	}
	return "Direct mode"
}
