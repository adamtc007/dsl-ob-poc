#!/bin/bash

# Hedge Fund Investor UI Setup and Run Script
# This script builds and runs all the necessary services for the Hedge Fund Investor UI

# Set the working directory to the script's location
cd "$(dirname "$0")"
SCRIPT_DIR="$(pwd)"

# Colors for console output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print banner
echo -e "${BLUE}================================${NC}"
echo -e "${GREEN}Hedge Fund DSL Agent UI Launcher${NC}"
echo -e "${BLUE}================================${NC}"

# Default API Key (for demonstration only - replace with your own)
DEFAULT_API_KEY="DEMO_API_KEY"
API_KEY=${GEMINI_API_KEY:-$DEFAULT_API_KEY}

# Function to clean up processes
cleanup() {
    echo -e "\n${YELLOW}Stopping all services...${NC}"

    # Kill backend server if PID file exists
    if [ -f "$SCRIPT_DIR/backend.pid" ]; then
        BACKEND_PID=$(cat "$SCRIPT_DIR/backend.pid")
        echo -e "Stopping backend server (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null || echo -e "${RED}Backend server was not running.${NC}"
        rm "$SCRIPT_DIR/backend.pid"
    fi

    # Kill frontend dev server if PID file exists
    if [ -f "$SCRIPT_DIR/frontend.pid" ]; then
        FRONTEND_PID=$(cat "$SCRIPT_DIR/frontend.pid")
        echo -e "Stopping frontend dev server (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null || echo -e "${RED}Frontend dev server was not running.${NC}"
        rm "$SCRIPT_DIR/frontend.pid"
    fi

    # Kill all child processes to make sure nothing is left running
    pkill -P $$ 2>/dev/null

    echo -e "${GREEN}Cleanup complete!${NC}"
    exit 0
}

# Cleanup on script exit
trap cleanup EXIT

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go 1.21+ to run this application.${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Node.js is not installed. Please install Node.js 18+ to run this application.${NC}"
    exit 1
fi

# Kill any existing processes using ports 8080 and 5173
echo -e "${YELLOW}Killing any existing processes on ports 8080 and 5173...${NC}"
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
lsof -ti:5173 | xargs kill -9 2>/dev/null || true

# Create necessary directories for mock implementation
echo -e "${YELLOW}Setting up required directories...${NC}"

mkdir -p "$SCRIPT_DIR/internal/datastore" 2>/dev/null
mkdir -p "$SCRIPT_DIR/internal/hf-agent" 2>/dev/null
mkdir -p "$SCRIPT_DIR/static" 2>/dev/null

# Create mock implementations if they don't exist
if [ ! -f "$SCRIPT_DIR/internal/datastore/datastore.go" ]; then
    echo -e "${YELLOW}Creating mock datastore implementation...${NC}"
    cat > "$SCRIPT_DIR/internal/datastore/datastore.go" << 'EOL'
package datastore

import (
    "context"
    "encoding/json"
)

// DataStore defines the interface for database operations
type DataStore interface {
    Close() error
    GetLatestDSLWithState(ctx context.Context, cbuID string) (*DSLState, error)
    InsertDSLWithState(ctx context.Context, cbuID string, dslText string, state string) (string, error)
    ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error)
}

// DSLState represents the DSL state record
type DSLState struct {
    ID              string `json:"id"`
    CBUID           string `json:"cbu_id"`
    DSLText         string `json:"dsl_text"`
    OnboardingState string `json:"onboarding_state"`
    VersionNumber   int    `json:"version_number"`
    CreatedAt       string `json:"created_at"`
}

// MockDataStore provides a minimal implementation for testing
type MockDataStore struct{}

// NewMockDataStore creates a new mock datastore
func NewMockDataStore() *MockDataStore {
    return &MockDataStore{}
}

// Close implements DataStore interface
func (m *MockDataStore) Close() error {
    return nil
}

// GetLatestDSLWithState implements DataStore interface
func (m *MockDataStore) GetLatestDSLWithState(ctx context.Context, cbuID string) (*DSLState, error) {
    // Return a mock state for testing
    return &DSLState{
        ID:              "mock-id",
        CBUID:           cbuID,
        DSLText:         "(case.create\n  (cbu.id \"" + cbuID + "\")\n  (nature-purpose \"UCITS hedge fund\"))",
        OnboardingState: "CREATE",
        VersionNumber:   1,
        CreatedAt:       "2023-01-01T00:00:00Z",
    }, nil
}

