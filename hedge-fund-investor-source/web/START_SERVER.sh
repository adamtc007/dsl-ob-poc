#!/bin/bash
# START_SERVER.sh - Startup script for Hedge Fund DSL Web Server
# Tests the new fuzzy entity resolution and multi-turn conversation features

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  Hedge Fund DSL Agent - Entity Resolution Testing             ║"
echo "╔════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Step 1: Check environment
echo -e "${YELLOW}Step 1: Checking environment...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go not found. Please install Go 1.21+${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go found: $(go version)${NC}"

# Check if we're in the right directory
if [ ! -f "server.go" ]; then
    echo -e "${RED}✗ Must run from web/ directory${NC}"
    echo "  cd hedge-fund-investor-source/web"
    exit 1
fi
echo -e "${GREEN}✓ In correct directory${NC}"

# Step 2: Check database connection (optional)
echo ""
echo -e "${YELLOW}Step 2: Database configuration...${NC}"

if [ -z "$DB_CONN_STRING" ]; then
    echo -e "${YELLOW}⚠ DB_CONN_STRING not set${NC}"
    echo "  Running in MOCK mode (in-memory only)"
    echo "  To use PostgreSQL, set:"
    echo -e "${BLUE}  export DB_CONN_STRING=\"postgres://localhost:5432/hf_investor?sslmode=disable\"${NC}"
    echo ""
    MODE="MOCK"
else
    echo -e "${GREEN}✓ DB_CONN_STRING set${NC}"
    echo "  Testing connection..."
    if psql "$DB_CONN_STRING" -c "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✓ Database connection successful${NC}"
        MODE="POSTGRES"
    else
        echo -e "${RED}✗ Cannot connect to database${NC}"
        echo "  Falling back to MOCK mode"
        unset DB_CONN_STRING
        MODE="MOCK"
    fi
fi

# Step 3: Check API key (optional)
echo ""
echo -e "${YELLOW}Step 3: AI Agent configuration...${NC}"

if [ -z "$GEMINI_API_KEY" ] && [ -z "$GOOGLE_API_KEY" ]; then
    echo -e "${YELLOW}⚠ GEMINI_API_KEY not set${NC}"
    echo "  Agent will use simulated responses"
    echo "  For real AI, set:"
    echo -e "${BLUE}  export GEMINI_API_KEY=\"your-google-ai-key\"${NC}"
    AGENT="SIMULATED"
else
    echo -e "${GREEN}✓ API key found${NC}"
    AGENT="GEMINI"
fi

# Step 4: Build the server
echo ""
echo -e "${YELLOW}Step 4: Building server...${NC}"

if go build -o hf-web-server . 2>&1; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Step 5: Display configuration summary
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  SERVER CONFIGURATION                                          ║${NC}"
echo -e "${GREEN}╠════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║${NC}  Mode:       ${BLUE}$MODE${NC}"
echo -e "${GREEN}║${NC}  Agent:      ${BLUE}$AGENT${NC}"
echo -e "${GREEN}║${NC}  Port:       ${BLUE}8080${NC}"
echo -e "${GREEN}║${NC}  URL:        ${BLUE}http://localhost:8080${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"

# Step 6: Display test scenarios
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  TEST SCENARIOS - Entity Resolution                           ║${NC}"
echo -e "${BLUE}╠════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${YELLOW}Scenario 1: Create New Investor${NC}"
echo -e "${BLUE}║${NC}    Type: ${GREEN}\"Create opportunity for Alpine Capital AG\"${NC}"
echo -e "${BLUE}║${NC}    Expected: Confirms new investor creation"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${YELLOW}Scenario 2: Exact Match${NC}"
echo -e "${BLUE}║${NC}    Type: ${GREEN}\"Start KYC for Alpine Capital AG\"${NC}"
echo -e "${BLUE}║${NC}    Expected: Auto-selects exact match (if exists)"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${YELLOW}Scenario 3: Fuzzy Match (Typo)${NC}"
echo -e "${BLUE}║${NC}    Type: ${GREEN}\"Start KYC for Alpne Captial\"${NC}"
echo -e "${BLUE}║${NC}    Expected: Finds \"Alpine Capital\" with similarity score"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${YELLOW}Scenario 4: Multiple Matches${NC}"
echo -e "${BLUE}║${NC}    Type: ${GREEN}\"Start KYC for Alpine Capital\"${NC}"
echo -e "${BLUE}║${NC}    Expected: Lists all \"Alpine Capital\" variants"
echo -e "${BLUE}║${NC}    Then type: ${GREEN}\"1\"${NC} to select first option"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${YELLOW}Scenario 5: Fund Resolution${NC}"
echo -e "${BLUE}║${NC}    Type: ${GREEN}\"Start KYC for Alpine Capital for Global Equity Fund\"${NC}"
echo -e "${BLUE}║${NC}    Expected: Resolves both investor and fund"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}║${NC}  ${YELLOW}Scenario 6: Cancel Action${NC}"
echo -e "${BLUE}║${NC}    Type: ${GREEN}\"Create opportunity for Test Corp\"${NC}"
echo -e "${BLUE}║${NC}    Then type: ${GREEN}\"no\"${NC} or ${GREEN}\"cancel\"${NC}"
echo -e "${BLUE}║${NC}    Expected: No investor created"
echo -e "${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"

# Step 7: Display database verification commands
if [ "$MODE" = "POSTGRES" ]; then
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  DATABASE VERIFICATION COMMANDS                                ║${NC}"
    echo -e "${BLUE}╠════════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}  ${YELLOW}View All Investors:${NC}"
    echo -e "${BLUE}║${NC}    ${GREEN}psql \$DB_CONN_STRING -c \"SELECT investor_code, legal_name, status FROM \\\"hf-investor\\\".hf_investors ORDER BY created_at DESC LIMIT 5\"${NC}"
    echo -e "${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}  ${YELLOW}View DSL History:${NC}"
    echo -e "${BLUE}║${NC}    ${GREEN}psql \$DB_CONN_STRING -c \"SELECT LEFT(dsl_text, 50), created_at FROM \\\"hf-investor\\\".hf_dsl_executions ORDER BY created_at DESC LIMIT 5\"${NC}"
    echo -e "${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}  ${YELLOW}View State Transitions:${NC}"
    echo -e "${BLUE}║${NC}    ${GREEN}psql \$DB_CONN_STRING -c \"SELECT from_state, to_state, transitioned_at FROM \\\"hf-investor\\\".hf_lifecycle_states ORDER BY transitioned_at DESC LIMIT 5\"${NC}"
    echo -e "${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
fi

# Step 8: Start the server
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  🚀 STARTING SERVER...                                         ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
echo ""

# Run the server
./hf-web-server

# Cleanup (only reached if server stops)
echo ""
echo -e "${YELLOW}Server stopped${NC}"
