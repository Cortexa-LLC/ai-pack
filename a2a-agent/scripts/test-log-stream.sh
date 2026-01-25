#!/bin/bash
# Test script for log streaming endpoint

SERVER_URL="http://localhost:8080"

echo "📡 Testing Log Streaming Endpoints"
echo "=================================="
echo ""

# Test 1: Get recent logs
echo "1️⃣  Fetching recent logs (limit=10):"
echo ""
curl -s "${SERVER_URL}/logs/recent?limit=10" | python3 -m json.tool
echo ""
echo ""

# Test 2: Stream logs (run for 10 seconds)
echo "2️⃣  Streaming realtime logs (10 seconds):"
echo ""
echo "   Press Ctrl+C to stop streaming"
echo ""
timeout 10 curl -s -N "${SERVER_URL}/logs/stream" || echo ""
echo ""
echo "✅ Test complete"