// InsertDSLWithState implements DataStore interface
func (m *MockDataStore) InsertDSLWithState(ctx context.Context, cbuID string, dslText string, state string) (string, error) {
    // Return a mock version ID
    return "mock-version-" + cbuID, nil
}

// ResolveValueFor implements DataStore interface
func (m *MockDataStore) ResolveValueFor(ctx context.Context, cbuID, attributeID string) (json.RawMessage, map[string]any, string, error) {
    // Return mock values
    return json.RawMessage(`"mock-value"`), map[string]any{"source": "mock"}, "RESOLVED", nil
}
EOL
fi

if [ ! -f "$SCRIPT_DIR/internal/hf-agent/agent.go" ]; then
    echo -e "${YELLOW}Creating mock agent implementation...${NC}"
    cat > "$SCRIPT_DIR/internal/hf-agent/agent.go" << 'EOL'
package hfagent

import (
    "context"
    "time"

    "github.com/google/uuid"
)

// HedgeFundDSLAgent manages interactions with the LLM for DSL generation
type HedgeFundDSLAgent struct {
    apiKey string
}

// DSLGenerationRequest represents a request to generate DSL
type DSLGenerationRequest struct {
    Instruction  string                 `json:"instruction"`
    CurrentState string                 `json:"current_state,omitempty"`
    InvestorID   string                 `json:"investor_id,omitempty"`
    Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// DSLGenerationResponse represents the response from the DSL generation
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

// NewHedgeFundDSLAgent creates a new agent with the given API key
func NewHedgeFundDSLAgent(ctx context.Context, apiKey string) (*HedgeFundDSLAgent, error) {
    return &HedgeFundDSLAgent{
        apiKey: apiKey,
    }, nil
}

// Close cleans up any resources used by the agent
func (a *HedgeFundDSLAgent) Close() error {
    return nil
}

// GenerateDSL generates DSL from natural language instructions
func (a *HedgeFundDSLAgent) GenerateDSL(ctx context.Context, req DSLGenerationRequest) (*DSLGenerationResponse, error) {
    // Mock response for demonstration purposes
    instruction := req.Instruction
    state := req.CurrentState
    if state == "" {
        state = "CREATE"
    }

    // Determine next state based on instruction and current state
    toState := "OPPORTUNITY"
    verb := "investor.start-opportunity"
    dsl := "(investor.start-opportunity\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0002} = \"CORPORATE\"\n  @attr{uuid-0003} = \"CH\")"

    if state == "OPPORTUNITY" && (contains(instruction, "kyc") || contains(instruction, "begin kyc")) {
        toState = "KYC_PENDING"
        verb = "kyc.begin"
        dsl = "(kyc.begin\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  (jurisdictions\n    (jurisdiction \"CH\"))\n  (documents\n    (document \"CertificateOfIncorporation\")\n    (document \"BoardResolution\")))"
    } else if state == "KYC_PENDING" && (contains(instruction, "approve") || contains(instruction, "kyc approve")) {
        toState = "KYC_APPROVED"
        verb = "kyc.approve"
        dsl = "(kyc.approve\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0004} = \"LOW\"\n  @attr{uuid-0005} = \"APPROVED\")"
    } else if state == "KYC_APPROVED" && (contains(instruction, "subscribe") || contains(instruction, "subscription")) {
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

// Helper function to check if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
    str, substr = toLowerCase(str), toLowerCase(substr)
    return str != "" && substr != "" && str != substr && str != "" && substr != "" && contains(str, substr)
}

// Helper function to convert a string to lowercase
func toLowerCase(str string) string {
    result := ""
    for _, c := range str {
        if c >= 'A' && c <= 'Z' {
            result += string(c + 32)
        } else {
            result += string(c)
        }
    }
    return result
}

// GetHedgeFundDSLVocabulary returns the DSL vocabulary
func GetHedgeFundDSLVocabulary() map[string]interface{} {
    // Mock vocabulary for demonstration purposes
    return map[string]interface{}{
        "verbs": map[string]interface{}{
            "investor.start-opportunity": map[string]interface{}{
                "description": "Start a new investor opportunity",
                "parameters": []string{"legal-name", "type", "domicile"},
                "allowed_states": []string{"CREATE", "OPPORTUNITY"},
                "transitions_to": "OPPORTUNITY",
            },
            "kyc.begin": map[string]interface{}{
                "description": "Begin KYC process for investor",
                "parameters": []string{"jurisdictions", "documents"},
                "allowed_states": []string{"OPPORTUNITY"},
                "transitions_to": "KYC_PENDING",
            },
            "kyc.approve": map[string]interface{}{
                "description": "Approve KYC for investor",
                "parameters": []string{"risk-rating", "status"},
                "allowed_states": []string{"KYC_PENDING"},
                "transitions_to": "KYC_APPROVED",
            },
        },
        "states": []string{
            "CREATE",
            "OPPORTUNITY",
            "KYC_PENDING",
            "KYC_APPROVED",
            "SUB_PENDING_CASH",
            "FUNDED_PENDING_NAV",
            "ISSUED",
            "ACTIVE",
        },
    }
}
EOL
fi

# Create go.mod file if it doesn't exist
if [ ! -f "$SCRIPT_DIR/go.mod" ]; then
    echo -e "${YELLOW}Creating go.mod file...${NC}"
    cat > "$SCRIPT_DIR/go.mod" << 'EOL'
module dsl-ob-poc/hedge-fund-investor-source/web

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.1
)

require golang.org/x/net v0.20.0 // indirect
EOL
fi

# Update Go dependencies
echo -e "${YELLOW}Updating Go dependencies...${NC}"
go mod tidy

# Install frontend dependencies
echo -e "${YELLOW}Installing frontend dependencies...${NC}"
cd "$SCRIPT_DIR/frontend"
npm install

# Set API key for the backend
export GEMINI_API_KEY="$API_KEY"
echo -e "${GREEN}Using API Key: ${YELLOW}$API_KEY${NC} (Set GEMINI_API_KEY env var to use your own key)"

# Start backend server
echo -e "${YELLOW}Starting backend server...${NC}"
cd "$SCRIPT_DIR"

# Fix the static paths in server.go if needed
sed -i.bak 's|./web/static|./static|g' server.go 2>/dev/null

go run server.go > backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > "$SCRIPT_DIR/backend.pid"
echo -e "${GREEN}Backend server started on port 8080 with PID: $BACKEND_PID${NC}"

# Give the backend a moment to start
sleep 2

# Check if backend is running
if ps -p $BACKEND_PID > /dev/null; then
    echo -e "${GREEN}Backend server is running.${NC}"
else
    echo -e "${RED}Backend server failed to start. Check backend.log for details.${NC}"
    echo -e "${YELLOW}Last 10 lines of backend.log:${NC}"
    tail -n 10 "$SCRIPT_DIR/backend.log"
    exit 1
fi

# Start frontend dev server
echo -e "${YELLOW}Starting frontend development server...${NC}"
cd "$SCRIPT_DIR/frontend"
npm run dev > frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > "$SCRIPT_DIR/frontend.pid"
echo -e "${GREEN}Frontend server started on port 5173 with PID: $FRONTEND_PID${NC}"

# Wait for frontend to start
sleep 5

# Display connection information
echo -e "${GREEN}All services are running!${NC}"
echo -e "${BLUE}================================${NC}"
echo -e "${GREEN}Backend Server:${YELLOW} http://localhost:8080${NC}"
echo -e "${GREEN}Backend API:${YELLOW} http://localhost:8080/api/health${NC}"
echo -e "${GREEN}Frontend Dev:${YELLOW} http://localhost:5173${NC}"
echo -e "${BLUE}================================${NC}"
echo -e ""
echo -e "${YELLOW}To test the UI:${NC}"
echo -e "  1. Open ${YELLOW}http://localhost:8080${NC} in your browser"
echo -e "  2. You should see the Hedge Fund DSL Chat Interface"
echo -e "  3. Try typing: ${GREEN}\"Create investor opportunity for Alpine Capital\"${NC}"
echo -e ""
echo -e "${GREEN}Press Ctrl+C to stop all services${NC}"
echo -e ""

# Wait for user to press Ctrl+C
wait
