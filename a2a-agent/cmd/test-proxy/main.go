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

func main() {
	fmt.Println("🧪 Testing A2A Agent Server Proxy Configuration")
	fmt.Println("================================================")

	// Load configuration from ~/.claude/agent-server.json
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📋 Configuration:\n")
	fmt.Printf("   Mode: %s\n", cfg.API.Mode)
	if cfg.API.Proxy != nil {
		fmt.Printf("   Proxy Base URL: %s\n", cfg.API.Proxy.BaseURL)
	}
	fmt.Printf("   Model: %s\n", cfg.API.AnthropicModel)

	// Get API key or bearer token
	fmt.Println("\n🔑 Authentication:")
	apiKey, isBearerToken, err := auth.GetAPIKey()

	if err != nil {
		fmt.Printf("❌ No authentication found: %v\n", err)
		fmt.Println("   Please set ANTHROPIC_API_KEY or ANTHROPIC_API_TOKEN")
		os.Exit(1)
	}

	if isBearerToken {
		fmt.Println("   ✓ Using Bearer token authentication")
	} else {
		fmt.Println("   ✓ Using API key authentication")
	}

	// Create HTTP client with proxy transport
	fmt.Println("\n🌐 Creating HTTP client with proxy:")
	var httpClient *http.Client

	if isBearerToken {
		// Bearer token mode
		fmt.Println("   Mode: Bearer token with proxy transport")
		httpClient = proxy.NewBearerTokenClient(apiKey, &cfg.API)
	} else {
		// API key mode
		fmt.Println("   Mode: API key with proxy transport")
		httpClient = proxy.NewHTTPClient(&cfg.API)
		if httpClient == nil {
			httpClient = &http.Client{}
		}
	}

	// Prepare API request
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
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP request
	apiURL := "https://api.anthropic.com/v1/messages"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	if !isBearerToken {
		req.Header.Set("x-api-key", apiKey)
	}

	// Test simple message
	fmt.Println("\n📤 Sending test message to Claude...")
	fmt.Println("   Prompt: 'Say hello and confirm proxy is working'")
	fmt.Println()

	// Execute request (proxy transport will rewrite the URL)
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("\n❌ API call failed: %v\n", err)
		fmt.Println("\n🔍 Debugging tips:")
		fmt.Println("   1. Check that the proxy URL is correct in ~/.claude/agent-server.json")
		fmt.Println("   2. Verify your authentication token is valid")
		fmt.Println("   3. Check network connectivity to proxy")
		fmt.Println("   4. Review the debug output above for URL rewriting details")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Check status code
	if resp.StatusCode != 200 {
		fmt.Printf("❌ API returned error status: %d\n", resp.StatusCode)
		fmt.Printf("   Response: %s\n", string(responseBody))
		os.Exit(1)
	}

	// Parse response
	var message map[string]interface{}
	if err := json.Unmarshal(responseBody, &message); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	// Print response
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

	// Pretty print full response for debugging
	fmt.Println("\n📄 Full Response (JSON):")
	responseJSON, _ := json.MarshalIndent(message, "   ", "  ")
	fmt.Printf("   %s\n", string(responseJSON))

	fmt.Println("\n✅ Proxy test completed successfully!")
	fmt.Println("\nThe A2A agent server proxy configuration is working correctly.")
}
