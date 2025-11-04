#!/bin/bash

API="http://localhost:8080/api/chat"

echo "=== Full Workflow Test ==="
echo ""

# Step 1: Create opportunity
echo "1️⃣  Creating opportunity for Acme Capital LP (Swiss corporate investor)..."
RESP1=$(curl -s -X POST $API \
  -H "Content-Type: application/json" \
  -d '{"message":"Create opportunity for Acme Capital LP, a corporate investor from Switzerland"}')
SESSION_ID=$(echo "$RESP1" | jq -r '.session_id')
echo "✓ Session: $SESSION_ID"
echo ""

# Step 2: Start KYC
echo "2️⃣  Starting KYC process..."
RESP2=$(curl -s -X POST $API \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"$SESSION_ID\",\"message\":\"Start standard KYC for this investor\"}")
echo ""

# Step 3: Collect document
echo "3️⃣  Collecting certificate of incorporation..."
RESP3=$(curl -s -X POST $API \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"$SESSION_ID\",\"message\":\"Collect certificate of incorporation document\"}")
echo ""

echo "=== Final Accumulated DSL ==="
echo "$RESP3" | jq -r '.dsl'
echo ""

echo "=== Verification ==="
DSL=$(echo "$RESP3" | jq -r '.dsl')
echo -n "✓ Contains investor.start-opportunity: "
echo "$DSL" | grep -q "investor.start-opportunity" && echo "YES" || echo "NO"
echo -n "✓ Contains kyc.begin: "
echo "$DSL" | grep -q "kyc.begin" && echo "YES" || echo "NO"
echo -n "✓ Contains kyc.collect-doc: "
echo "$DSL" | grep -q "kyc.collect-doc" && echo "YES" || echo "NO"
echo -n "✓ No placeholders (<investor_id>): "
echo "$DSL" | grep -q "<investor_id>" && echo "FAIL" || echo "PASS"
