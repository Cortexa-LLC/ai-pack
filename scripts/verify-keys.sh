#!/bin/bash
set -eo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================================================${NC}"
echo -e "${BLUE}   API Key Verification${NC}"
echo -e "${BLUE}======================================================================${NC}"
echo ""

# Find shell config file
SHELL_CONFIG=""
if [ -f "$HOME/.bash_profile" ]; then
    SHELL_CONFIG="$HOME/.bash_profile"
    echo -e "${BLUE}Found shell config: ~/.bash_profile${NC}"
elif [ -f "$HOME/.bashrc" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
    echo -e "${BLUE}Found shell config: ~/.bashrc${NC}"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
    echo -e "${BLUE}Found shell config: ~/.zshrc${NC}"
fi

echo ""

# Extract keys from shell config if not in environment
if [ -z "${OPENAI_API_KEY:-}" ] && [ -n "$SHELL_CONFIG" ]; then
    export OPENAI_API_KEY=$(grep -E "^export OPENAI_API_KEY=" "$SHELL_CONFIG" 2>/dev/null | sed -E 's/^export OPENAI_API_KEY="?([^"]*)"?$/\1/' | tail -1)
fi

if [ -z "${ANTHROPIC_API_KEY:-}" ] && [ -n "$SHELL_CONFIG" ]; then
    export ANTHROPIC_API_KEY=$(grep -E "^export ANTHROPIC_API_KEY=" "$SHELL_CONFIG" 2>/dev/null | sed -E 's/^export ANTHROPIC_API_KEY="?([^"]*)"?$/\1/' | tail -1)
fi

# Check OpenAI
if [ -z "${OPENAI_API_KEY:-}" ]; then
    echo -e "${RED}✗ OPENAI_API_KEY is not set${NC}"
    echo -e "${YELLOW}  Add to ~/.bash_profile: export OPENAI_API_KEY=\"sk-...\"${NC}"
else
    echo -e "${GREEN}✓ OPENAI_API_KEY is set${NC}"
    echo -e "  Key prefix: ${OPENAI_API_KEY:0:20}..."

    echo -e "${BLUE}  Testing OpenAI API...${NC}"
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /tmp/openai_test.json -X POST https://api.openai.com/v1/chat/completions \
        -H "Authorization: Bearer $OPENAI_API_KEY" \
        -H "Content-Type: application/json" \
        -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"test"}],"max_tokens":5}' 2>/dev/null)

    BODY=$(cat /tmp/openai_test.json 2>/dev/null || echo "")

    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}  ✓ OpenAI API key is VALID!${NC}"
    elif [ "$HTTP_CODE" = "401" ]; then
        echo -e "${RED}  ✗ OpenAI API key is INVALID (401 Unauthorized)${NC}"
        echo -e "${YELLOW}  Check your key at: https://platform.openai.com/api-keys${NC}"
    else
        echo -e "${YELLOW}  ⚠ OpenAI API test returned HTTP $HTTP_CODE${NC}"
        echo -e "${YELLOW}  Response: ${BODY:0:100}${NC}"
    fi
fi

echo ""

# Check Anthropic
if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
    echo -e "${RED}✗ ANTHROPIC_API_KEY is not set${NC}"
    echo -e "${YELLOW}  Add to ~/.bash_profile: export ANTHROPIC_API_KEY=\"sk-ant-...\"${NC}"
else
    echo -e "${GREEN}✓ ANTHROPIC_API_KEY is set${NC}"
    echo -e "  Key prefix: ${ANTHROPIC_API_KEY:0:20}..."

    echo -e "${BLUE}  Testing Anthropic API...${NC}"
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /tmp/anthropic_test.json -X POST https://api.anthropic.com/v1/messages \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -H "Content-Type: application/json" \
        -d '{"model":"claude-sonnet-4-5-20250929","max_tokens":10,"messages":[{"role":"user","content":"test"}]}' 2>/dev/null)

    BODY=$(cat /tmp/anthropic_test.json 2>/dev/null || echo "")

    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}  ✓ Anthropic API key is VALID!${NC}"
    elif [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
        echo -e "${RED}  ✗ Anthropic API key is INVALID (HTTP $HTTP_CODE)${NC}"
        echo -e "${YELLOW}  Check your key at: https://console.anthropic.com/settings/keys${NC}"
    else
        echo -e "${YELLOW}  ⚠ Anthropic API test returned HTTP $HTTP_CODE${NC}"
        echo -e "${YELLOW}  Response: ${BODY:0:100}${NC}"
    fi
fi

echo ""
echo -e "${BLUE}======================================================================${NC}"
echo -e "${YELLOW}Next Steps:${NC}"
echo ""

if [ -n "${OPENAI_API_KEY:-}" ] && [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    echo -e "1. Restart the agent server:"
    echo -e "   ${BLUE}pkill agent-server && ./bin/agent-server --server${NC}"
    echo ""
    echo -e "2. Check server logs for multi-provider initialization:"
    echo -e "   ${BLUE}tail -f /tmp/agent-server.log | grep -E 'openai_client_initialized|provider'${NC}"
else
    echo -e "1. Add missing API keys to ~/.bash_profile"
    echo -e "2. Run this script again to verify"
fi

echo -e "${BLUE}======================================================================${NC}"
