package dslstate

import (
	"context"
	"strings"
	"testing"
)

// MockAgent implements a simple mock for testing
type MockAgent struct {
	intents   []*OperationIntent
	callCount int
}

func (m *MockAgent) DecideOperation(ctx context.Context, instruction string, currentDSL string, context Context) (*OperationIntent, error) {
	if m.callCount >= len(m.intents) {
		return nil, nil
	}
	intent := m.intents[m.callCount]
	m.callCount++
	return intent, nil
}

func (m *MockAgent) Close() error {
	return nil
}

func TestNewManager(t *testing.T) {
	mockAgent := &MockAgent{}
	manager := NewManager(mockAgent)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.CurrentDSL != "" {
		t.Errorf("Expected empty CurrentDSL, got: %s", manager.CurrentDSL)
	}

	if manager.Context.InvestorID != "" {
		t.Errorf("Expected empty InvestorID, got: %s", manager.Context.InvestorID)
	}
}

func TestAppendOperation_StartOpportunity(t *testing.T) {
	mockAgent := &MockAgent{
		intents: []*OperationIntent{
			{
				Operation: "create_opportunity",
				Verb:      "investor.start-opportunity",
				Parameters: map[string]interface{}{
					"legal-name": "Acme Capital LP",
					"type":       "CORPORATE",
					"domicile":   "CH",
				},
				ToState:     "OPPORTUNITY",
				Explanation: "Created investor opportunity",
				Confidence:  0.95,
			},
		},
	}

	manager := NewManager(mockAgent)
	ctx := context.Background()

	completeDSL, fragment, response, err := manager.AppendOperation(ctx, "Create opportunity for Acme Capital LP")
	if err != nil {
		t.Fatalf("AppendOperation failed: %v", err)
	}

	// Check response metadata
	if response.Verb != "investor.start-opportunity" {
		t.Errorf("Expected verb 'investor.start-opportunity', got: %s", response.Verb)
	}

	// Check that DSL was generated
	if fragment == "" {
		t.Error("Fragment should not be empty")
	}

	// Check DSL structure
	if !strings.Contains(fragment, "(investor.start-opportunity") {
		t.Errorf("Fragment should contain verb, got: %s", fragment)
	}

	if !strings.Contains(fragment, "Acme Capital LP") {
		t.Errorf("Fragment should contain legal name, got: %s", fragment)
	}

	if !strings.Contains(fragment, "CORPORATE") {
		t.Errorf("Fragment should contain type, got: %s", fragment)
	}

	// Check that investor UUID was generated
	if manager.Context.InvestorID == "" {
		t.Error("InvestorID should have been generated")
	}

	// Check state transition
	if manager.Context.CurrentState != "OPPORTUNITY" {
		t.Errorf("Expected state 'OPPORTUNITY', got: %s", manager.Context.CurrentState)
	}

	// Check complete DSL equals fragment for first operation
	if completeDSL != fragment {
		t.Errorf("For first operation, completeDSL should equal fragment")
	}

	// Check accumulated DSL
	if manager.CurrentDSL != fragment {
		t.Errorf("CurrentDSL should equal fragment")
	}
}

func TestAppendOperation_KYCWithContext(t *testing.T) {
	mockAgent := &MockAgent{
		intents: []*OperationIntent{
			{
				Operation: "create_opportunity",
				Verb:      "investor.start-opportunity",
				Parameters: map[string]interface{}{
					"legal-name": "Test Investor",
					"type":       "INDIVIDUAL",
				},
				ToState:     "OPPORTUNITY",
				Explanation: "Created investor",
				Confidence:  0.95,
			},
			{
				Operation: "start_kyc",
				Verb:      "kyc.begin",
				Parameters: map[string]interface{}{
					"investor": "<investor_id>", // Placeholder - should be resolved
					"tier":     "STANDARD",
				},
				FromState:   "OPPORTUNITY",
				ToState:     "KYC_PENDING",
				Explanation: "Started KYC",
				Confidence:  0.95,
			},
		},
	}

	manager := NewManager(mockAgent)
	ctx := context.Background()

	// First operation - create opportunity
	_, _, _, err := manager.AppendOperation(ctx, "Create opportunity")
	if err != nil {
		t.Fatalf("First operation failed: %v", err)
	}

	investorUUID := manager.Context.InvestorID
	if investorUUID == "" {
		t.Fatal("InvestorID should have been generated")
	}

	// Second operation - start KYC (should use investor UUID from context)
	completeDSL, fragment, response, err := manager.AppendOperation(ctx, "Start KYC")
	if err != nil {
		t.Fatalf("Second operation failed: %v", err)
	}

	// Check that KYC operation uses actual UUID, not placeholder
	if strings.Contains(fragment, "<investor_id>") {
		t.Errorf("Fragment should not contain placeholder, got: %s", fragment)
	}

	if !strings.Contains(fragment, investorUUID) {
		t.Errorf("Fragment should contain actual investor UUID %s, got: %s", investorUUID, fragment)
	}

	// Check state transition
	if manager.Context.CurrentState != "KYC_PENDING" {
		t.Errorf("Expected state 'KYC_PENDING', got: %s", manager.Context.CurrentState)
	}

	// Check that complete DSL contains both operations
	if !strings.Contains(completeDSL, "investor.start-opportunity") {
		t.Error("Complete DSL should contain start-opportunity")
	}

	if !strings.Contains(completeDSL, "kyc.begin") {
		t.Error("Complete DSL should contain kyc.begin")
	}

	// Check DSL accumulation
	dslLines := strings.Split(completeDSL, "\n\n")
	if len(dslLines) != 2 {
		t.Errorf("Expected 2 operations separated by blank line, got: %d", len(dslLines))
	}

	// Verify response metadata
	if response.ToState != "KYC_PENDING" {
		t.Errorf("Expected ToState 'KYC_PENDING', got: %s", response.ToState)
	}
}

