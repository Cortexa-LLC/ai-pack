#!/bin/bash
# AI-Pack Agent Server Startup Script
# A2A Protocol + Streaming + Parallel Execution

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                                  ║${NC}"
echo -e "${BLUE}║                     AI-Pack Agent Server                         ║${NC}"
echo -e "${BLUE}║                                                                  ║${NC}"
echo -e "${BLUE}║    Features: A2A Protocol + SSE Streaming + Parallel Execution  ║${NC}"
echo -e "${BLUE}║                                                                  ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed${NC}"
    echo ""
    echo "Please install Go 1.21+ first:"
    echo ""
    echo "  macOS:   brew install go"
    echo "  Linux:   https://go.dev/dl/"
    echo "  Windows: https://go.dev/dl/"
    echo ""
    echo "Then run this script again."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}✅ Go ${GO_VERSION} installed${NC}"

# Check API key (either env var or Claude Code login)
if [ -z "$ANTHROPIC_API_KEY" ]; then
    # Check if Claude Code is configured
    if [ -f "$HOME/.claude/settings.json" ] && grep -q "apiKeyHelper" "$HOME/.claude/settings.json"; then
        echo -e "${GREEN}✅ Using Claude Code authentication${NC}"
    else
        echo -e "${YELLOW}⚠️  Warning: ANTHROPIC_API_KEY not set and Claude Code not configured${NC}"
        echo ""
        echo "Option 1 - Set API key manually:"
        echo "  export ANTHROPIC_API_KEY=\"your-key-here\""
        echo ""
        echo "Option 2 - Use Claude Code login (if you're already logged in):"
        echo "  claude login"
        echo ""
        exit 1
    fi
else
    echo -e "${GREEN}✅ ANTHROPIC_API_KEY configured${NC}"
fi
echo ""

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy
echo -e "${GREEN}✅ Dependencies installed${NC}"
echo ""

# Display features
echo -e "${BLUE}🎯 Features:${NC}"
echo "   - A2A Protocol Compliance (JSON-RPC 2.0)  ✅"
echo "   - SSE Streaming (Real-time progress)       ✅"
echo "   - Parallel Execution (configurable)        ✅"
echo "   - Structured Logging & Metrics             ✅"
echo ""

# Start server
echo -e "${GREEN}🔥 Starting AI-Pack Agent Server...${NC}"
echo ""

go run cmd/agent-server/main.go
