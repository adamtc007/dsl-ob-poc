#!/bin/bash

################################################################################
# Hedge Fund Investor UI Launcher Script
#
# This script starts all necessary services for the Hedge Fund DSL UI:
# - Backend Go server on port 8080
# - Mock implementations for testing
#
# Usage: ./run-hf-ui.sh
################################################################################

set -e  # Exit on error

# Get the script directory (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$SCRIPT_DIR/hedge-fund-investor-source/web"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print banner
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}  Hedge Fund DSL Agent UI Launcher${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Default API Key (for demonstration - replace with your own)
DEFAULT_API_KEY="DEMO_API_KEY"
API_KEY=${GEMINI_API_KEY:-$DEFAULT_API_KEY}

################################################################################
# Cleanup function to stop all services
################################################################################
cleanup() {
    echo ""
    echo -e "${YELLOW}Stopping all services...${NC}"

    # Kill processes by PID files
    for pidfile in "$WEB_DIR"/*.pid; do
        if [ -f "$pidfile" ]; then
            PID=$(cat "$pidfile")
            if ps -p $PID > /dev/null 2>&1; then
                echo -e "  Stopping process ${PID}..."
                kill $PID 2>/dev/null || true
            fi
            rm "$pidfile"
        fi
    done

    # Kill any processes on ports 8080 and 5173
    echo -e "  Cleaning up ports 8080 and 5173..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    lsof -ti:5173 | xargs kill -9 2>/dev/null || true

    echo -e "${GREEN}Cleanup complete!${NC}"
    exit 0
}

# Register cleanup function to run on exit
trap cleanup EXIT INT TERM

################################################################################
# Check prerequisites
################################################################################
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21+ to continue.${NC}"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} Go $(go version | awk '{print $3}')"

# Check if the web directory exists
if [ ! -d "$WEB_DIR" ]; then
    echo -e "${RED}Error: Web directory not found at $WEB_DIR${NC}"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} Web directory found"

################################################################################
# Kill any existing processes
################################################################################
echo ""
echo -e "${YELLOW}Cleaning up existing processes...${NC}"
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
lsof -ti:5173 | xargs kill -9 2>/dev/null || true
echo -e "  ${GREEN}✓${NC} Ports 8080 and 5173 cleared"

################################################################################
# Setup directories and mock implementations
################################################################################
echo ""
echo -e "${YELLOW}Setting up environment...${NC}"

cd "$WEB_DIR"

# Create necessary directories
mkdir -p internal/datastore internal/hf-agent static/assets

# Create mock datastore if it doesn't exist
if [ ! -f "internal/datastore/datastore.go" ]; then
    echo -e "  Creating mock datastore implementation..."
    cat > internal/datastore/datastore.go << 'EOF'
package datastore

import (
    "context"
    "encoding/json"
)

type DataStore interface {
    Close() error
    GetLatestDSLWithState(ctx context.Context, cbuID string) (*DSLState, error)
    InsertDSLWithState(ctx context.Context, cbuID string, dslText string, state string) (string, error)
    ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error)
}

type DSLState struct {
    ID              string `json:"id"`
    CBUID           string `json:"cbu_id"`
    DSLText         string `json:"dsl_text"`
    OnboardingState string `json:"onboarding_state"`
    VersionNumber   int    `json:"version_number"`
    CreatedAt       string `json:"created_at"`
}

type MockDataStore struct{}

func NewMockDataStore() *MockDataStore {
    return &MockDataStore{}
}

func (m *MockDataStore) Close() error {
    return nil
}

func (m *MockDataStore) GetLatestDSLWithState(ctx context.Context, cbuID string) (*DSLState, error) {
    return &DSLState{
        ID:              "mock-id",
        CBUID:           cbuID,
        DSLText:         "(case.create\n  (cbu.id \"" + cbuID + "\")\n  (nature-purpose \"UCITS hedge fund\"))",
        OnboardingState: "CREATE",
        VersionNumber:   1,
        CreatedAt:       "2023-01-01T00:00:00Z",
    }, nil
}

func (m *MockDataStore) InsertDSLWithState(ctx context.Context, cbuID string, dslText string, state string) (string, error) {
    return "mock-version-" + cbuID, nil
}

func (m *MockDataStore) ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error) {
    return json.RawMessage(`"mock-value"`), map[string]any{"source": "mock"}, "RESOLVED", nil
}
EOF
fi

# Create mock agent if it doesn't exist
if [ ! -f "internal/hf-agent/agent.go" ]; then
    echo -e "  Creating mock agent implementation..."
    cat > internal/hf-agent/agent.go << 'EOF'
package hfagent

import (
    "context"
    "strings"
    "time"

    "github.com/google/uuid"
)

type HedgeFundDSLAgent struct {
    apiKey string
}

type DSLGenerationRequest struct {
    Instruction  string                 `json:"instruction"`
    CurrentState string                 `json:"current_state,omitempty"`
    InvestorID   string                 `json:"investor_id,omitempty"`
    Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

type DSLGenerationResponse struct {
    Verb        string                 `json:"verb"`
    Parameters  map[string]interface{} `json:"parameters"`
    Explanation string                 `json:"explanation"`
    DSL         string                 `json:"dsl"`
    FromState   string                 `json:"from_state"`
    ToState     string                 `json:"to_state"`
    Confidence  float64                `json:"confidence"`
    GeneratedAt string                 `json:"generated_at"`
}

func NewHedgeFundDSLAgent(ctx context.Context, apiKey string) (*HedgeFundDSLAgent, error) {
    return &HedgeFundDSLAgent{apiKey: apiKey}, nil
}

func (a *HedgeFundDSLAgent) Close() error {
    return nil
}

func (a *HedgeFundDSLAgent) GenerateDSL(ctx context.Context, req DSLGenerationRequest) (*DSLGenerationResponse, error) {
    instruction := strings.ToLower(req.Instruction)
    state := req.CurrentState
    if state == "" {
        state = "CREATE"
    }

    toState := "OPPORTUNITY"
    verb := "investor.start-opportunity"
    dsl := "(investor.start-opportunity\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0002} = \"CORPORATE\"\n  @attr{uuid-0003} = \"CH\")"

    if state == "OPPORTUNITY" && (strings.Contains(instruction, "kyc") || strings.Contains(instruction, "begin")) {
        toState = "KYC_PENDING"
        verb = "kyc.begin"
        dsl = "(kyc.begin\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  (jurisdictions\n    (jurisdiction \"CH\"))\n  (documents\n    (document \"CertificateOfIncorporation\")))"
    } else if state == "KYC_PENDING" && strings.Contains(instruction, "approve") {
        toState = "KYC_APPROVED"
        verb = "kyc.approve"
        dsl = "(kyc.approve\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0004} = \"LOW\")"
    } else if state == "KYC_APPROVED" && strings.Contains(instruction, "subscribe") {
        toState = "SUB_PENDING_CASH"
        verb = "subscribe.request"
        dsl = "(subscribe.request\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0006} = \"1000000\"\n  @attr{uuid-0007} = \"USD\")"
    }

    return &DSLGenerationResponse{
        Verb:        verb,
        Parameters:  map[string]interface{}{"investor": uuid.New().String()},
        Explanation: "Generated " + verb + " operation based on your request",
        DSL:         dsl,
        FromState:   state,
        ToState:     toState,
        Confidence:  0.95,
        GeneratedAt: time.Now().UTC().Format(time.RFC3339),
    }, nil
}

func GetHedgeFundDSLVocabulary() map[string]interface{} {
    return map[string]interface{}{
        "verbs": map[string]interface{}{
            "investor.start-opportunity": map[string]interface{}{
                "description":    "Start a new investor opportunity",
                "parameters":     []string{"legal-name", "type", "domicile"},
                "allowed_states": []string{"CREATE", "OPPORTUNITY"},
                "transitions_to": "OPPORTUNITY",
            },
        },
        "states": []string{"CREATE", "OPPORTUNITY", "KYC_PENDING", "KYC_APPROVED"},
    }
}
EOF
fi

# Fix static paths in server.go
echo -e "  Fixing static file paths in server.go..."
sed -i.bak 's|./web/static|./static|g' server.go 2>/dev/null

# Update Go dependencies
echo -e "  Updating Go dependencies..."
go mod tidy > /dev/null 2>&1

echo -e "  ${GREEN}✓${NC} Environment setup complete"

################################################################################
# Start backend server
################################################################################
echo ""
echo -e "${YELLOW}Starting backend server...${NC}"

export GEMINI_API_KEY="$API_KEY"

# Start the server in the background
go run server.go > backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > backend.pid

# Wait for server to start
sleep 3

# Check if server is running
if ps -p $BACKEND_PID > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} Backend server started (PID: $BACKEND_PID)"

    # Test the health endpoint
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Backend API is responding"
    else
        echo -e "  ${RED}✗${NC} Backend API is not responding"
        echo -e "  ${YELLOW}Check backend.log for details${NC}"
    fi
else
    echo -e "  ${RED}✗${NC} Backend server failed to start"
    echo -e "  ${YELLOW}Last 10 lines of backend.log:${NC}"
    tail -n 10 backend.log
    exit 1
fi

################################################################################
# Display connection information
################################################################################
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}   All services are running!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}🌐 Access the application:${NC}"
echo -e "   ${YELLOW}http://localhost:8080${NC}"
echo -e "   ${YELLOW}http://127.0.0.1:8080${NC}"
echo ""
echo -e "${GREEN}🔗 API endpoints:${NC}"
echo -e "   Health: ${YELLOW}http://localhost:8080/api/health${NC}"
echo -e "   Chat:   ${YELLOW}http://localhost:8080/api/chat${NC}"
echo ""
echo -e "${GREEN}💡 Try these example prompts:${NC}"
echo -e "   • ${BLUE}\"Create an investor opportunity for Swiss corporation Alpine Capital\"${NC}"
echo -e "   • ${BLUE}\"Start KYC process for the investor\"${NC}"
echo -e "   • ${BLUE}\"Approve KYC with low risk rating\"${NC}"
echo -e "   • ${BLUE}\"Create a subscription request for \$5 million USD\"${NC}"
echo ""
echo -e "${GREEN}ℹ️  Server logs:${NC}"
echo -e "   Backend: ${YELLOW}$WEB_DIR/backend.log${NC}"
echo ""
echo -e "${RED}Press Ctrl+C to stop all services${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Test the server with a sample request
echo -e "${YELLOW}Testing server with sample request...${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/chat \
    -H "Content-Type: application/json" \
    -d '{"message":"test"}')

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Server is responding to API requests${NC}"
else
    echo -e "${RED}✗ Server is not responding to API requests${NC}"
fi

echo ""
echo -e "${GREEN}Server is ready! Open your browser to http://localhost:8080${NC}"
echo ""

# Keep the script running until Ctrl+C
wait $BACKEND_PID
