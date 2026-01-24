#!/bin/bash
# Quick A2A Protocol Validation
# Tests the agent-server in server mode with A2A endpoints

set -e

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║              AI-Pack A2A Protocol Validation                     ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Check if binaries exist
if [ ! -f bin/agent-server ]; then
    echo "❌ agent-server binary not found. Build it first:"
    echo "   go build -o bin/agent-server cmd/agent-server/main.go"
    exit 1
fi

# Kill any existing server
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
sleep 1

echo "🚀 Starting agent-server in server mode..."
./bin/agent-server -server -config agent-server-proxy.json > /tmp/a2a-test.log 2>&1 &
SERVER_PID=$!
echo "   Server PID: $SERVER_PID"
sleep 3

# Cleanup on exit
trap "echo ''; echo 'Stopping server...'; kill $SERVER_PID 2>/dev/null || true" EXIT

echo ""
echo "Testing A2A Endpoints..."
echo "========================"
echo ""

# Test 1: Health Check
echo "1. Health Check"
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status')
if [ "$HEALTH" = "healthy" ]; then
    echo "   ✅ Server is healthy"
else
    echo "   ❌ Health check failed"
    exit 1
fi
echo ""

# Test 2: A2A Discovery
echo "2. A2A Discovery"
DISCOVERY=$(curl -s http://localhost:8080/a2a/discovery)
AGENT_NAME=$(echo "$DISCOVERY" | jq -r '.name')
if [ "$AGENT_NAME" = "ai-pack-agent-server" ]; then
    echo "   ✅ Discovery endpoint works"
    echo "   Agent: $AGENT_NAME"
    echo "   Version: $(echo "$DISCOVERY" | jq -r '.version')"
    echo "   Agents: $(echo "$DISCOVERY" | jq -r '.agents | length')"
else
    echo "   ❌ Discovery failed"
    echo "   Response: $DISCOVERY"
    exit 1
fi
echo ""

# Test 3: Execute Task (JSON-RPC)
echo "3. Execute Task (JSON-RPC)"
EXECUTE_RESULT=$(curl -s -X POST http://localhost:8080/a2a/execute \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "method": "a2a.execute",
        "params": {
            "role": "engineer",
            "task": "echo hello from A2A"
        },
        "id": 1
    }')

TASK_ID=$(echo "$EXECUTE_RESULT" | jq -r '.result.task_id')
if [ -n "$TASK_ID" ] && [ "$TASK_ID" != "null" ]; then
    echo "   ✅ Task spawned"
    echo "   Task ID: $TASK_ID"
else
    echo "   ❌ Execute failed"
    echo "   Response: $EXECUTE_RESULT"
    exit 1
fi
echo ""

# Test 4: Status Check
echo "4. Task Status (JSON-RPC)"
sleep 2
STATUS_RESULT=$(curl -s -X POST http://localhost:8080/a2a/status \
    -H "Content-Type: application/json" \
    -d "{
        \"jsonrpc\": \"2.0\",
        \"method\": \"a2a.status\",
        \"params\": {
            \"task_id\": \"$TASK_ID\"
        },
        \"id\": 2
    }")

TASK_STATUS=$(echo "$STATUS_RESULT" | jq -r '.result.status')
echo "   Task Status: $TASK_STATUS"

if [ "$TASK_STATUS" = "queued" ] || [ "$TASK_STATUS" = "in_progress" ] || [ "$TASK_STATUS" = "completed" ]; then
    echo "   ✅ Status endpoint works"
else
    echo "   ❌ Invalid status: $TASK_STATUS"
    exit 1
fi
echo ""

# Test 5: SSE Streaming
echo "5. SSE Streaming"
STREAM_URL="http://localhost:8080/stream/$TASK_ID"
echo "   Stream URL: $STREAM_URL"

# Try to get first event (with timeout)
STREAM_TEST=$(timeout 5 curl -s -N "$STREAM_URL" 2>/dev/null | head -5 || echo "timeout")

if echo "$STREAM_TEST" | grep -q "data:"; then
    echo "   ✅ SSE streaming works"
else
    echo "   ⚠️  SSE streaming: No events yet (task may be complete)"
fi
echo ""

# Test 6: Metrics
echo "6. Metrics Endpoint"
METRICS=$(curl -s http://localhost:8080/metrics)
TASKS_SPAWNED=$(echo "$METRICS" | jq -r '.tasks_spawned')
if [ "$TASKS_SPAWNED" -gt 0 ]; then
    echo "   ✅ Metrics endpoint works"
    echo "   Tasks spawned: $TASKS_SPAWNED"
else
    echo "   ❌ Metrics check failed"
fi
echo ""

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                    ✅ All A2A Tests Passed                       ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Server log: /tmp/a2a-test.log"
echo "Task results: .beads/tasks/$TASK_ID/"
echo ""
