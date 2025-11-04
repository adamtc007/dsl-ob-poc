package state

import (
	"testing"

	"dsl-ob-poc/internal/hf-investor/domain"

	"github.com/google/uuid"
)

func TestNewHedgeFundStateMachine(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	if sm == nil {
		t.Fatal("Expected state machine to be created, got nil")
	}

	if len(sm.transitions) == 0 {
		t.Error("Expected state machine to have transitions defined")
	}

	if len(sm.guards) == 0 {
		t.Error("Expected state machine to have guard conditions defined")
	}
}

func TestCanTransition(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	tests := []struct {
		name      string
		fromState string
		toState   string
		expected  bool
	}{
		// Valid forward transitions
		{"Opportunity to Prechecks", domain.InvestorStatusOpportunity, domain.InvestorStatusPrechecks, true},
		{"Prechecks to KYC Pending", domain.InvestorStatusPrechecks, domain.InvestorStatusKYCPending, true},
		{"KYC Pending to KYC Approved", domain.InvestorStatusKYCPending, domain.InvestorStatusKYCApproved, true},
		{"KYC Approved to Sub Pending Cash", domain.InvestorStatusKYCApproved, domain.InvestorStatusSubPendingCash, true},
		{"Sub Pending Cash to Funded Pending NAV", domain.InvestorStatusSubPendingCash, domain.InvestorStatusFundedPendingNAV, true},
		{"Funded Pending NAV to Issued", domain.InvestorStatusFundedPendingNAV, domain.InvestorStatusIssued, true},
		{"Issued to Active", domain.InvestorStatusIssued, domain.InvestorStatusActive, true},
		{"Active to Redeem Pending", domain.InvestorStatusActive, domain.InvestorStatusRedeemPending, true},
		{"Redeem Pending to Redeemed", domain.InvestorStatusRedeemPending, domain.InvestorStatusRedeemed, true},
		{"Redeemed to Offboarded", domain.InvestorStatusRedeemed, domain.InvestorStatusOffboarded, true},

		// Valid backward transitions
		{"KYC Pending back to Prechecks", domain.InvestorStatusKYCPending, domain.InvestorStatusPrechecks, true},
		{"Sub Pending Cash back to KYC Approved", domain.InvestorStatusSubPendingCash, domain.InvestorStatusKYCApproved, true},
		{"Redeem Pending back to Active", domain.InvestorStatusRedeemPending, domain.InvestorStatusActive, true},

		// Valid re-investment transitions
		{"Active additional subscription", domain.InvestorStatusActive, domain.InvestorStatusSubPendingCash, true},
		{"Redeemed reinvestment", domain.InvestorStatusRedeemed, domain.InvestorStatusSubPendingCash, true},

		// Invalid transitions
		{"Skip Prechecks", domain.InvestorStatusOpportunity, domain.InvestorStatusKYCPending, false},
		{"Skip KYC", domain.InvestorStatusPrechecks, domain.InvestorStatusSubPendingCash, false},
		{"Skip cash confirmation", domain.InvestorStatusKYCApproved, domain.InvestorStatusFundedPendingNAV, false},
		{"Direct to Active", domain.InvestorStatusOpportunity, domain.InvestorStatusActive, false},
		{"From Offboarded", domain.InvestorStatusOffboarded, domain.InvestorStatusActive, false},

		// Self-transitions (invalid)
		{"Same state", domain.InvestorStatusActive, domain.InvestorStatusActive, false},

		// Invalid states
		{"From invalid state", "INVALID_STATE", domain.InvestorStatusActive, false},
		{"To invalid state", domain.InvestorStatusActive, "INVALID_STATE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.CanTransition(tt.fromState, tt.toState)
			if result != tt.expected {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.fromState, tt.toState, result, tt.expected)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	investor := &domain.HedgeFundInvestor{
		InvestorID:   uuid.New(),
		InvestorCode: "TEST-001",
		Type:         domain.InvestorTypeCorporate,
		LegalName:    "Test Corporation",
		Domicile:     "US",
		Status:       domain.InvestorStatusOpportunity,
	}

	tests := []struct {
		name        string
		investor    *domain.HedgeFundInvestor
		toState     string
		context     map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid transition without guards",
			investor:    investor,
			toState:     domain.InvestorStatusPrechecks,
			context:     map[string]interface{}{"indication": true},
			expectError: false,
		},
		{
			name:        "Invalid structural transition",
			investor:    investor,
			toState:     domain.InvestorStatusActive,
			context:     map[string]interface{}{},
			expectError: true,
			errorMsg:    "invalid state transition",
		},
		{
			name:        "Valid transition with missing guard condition",
			investor:    investor,
			toState:     domain.InvestorStatusPrechecks,
			context:     map[string]interface{}{}, // Missing indication
			expectError: true,
			errorMsg:    "guard condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the investor's status for each test
			testInvestor := *tt.investor
			testInvestor.Status = domain.InvestorStatusOpportunity

			err := sm.ValidateTransition(&testInvestor, tt.toState, tt.context)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', but got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestTransitionState(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	investor := &domain.HedgeFundInvestor{
		InvestorID:   uuid.New(),
		InvestorCode: "TEST-001",
		Type:         domain.InvestorTypeCorporate,
		LegalName:    "Test Corporation",
		Domicile:     "US",
		Status:       domain.InvestorStatusOpportunity,
	}

	// Test successful transition
	context := map[string]interface{}{
		"indication": true,
	}

	lifecycleState, err := sm.TransitionState(
		investor,
		domain.InvestorStatusPrechecks,
		"investor.record-indication",
		context,
		"test-user",
	)

	if err != nil {
		t.Errorf("Expected successful transition but got error: %s", err.Error())
	}

	if lifecycleState == nil {
		t.Fatal("Expected lifecycle state to be created")
	}

	// Verify the lifecycle state
	if lifecycleState.InvestorID != investor.InvestorID {
		t.Errorf("Expected investor ID %s, got %s", investor.InvestorID, lifecycleState.InvestorID)
	}

	if lifecycleState.ToState != domain.InvestorStatusPrechecks {
		t.Errorf("Expected to state %s, got %s", domain.InvestorStatusPrechecks, lifecycleState.ToState)
	}

	if lifecycleState.FromState == nil || *lifecycleState.FromState != domain.InvestorStatusOpportunity {
		t.Errorf("Expected from state %s", domain.InvestorStatusOpportunity)
	}

	if lifecycleState.TransitionTrigger == nil || *lifecycleState.TransitionTrigger != "investor.record-indication" {
		t.Errorf("Expected trigger 'investor.record-indication'")
	}

	if lifecycleState.TransitionedBy == nil || *lifecycleState.TransitionedBy != "test-user" {
		t.Errorf("Expected transitioned by 'test-user'")
	}

	// Verify the investor status was updated
	if investor.Status != domain.InvestorStatusPrechecks {
		t.Errorf("Expected investor status to be updated to %s, got %s", domain.InvestorStatusPrechecks, investor.Status)
	}

	// Test failed transition
	_, err = sm.TransitionState(
		investor,
		domain.InvestorStatusActive, // Invalid transition
		"invalid.verb",
		map[string]interface{}{},
		"test-user",
	)

	if err == nil {
		t.Errorf("Expected error for invalid transition but got none")
	}
}

func TestGetValidTransitions(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	tests := []struct {
		name           string
		fromState      string
		expectedCount  int
		expectedStates []string
	}{
		{
			name:           "From Opportunity",
			fromState:      domain.InvestorStatusOpportunity,
			expectedCount:  1,
			expectedStates: []string{domain.InvestorStatusPrechecks},
		},
		{
			name:           "From Active",
			fromState:      domain.InvestorStatusActive,
			expectedCount:  2,
			expectedStates: []string{domain.InvestorStatusRedeemPending, domain.InvestorStatusSubPendingCash},
		},
		{
			name:           "From Offboarded",
			fromState:      domain.InvestorStatusOffboarded,
			expectedCount:  0,
			expectedStates: []string{},
		},
		{
			name:           "Invalid state",
			fromState:      "INVALID_STATE",
			expectedCount:  0,
			expectedStates: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transitions := sm.GetValidTransitions(tt.fromState)

			if len(transitions) != tt.expectedCount {
				t.Errorf("Expected %d transitions, got %d", tt.expectedCount, len(transitions))
			}

			for _, expectedState := range tt.expectedStates {
				found := false
				for _, actualState := range transitions {
					if actualState == expectedState {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected transition to %s not found", expectedState)
				}
			}
		})
	}
}

func TestGetGuardConditions(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	// Test transition with guard conditions
	guards := sm.GetGuardConditions(domain.InvestorStatusOpportunity, domain.InvestorStatusPrechecks)
	if len(guards) == 0 {
		t.Error("Expected guard conditions for opportunity to prechecks transition")
	}

	// Verify specific guard condition
	foundIndicationGuard := false
	for _, guard := range guards {
		if guard.Name == "indication_recorded" {
			foundIndicationGuard = true
			if guard.Description == "" {
				t.Error("Expected guard condition to have description")
			}
			if guard.Check == nil {
				t.Error("Expected guard condition to have check function")
			}
		}
	}

	if !foundIndicationGuard {
		t.Error("Expected to find indication_recorded guard condition")
	}

	// Test transition without guard conditions
	guards = sm.GetGuardConditions(domain.InvestorStatusIssued, domain.InvestorStatusActive)
	if len(guards) != 0 {
		t.Errorf("Expected no guard conditions for issued to active transition, got %d", len(guards))
	}

	// Test invalid transition
	guards = sm.GetGuardConditions("INVALID_FROM", "INVALID_TO")
	if len(guards) != 0 {
		t.Errorf("Expected no guard conditions for invalid transition, got %d", len(guards))
	}
}

func TestIsTerminalState(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	tests := []struct {
		name     string
		state    string
		expected bool
	}{
		{"Offboarded is terminal", domain.InvestorStatusOffboarded, true},
		{"Opportunity is not terminal", domain.InvestorStatusOpportunity, false},
		{"Active is not terminal", domain.InvestorStatusActive, false},
		{"Invalid state is terminal", "INVALID_STATE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.IsTerminalState(tt.state)
			if result != tt.expected {
				t.Errorf("IsTerminalState(%s) = %v, want %v", tt.state, result, tt.expected)
			}
		})
	}
}

func TestGetAllStates(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	states := sm.GetAllStates()

	expectedStates := []string{
		domain.InvestorStatusOpportunity,
		domain.InvestorStatusPrechecks,
		domain.InvestorStatusKYCPending,
		domain.InvestorStatusKYCApproved,
		domain.InvestorStatusSubPendingCash,
		domain.InvestorStatusFundedPendingNAV,
		domain.InvestorStatusIssued,
		domain.InvestorStatusActive,
		domain.InvestorStatusRedeemPending,
		domain.InvestorStatusRedeemed,
		domain.InvestorStatusOffboarded,
	}

	if len(states) < len(expectedStates) {
		t.Errorf("Expected at least %d states, got %d", len(expectedStates), len(states))
	}

	// Verify all expected states are present
	for _, expectedState := range expectedStates {
		found := false
		for _, actualState := range states {
			if actualState == expectedState {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected state %s not found in GetAllStates result", expectedState)
		}
	}
}

func TestStateTransitionPath(t *testing.T) {
	sm := NewHedgeFundStateMachine()

	tests := []struct {
		name        string
		fromState   string
		toState     string
		expectError bool
		minLength   int
	}{
		{
			name:        "Same state",
			fromState:   domain.InvestorStatusActive,
			toState:     domain.InvestorStatusActive,
			expectError: false,
			minLength:   1,
		},
		{
			name:        "Adjacent states",
			fromState:   domain.InvestorStatusOpportunity,
			toState:     domain.InvestorStatusPrechecks,
			expectError: false,
			minLength:   2,
		},
		{
			name:        "Multi-step path",
			fromState:   domain.InvestorStatusOpportunity,
			toState:     domain.InvestorStatusActive,
			expectError: false,
			minLength:   7, // Should go through all intermediate states
		},
		{
			name:        "No path exists",
			fromState:   domain.InvestorStatusOffboarded,
			toState:     domain.InvestorStatusActive,
			expectError: true,
		},
		{
			name:        "Invalid from state",
			fromState:   "INVALID_STATE",
			toState:     domain.InvestorStatusActive,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := sm.StateTransitionPath(tt.fromState, tt.toState)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}

				if len(path) < tt.minLength {
					t.Errorf("Expected path length >= %d, got %d", tt.minLength, len(path))
				}

				// Verify path starts and ends correctly
				if len(path) > 0 {
					if path[0] != tt.fromState {
						t.Errorf("Expected path to start with %s, got %s", tt.fromState, path[0])
					}
					if path[len(path)-1] != tt.toState {
						t.Errorf("Expected path to end with %s, got %s", tt.toState, path[len(path)-1])
					}
				}

				// Verify all transitions in path are valid
				for i := 0; i < len(path)-1; i++ {
					if !sm.CanTransition(path[i], path[i+1]) {
						t.Errorf("Invalid transition in path: %s -> %s", path[i], path[i+1])
					}
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			str[:len(substr)] == substr ||
			str[len(str)-len(substr):] == substr ||
			containsInMiddle(str, substr))
}

func containsInMiddle(str, substr string) bool {
	for i := 1; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
