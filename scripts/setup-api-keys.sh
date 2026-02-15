#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   AI-Pack Agent Server - API Key Setup${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Detect shell config file
detect_shell_config() {
    if [ -n "${ZSH_VERSION:-}" ]; then
        echo "$HOME/.zshrc"
    elif [ -n "${BASH_VERSION:-}" ]; then
        if [ -f "$HOME/.bash_profile" ]; then
            echo "$HOME/.bash_profile"
        else
            echo "$HOME/.bashrc"
        fi
    else
        echo "$HOME/.profile"
    fi
}

SHELL_CONFIG=$(detect_shell_config)

echo -e "${BLUE}Shell config:${NC} $SHELL_CONFIG"
echo ""

# Check existing keys
check_existing_keys() {
    local openai_set=false
    local anthropic_set=false

    if [ -n "${OPENAI_API_KEY:-}" ]; then
        openai_set=true
        echo -e "${GREEN}✓${NC} OPENAI_API_KEY is already set"
    else
        echo -e "${YELLOW}⚠${NC} OPENAI_API_KEY is not set"
    fi

    if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
        anthropic_set=true
        echo -e "${GREEN}✓${NC} ANTHROPIC_API_KEY is already set"
    else
        echo -e "${YELLOW}⚠${NC} ANTHROPIC_API_KEY is not set"
    fi

    echo ""

    if $openai_set && $anthropic_set; then
        echo -e "${GREEN}✓ Both API keys are configured!${NC}"
        echo ""
        test_api_keys
        return 0
    fi

    return 1
}

# Test API keys
test_api_keys() {
    echo -e "${BLUE}Testing API keys...${NC}"
    echo ""

    # Test OpenAI
    if [ -n "${OPENAI_API_KEY:-}" ]; then
        echo -e "${BLUE}Testing OpenAI API...${NC}"
        if curl -s -X POST https://api.openai.com/v1/chat/completions \
            -H "Authorization: Bearer $OPENAI_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"test"}],"max_tokens":5}' \
            | grep -q "choices"; then
            echo -e "${GREEN}✓ OpenAI API key is valid${NC}"
        else
            echo -e "${RED}✗ OpenAI API key test failed${NC}"
        fi
        echo ""
    fi

    # Test Anthropic
    if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
        echo -e "${BLUE}Testing Anthropic API...${NC}"
        if curl -s -X POST https://api.anthropic.com/v1/messages \
            -H "x-api-key: $ANTHROPIC_API_KEY" \
            -H "anthropic-version: 2023-06-01" \
            -H "Content-Type: application/json" \
            -d '{"model":"claude-3-5-haiku-20241022","max_tokens":10,"messages":[{"role":"user","content":"test"}]}' \
            | grep -q "content"; then
            echo -e "${GREEN}✓ Anthropic API key is valid${NC}"
        else
            echo -e "${RED}✗ Anthropic API key test failed${NC}"
        fi
        echo ""
    fi
}

# Setup OpenAI
setup_openai() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}OpenAI API Key Setup${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Get your API key from: https://platform.openai.com/api-keys"
    echo ""
    echo -e "${YELLOW}Models available:${NC}"
    echo "  • gpt-5.2       - \$5.00/\$15.00 per 1M tokens (primary)"
    echo "  • gpt-5.2-mini  - \$0.60/\$2.40 per 1M tokens (bulk)"
    echo "  • gpt-4o-mini   - \$0.15/\$0.60 per 1M tokens (cheapest)"
    echo ""

    read -p "Do you want to set up OpenAI API key? (y/n): " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -sp "Enter your OpenAI API key (sk-...): " OPENAI_KEY
        echo ""

        if [[ ! $OPENAI_KEY =~ ^sk- ]]; then
            echo -e "${RED}✗ Invalid key format (should start with 'sk-')${NC}"
            return 1
        fi

        # Add to shell config
        if grep -q "OPENAI_API_KEY" "$SHELL_CONFIG"; then
            # Update existing
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s|^export OPENAI_API_KEY=.*|export OPENAI_API_KEY=\"$OPENAI_KEY\"|" "$SHELL_CONFIG"
            else
                sed -i "s|^export OPENAI_API_KEY=.*|export OPENAI_API_KEY=\"$OPENAI_KEY\"|" "$SHELL_CONFIG"
            fi
            echo -e "${GREEN}✓ Updated OPENAI_API_KEY in $SHELL_CONFIG${NC}"
        else
            # Add new
            echo "" >> "$SHELL_CONFIG"
            echo "# AI-Pack Agent Server - OpenAI API Key" >> "$SHELL_CONFIG"
            echo "export OPENAI_API_KEY=\"$OPENAI_KEY\"" >> "$SHELL_CONFIG"
            echo -e "${GREEN}✓ Added OPENAI_API_KEY to $SHELL_CONFIG${NC}"
        fi

        # Set for current session
        export OPENAI_API_KEY="$OPENAI_KEY"
        echo ""
    fi
}