func TestAppendOperation_MultipleOperations(t *testing.T) {
	mockAgent := &MockAgent{
		intents: []*OperationIntent{
			{
				Operation:  "create_opportunity",
				Verb:       "investor.start-opportunity",
				Parameters: map[string]interface{}{"legal-name": "Investor1", "type": "INDIVIDUAL"},
				ToState:    "OPPORTUNITY",
			},
			{
				Operation:  "start_kyc",
				Verb:       "kyc.begin",
				Parameters: map[string]interface{}{"tier": "STANDARD"},
				ToState:    "KYC_PENDING",
			},
			{
				Operation:  "collect_document",
				Verb:       "kyc.collect-doc",
				Parameters: map[string]interface{}{"doc-type": "Passport", "subject": "Investor1"},
				ToState:    "KYC_PENDING",
			},
		},
	}

	manager := NewManager(mockAgent)
	ctx := context.Background()

	// Execute three operations
	for i := 0; i < 3; i++ {
		_, _, _, err := manager.AppendOperation(ctx, "operation")
		if err != nil {
			t.Fatalf("Operation %d failed: %v", i+1, err)
		}
	}

	// Check that all three operations are in accumulated DSL
	completeDSL := manager.GetCurrentDSL()
	operations := []string{"investor.start-opportunity", "kyc.begin", "kyc.collect-doc"}

	for _, op := range operations {
		if !strings.Contains(completeDSL, op) {
			t.Errorf("Complete DSL should contain %s, got: %s", op, completeDSL)
		}
	}

	// Check that operations are separated
	dslParts := strings.Split(completeDSL, "\n\n")
	if len(dslParts) != 3 {
		t.Errorf("Expected 3 operations, got: %d", len(dslParts))
	}
}

func TestValidateDSL_Success(t *testing.T) {
	manager := NewManager(&MockAgent{})

	tests := []struct {
		name string
		dsl  string
		verb string
	}{
		{
			name: "Valid investor.start-opportunity",
			dsl:  "(investor.start-opportunity\n  :legal-name \"Test\"\n  :type \"INDIVIDUAL\")",
			verb: "investor.start-opportunity",
		},
		{
			name: "Valid kyc.begin",
			dsl:  "(kyc.begin\n  :investor \"uuid-123\"\n  :tier \"STANDARD\")",
			verb: "kyc.begin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateDSL(tt.dsl, tt.verb)
			if err != nil {
				t.Errorf("Validation should pass, got error: %v", err)
			}
		})
	}
}

func TestValidateDSL_Failures(t *testing.T) {
	manager := NewManager(&MockAgent{})

	tests := []struct {
		name        string
		dsl         string
		verb        string
		expectError string
	}{
		{
			name:        "Missing opening paren",
			dsl:         "kyc.begin :tier \"STANDARD\")",
			verb:        "kyc.begin",
			expectError: "must start with",
		},
		{
			name:        "Missing closing paren",
			dsl:         "(kyc.begin :tier \"STANDARD\"",
			verb:        "kyc.begin",
			expectError: "must end with",
		},
		{
			name:        "Unresolved placeholder",
			dsl:         "(kyc.begin :investor \"<investor_id>\" :tier \"STANDARD\")",
			verb:        "kyc.begin",
			expectError: "unresolved placeholders",
		},
		{
			name:        "Verb not in DSL",
			dsl:         "(wrong.verb :test \"value\")",
			verb:        "kyc.begin",
			expectError: "does not contain expected verb",
		},
		{
			name:        "Unapproved verb",
			dsl:         "(invalid.verb :test \"value\")",
			verb:        "invalid.verb",
			expectError: "not in approved vocabulary",
		},
		{
			name:        "Unbalanced parentheses",
			dsl:         "((kyc.begin :tier \"STANDARD\")",
			verb:        "kyc.begin",
			expectError: "unbalanced parentheses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateDSL(tt.dsl, tt.verb)
			if err == nil {
				t.Error("Expected validation error, got nil")
			} else if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectError, err)
			}
		})
	}
}

