#!/bin/bash

API="http://localhost:8080/api/chat"

echo "=== Test 1: Create opportunity for Henry Cearns ==="
RESP1=$(curl -s -X POST $API \
  -H "Content-Type: application/json" \
  -d '{"message":"Create an opportunity for investor Henry Cearns"}')

echo "$RESP1" | jq -r '.dsl'
SESSION_ID=$(echo "$RESP1" | jq -r '.session_id')
echo "Session ID: $SESSION_ID"
echo ""

echo "=== Test 2: Start KYC for Henry Cearns (should use his investor_id) ==="
RESP2=$(curl -s -X POST $API \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"$SESSION_ID\",\"message\":\"Start KYC for this investor\"}")

echo "$RESP2" | jq -r '.dsl'
echo ""

echo "=== Check if investor UUID was used (should NOT see <investor_id>) ==="
echo "$RESP2" | jq -r '.dsl' | grep -o '<investor_id>' && echo "❌ FAIL: Placeholder found" || echo "✅ PASS: Real UUID used"