# Setup Anthropic
setup_anthropic() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}Anthropic API Key Setup${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Get your API key from: https://console.anthropic.com/settings/keys"
    echo ""
    echo -e "${YELLOW}Models available:${NC}"
    echo "  • claude-sonnet-3-5-20241022 - \$3.00/\$15.00 per 1M tokens"
    echo "  • claude-opus-3-5-20241022   - \$15.00/\$75.00 per 1M tokens"
    echo ""
    echo -e "${YELLOW}Note:${NC} Use Claude for critical tasks only (10% of work)"
    echo ""

    read -p "Do you want to set up Anthropic API key? (y/n): " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -sp "Enter your Anthropic API key (sk-ant-...): " ANTHROPIC_KEY
        echo ""

        if [[ ! $ANTHROPIC_KEY =~ ^sk-ant- ]]; then
            echo -e "${RED}✗ Invalid key format (should start with 'sk-ant-')${NC}"
            return 1
        fi

        # Add to shell config
        if grep -q "ANTHROPIC_API_KEY" "$SHELL_CONFIG"; then
            # Update existing
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s|^export ANTHROPIC_API_KEY=.*|export ANTHROPIC_API_KEY=\"$ANTHROPIC_KEY\"|" "$SHELL_CONFIG"
            else
                sed -i "s|^export ANTHROPIC_API_KEY=.*|export ANTHROPIC_API_KEY=\"$ANTHROPIC_KEY\"|" "$SHELL_CONFIG"
            fi
            echo -e "${GREEN}✓ Updated ANTHROPIC_API_KEY in $SHELL_CONFIG${NC}"
        else
            # Add new
            echo "" >> "$SHELL_CONFIG"
            echo "# AI-Pack Agent Server - Anthropic API Key" >> "$SHELL_CONFIG"
            echo "export ANTHROPIC_API_KEY=\"$ANTHROPIC_KEY\"" >> "$SHELL_CONFIG"
            echo -e "${GREEN}✓ Added ANTHROPIC_API_KEY to $SHELL_CONFIG${NC}"
        fi

        # Set for current session
        export ANTHROPIC_API_KEY="$ANTHROPIC_KEY"
        echo ""
    fi
}

# Main flow
main() {
    echo -e "${BLUE}Checking current configuration...${NC}"
    echo ""

    if check_existing_keys; then
        echo -e "${GREEN}✓ Setup complete!${NC}"
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "1. Restart your agent server: pkill agent-server && ./bin/agent-server --server"
        echo "2. Check logs: tail -f /tmp/agent-server.log"
        echo ""
        exit 0
    fi

    echo -e "${YELLOW}Let's set up your API keys...${NC}"
    echo ""

    # Setup each provider
    if [ -z "${OPENAI_API_KEY:-}" ]; then
        setup_openai
    fi

    if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
        setup_anthropic
    fi

    # Final summary
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}Setup Complete!${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""

    # Test the keys
    test_api_keys

    echo -e "${YELLOW}Next steps:${NC}"
    echo ""
    echo "1. Reload your shell config:"
    echo -e "   ${BLUE}source $SHELL_CONFIG${NC}"
    echo ""
    echo "2. Verify keys are set:"
    echo -e "   ${BLUE}echo \$OPENAI_API_KEY | cut -c1-10${NC}"
    echo -e "   ${BLUE}echo \$ANTHROPIC_API_KEY | cut -c1-10${NC}"
    echo ""
    echo "3. Restart agent server:"
    echo -e "   ${BLUE}pkill agent-server && ./bin/agent-server --server${NC}"
    echo ""
    echo "4. Check logs for 'openai_client_initialized':"
    echo -e "   ${BLUE}tail -f /tmp/agent-server.log | grep -i openai${NC}"
    echo ""

    echo -e "${GREEN}✓ All done!${NC}"
}

# Run main
main
