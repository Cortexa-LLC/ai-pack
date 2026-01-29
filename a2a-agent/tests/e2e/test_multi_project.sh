#!/bin/bash
# Test multi-project Beads support

set -e

echo "============================================"
echo "Testing Multi-Project Beads Support"
echo "============================================"
echo ""

# Test 1: Verify we're in the right directory
echo "Test 1: Checking current directory..."
pwd
echo "✅ Working directory: $(pwd)"
echo ""

# Test 2: Verify .beads exists
echo "Test 2: Checking for .beads directory..."
if [ -d ".beads" ]; then
    echo "✅ .beads directory found"
else
    echo "❌ .beads directory not found"
    exit 1
fi
echo ""

# Test 3: Verify Beads task exists
echo "Test 3: Checking if Beads task xasm++-asp exists..."
if bd show xasm++-asp > /dev/null 2>&1; then
    echo "✅ Beads task xasm++-asp exists"
    bd show xasm++-asp | head -3
else
    echo "❌ Beads task xasm++-asp not found"
    exit 1
fi
echo ""

# Test 4: Check agent server is running
echo "Test 4: Checking if agent-server is running..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Agent server is running"
else
    echo "❌ Agent server is not running"
    exit 1
fi
echo ""

# Test 5: Check agent CLI can detect project root
echo "Test 5: Testing project root detection..."
PROJECT_ROOT=$(git rev-parse --show-superproject-working-tree || git rev-parse --show-toplevel)
echo "✅ Detected project root: $PROJECT_ROOT"
echo ""

# Test 6: Verify server working directory
echo "Test 6: Checking server working directory..."
SERVER_PID=$(pgrep -f "agent-server --server" | head -1)
if [ -n "$SERVER_PID" ]; then
    SERVER_CWD=$(lsof -p $SERVER_PID | grep cwd | awk '{print $NF}')
    echo "   Server PID: $SERVER_PID"
    echo "   Server CWD: $SERVER_CWD"
    echo "   Project root: $PROJECT_ROOT"

    if [ "$SERVER_CWD" != "$PROJECT_ROOT" ]; then
        echo "✅ Server is in different directory (multi-project mode enabled)"
    else
        echo "ℹ️  Server is in same directory"
    fi
else
    echo "⚠️  Could not find server PID"
fi
echo ""

echo "============================================"
echo "Ready to test agent command!"
echo "============================================"
echo ""
echo "Run the following command to test:"
echo "  agent engineer xasm++-asp --stream"
echo ""
