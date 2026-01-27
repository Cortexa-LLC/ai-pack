package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/auth"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/proxy"
)

func loadAndPrintConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	fmt.Printf("\n📋 Configuration:\n")
	fmt.Printf("   Mode: %s\n", cfg.API.Mode)
	if cfg.API.Proxy != nil {
		fmt.Printf("   Proxy Base URL: %s\n", cfg.API.Proxy.BaseURL)
	}
	fmt.Printf("   Model: %s\n", cfg.API.AnthropicModel)
	return cfg, nil
}

func getAuthentication() (string, bool, error) {
	fmt.Println("\n🔑 Authentication:")
	apiKey, isBearerToken, err := auth.GetAPIKey()
	if err != nil {
		return "", false, fmt.Errorf("no authentication found: %v", err)
	}

	if isBearerToken {
		fmt.Println("   ✓ Using Bearer token authentication")
	} else {
		fmt.Println("   ✓ Using API key authentication")
	}
	return apiKey, isBearerToken, nil
}

func createHTTPClient(apiKey string, isBearerToken bool, cfg *config.Config) *http.Client {
	fmt.Println("\n🌐 Creating HTTP client with proxy:")

	if isBearerToken {
		fmt.Println("   Mode: Bearer token with proxy transport")
		return proxy.NewBearerTokenClient(apiKey, &cfg.API)
	}

	fmt.Println("   Mode: API key with proxy transport")
	httpClient := proxy.NewHTTPClient(&cfg.API)
	if httpClient == nil {
		return &http.Client{}
	}
	return httpClient
}

func createTestRequest(cfg *config.Config, apiKey string, isBearerToken bool) (*http.Request, error) {
	requestBody := map[string]interface{}{
		"model":      cfg.API.AnthropicModel,
		"max_tokens": 100,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]string{
					{
						"type": "text",
						"text": "Say hello and confirm you received this test message through the proxy. Keep it brief.",
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	apiURL := "https://api.anthropic.com/v1/messages"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	if !isBearerToken {
		req.Header.Set("x-api-key", apiKey)
	}

	return req, nil
}

func executeRequest(httpClient *http.Client, req *http.Request) ([]byte, error) {
	fmt.Println("\n📤 Sending test message to Claude...")
	fmt.Println("   Prompt: 'Say hello and confirm proxy is working'")
	fmt.Println()

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("\n🔍 Debugging tips:")
		fmt.Println("   1. Check that the proxy URL is correct in ~/.claude/agent-server.json")
		fmt.Println("   2. Verify your authentication token is valid")
		fmt.Println("   3. Check network connectivity to proxy")
		fmt.Println("   4. Review the debug output above for URL rewriting details")
		return nil, fmt.Errorf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned error status: %d\n   Response: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func printResponse(responseBody []byte) error {
	var message map[string]interface{}
	if err := json.Unmarshal(responseBody, &message); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Println("\n✅ Success! Response received:")
	if model, ok := message["model"].(string); ok {
		fmt.Printf("   Model: %s\n", model)
	}
	if role, ok := message["role"].(string); ok {
		fmt.Printf("   Role: %s\n", role)
	}
	if stopReason, ok := message["stop_reason"].(string); ok {
		fmt.Printf("   Stop Reason: %s\n", stopReason)
	}

	// Extract text content
	if content, ok := message["content"].([]interface{}); ok && len(content) > 0 {
		if contentBlock, ok := content[0].(map[string]interface{}); ok {
			if text, ok := contentBlock["text"].(string); ok {
				fmt.Printf("\n   Response: %s\n", text)
			}
		}
	}

	// Print usage stats
	if usage, ok := message["usage"].(map[string]interface{}); ok {
		fmt.Printf("\n📊 Token Usage:\n")
		if inputTokens, ok := usage["input_tokens"].(float64); ok {
			fmt.Printf("   Input: %.0f\n", inputTokens)
		}
		if outputTokens, ok := usage["output_tokens"].(float64); ok {
			fmt.Printf("   Output: %.0f\n", outputTokens)
		}
	}

	fmt.Println("\n✅ Proxy configuration is working correctly!")
	return nil
}

func main() {
	fmt.Println("🧪 Testing A2A Agent Server Proxy Configuration")
	fmt.Println("================================================")

	cfg, err := loadAndPrintConfig()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	apiKey, isBearerToken, err := getAuthentication()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("   Please set ANTHROPIC_API_KEY or ANTHROPIC_API_TOKEN")
		os.Exit(1)
	}

	httpClient := createHTTPClient(apiKey, isBearerToken, cfg)

	req, err := createTestRequest(cfg, apiKey, isBearerToken)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	responseBody, err := executeRequest(httpClient, req)
	if err != nil {
		fmt.Printf("\n❌ %v\n", err)
		os.Exit(1)
	}

	if err := printResponse(responseBody); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
}