func TestGenerateDSLFromIntent_InvestorStartOpportunity(t *testing.T) {
	manager := NewManager(&MockAgent{})

	intent := &OperationIntent{
		Operation: "create_opportunity",
		Verb:      "investor.start-opportunity",
		Parameters: map[string]interface{}{
			"legal-name": "Test Corp",
			"type":       "CORPORATE",
			"domicile":   "US",
		},
	}

	dsl := manager.generateDSLFromIntent(intent)

	expectedParts := []string{
		"(investor.start-opportunity",
		":legal-name \"Test Corp\"",
		":type \"CORPORATE\"",
		":domicile \"US\"",
		")",
	}

	for _, part := range expectedParts {
		if !strings.Contains(dsl, part) {
			t.Errorf("DSL should contain %q, got: %s", part, dsl)
		}
	}
}

func TestGenerateDSLFromIntent_KYCBegin(t *testing.T) {
	manager := NewManager(&MockAgent{})
	manager.Context.InvestorID = "test-uuid-123"

	intent := &OperationIntent{
		Operation: "start_kyc",
		Verb:      "kyc.begin",
		Parameters: map[string]interface{}{
			"tier": "ENHANCED",
		},
	}

	dsl := manager.generateDSLFromIntent(intent)

	expectedParts := []string{
		"(kyc.begin",
		":investor \"test-uuid-123\"", // Should use actual UUID from context
		":tier \"ENHANCED\"",
		")",
	}

	for _, part := range expectedParts {
		if !strings.Contains(dsl, part) {
			t.Errorf("DSL should contain %q, got: %s", part, dsl)
		}
	}

	// Should NOT contain placeholder
	if strings.Contains(dsl, "<investor_id>") {
		t.Errorf("DSL should not contain placeholder, got: %s", dsl)
	}
}

func TestExtractAndTrackEntities(t *testing.T) {
	manager := NewManager(&MockAgent{})

	intent := &OperationIntent{
		Operation: "create_opportunity",
		Verb:      "investor.start-opportunity",
		Parameters: map[string]interface{}{
			"legal-name": "Acme Corp",
			"type":       "CORPORATE",
			"domicile":   "UK",
		},
	}

	manager.extractAndTrackEntities(intent)

	// Check that investor UUID was generated
	if manager.Context.InvestorID == "" {
		t.Error("InvestorID should have been generated")
	}

	// Check that context was populated
	if manager.Context.InvestorName != "Acme Corp" {
		t.Errorf("Expected InvestorName 'Acme Corp', got: %s", manager.Context.InvestorName)
	}

	if manager.Context.InvestorType != "CORPORATE" {
		t.Errorf("Expected InvestorType 'CORPORATE', got: %s", manager.Context.InvestorType)
	}

	if manager.Context.Domicile != "UK" {
		t.Errorf("Expected Domicile 'UK', got: %s", manager.Context.Domicile)
	}

	// Check that UUID was added to intent parameters
	if id, ok := intent.Parameters["investor"].(string); !ok || id == "" {
		t.Error("Investor UUID should have been added to intent parameters")
	}
}

func TestReset(t *testing.T) {
	manager := NewManager(&MockAgent{})

	// Set some state
	manager.CurrentDSL = "(test dsl)"
	manager.Context.InvestorID = "test-uuid"
	manager.Context.CurrentState = "TEST_STATE"
	manager.Context.InvestorName = "Test"

	// Reset
	manager.Reset()

	// Check that everything was cleared
	if manager.CurrentDSL != "" {
		t.Errorf("CurrentDSL should be empty after reset, got: %s", manager.CurrentDSL)
	}

	if manager.Context.InvestorID != "" {
		t.Errorf("InvestorID should be empty after reset, got: %s", manager.Context.InvestorID)
	}

	if manager.Context.CurrentState != "" {
		t.Errorf("CurrentState should be empty after reset, got: %s", manager.Context.CurrentState)
	}

	if manager.Context.InvestorName != "" {
		t.Errorf("InvestorName should be empty after reset, got: %s", manager.Context.InvestorName)
	}
}

func TestGetCurrentDSL(t *testing.T) {
	manager := NewManager(&MockAgent{})
	manager.CurrentDSL = "(test dsl content)"

	dsl := manager.GetCurrentDSL()

	if dsl != manager.CurrentDSL {
		t.Errorf("GetCurrentDSL returned wrong value")
	}
}

func TestGetContext(t *testing.T) {
	manager := NewManager(&MockAgent{})
	manager.Context.InvestorID = "test-uuid"
	manager.Context.CurrentState = "TEST"

	ctx := manager.GetContext()

	if ctx.InvestorID != "test-uuid" {
		t.Errorf("GetContext returned wrong InvestorID")
	}

	if ctx.CurrentState != "TEST" {
		t.Errorf("GetContext returned wrong CurrentState")
	}
}
