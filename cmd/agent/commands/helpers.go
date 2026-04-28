package commands

import (
	agentclient "github.com/cortexa-llc/ai-pack/cmd/agent/client"
)

// getServerURL returns the agent server base URL
// Uses DefaultBaseURL which can be overridden by tests
func getServerURL() string {
	return agentclient.DefaultBaseURL
}
