#!/bin/bash
# Test Engineer Skill with Code-Monkey Integration

echo "==================================="
echo "Engineer Skill Integration Test"
echo "==================================="
echo ""

echo "1. Checking engineer role updates..."
if grep -q "code-monkey agent" /Users/bryanw/Projects/Claude/ai-pack/roles/engineer.md; then
    echo "   ✅ Engineer role mentions code-monkey agent"
else
    echo "   ❌ Engineer role missing code-monkey reference"
    exit 1
fi

if grep -q "Background File Operations" /Users/bryanw/Projects/Claude/ai-pack/roles/engineer.md; then
    echo "   ✅ Background operations section exists"
else
    echo "   ❌ Background operations section missing"
    exit 1
fi

if grep -q "Bug #13890" /Users/bryanw/Projects/Claude/ai-pack/roles/engineer.md; then
    echo "   ✅ Bug workaround documented"
else
    echo "   ❌ Bug workaround not documented"
    exit 1
fi

echo ""
echo "2. Checking engineer command updates..."
if grep -q "background agents" /Users/bryanw/Projects/Claude/ai-pack/templates/.claude/commands/ai-pack/engineer.md; then
    echo "   ✅ Engineer command mentions background agents"
else
    echo "   ❌ Engineer command missing background agent reference"
    exit 1
fi

echo ""
echo "3. Checking code-monkey agent configuration..."
if [ -f /Users/bryanw/Projects/Claude/ai-pack/.claude/agents/code-monkey.md ]; then
    echo "   ✅ Code-monkey agent exists"
    
    if grep -q "Bash tool" /Users/bryanw/Projects/Claude/ai-pack/.claude/agents/code-monkey.md; then
        echo "   ✅ Bash workaround instructions present"
    else
        echo "   ❌ Bash workaround instructions missing"
        exit 1
    fi
else
    echo "   ❌ Code-monkey agent file not found"
    exit 1
fi

echo ""
echo "==================================="
echo "✅ All integration tests passed!"
echo "==================================="
echo ""
echo "Engineer skill is ready to use code-monkey agent for background operations."
echo ""
echo "Usage example:"
echo "  /ai-pack engineer"
echo "  # When task requires multiple file creation:"
echo "  # Use Task tool with code-monkey agent in background"
echo ""
