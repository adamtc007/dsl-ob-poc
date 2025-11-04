package dsl

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetHedgeFundDSLVocabulary(t *testing.T) {
	vocab := GetHedgeFundDSLVocabulary()

	// Test vocabulary metadata
	if vocab.Domain != "hedge-fund-investor" {
		t.Errorf("Expected domain 'hedge-fund-investor', got %s", vocab.Domain)
	}

	if vocab.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", vocab.Version)
	}

	// Test that expected verbs exist
	expectedVerbs := []string{
		"investor.start-opportunity",
		"investor.record-indication",
		"kyc.begin",
		"kyc.collect-doc",
		"kyc.screen",
		"kyc.approve",
		"tax.capture",
		"bank.set-instruction",
		"subscribe.request",
		"cash.confirm",
		"deal.nav",
		"subscribe.issue",
		"kyc.refresh-schedule",
		"screen.continuous",
		"redeem.request",
		"redeem.settle",
		"offboard.close",
	}

	for _, expectedVerb := range expectedVerbs {
		if _, exists := vocab.Verbs[expectedVerb]; !exists {
			t.Errorf("Expected verb %s not found in vocabulary", expectedVerb)
		}
	}

	// Test a specific verb definition
	startOpportunityVerb := vocab.Verbs["investor.start-opportunity"]
	if startOpportunityVerb.Name != "investor.start-opportunity" {
		t.Errorf("Expected verb name 'investor.start-opportunity', got %s", startOpportunityVerb.Name)
	}

	if startOpportunityVerb.Domain != "hedge-fund-investor" {
		t.Errorf("Expected verb domain 'hedge-fund-investor', got %s", startOpportunityVerb.Domain)
	}

	if startOpportunityVerb.Category != "opportunity" {
		t.Errorf("Expected verb category 'opportunity', got %s", startOpportunityVerb.Category)
	}

	// Test verb arguments
	expectedArgs := []string{"legal-name", "type", "domicile", "source"}
	for _, expectedArg := range expectedArgs {
		if _, exists := startOpportunityVerb.Args[expectedArg]; !exists {
			t.Errorf("Expected argument %s not found in verb %s", expectedArg, startOpportunityVerb.Name)
		}
	}

	// Test required vs optional arguments
	if !startOpportunityVerb.Args["legal-name"].Required {
		t.Errorf("Expected legal-name to be required")
	}

	if startOpportunityVerb.Args["source"].Required {
		t.Errorf("Expected source to be optional")
	}

	// Test enum values
	typeArg := startOpportunityVerb.Args["type"]
	if typeArg.Type != "enum" {
		t.Errorf("Expected type argument to be enum, got %s", typeArg.Type)
	}

	expectedEnumValues := []string{"INDIVIDUAL", "CORPORATE", "TRUST", "FOHF", "NOMINEE"}
	if len(typeArg.Values) != len(expectedEnumValues) {
		t.Errorf("Expected %d enum values, got %d", len(expectedEnumValues), len(typeArg.Values))
	}

	// Test state transitions
	if startOpportunityVerb.StateChange == nil {
		t.Errorf("Expected state change definition for investor.start-opportunity")
	} else {
		if len(startOpportunityVerb.StateChange.FromStates) != 0 {
			t.Errorf("Expected no from states for initial verb, got %d", len(startOpportunityVerb.StateChange.FromStates))
		}

		if startOpportunityVerb.StateChange.ToState != "OPPORTUNITY" {
			t.Errorf("Expected to state 'OPPORTUNITY', got %s", startOpportunityVerb.StateChange.ToState)
		}
	}
}

func TestGenerateHedgeFundDSL(t *testing.T) {
	operation := &HedgeFundDSLOperation{
		Verb: "investor.start-opportunity",
		Args: map[string]interface{}{
			"legal-name": "Test Corporation",
			"type":       "CORPORATE",
			"domicile":   "US",
			"source":     "Referral",
		},
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	dslText := GenerateHedgeFundDSL(operation)

	// Test that DSL contains expected elements
	expectedElements := []string{
		"(investor.start-opportunity",
		":legal-name \"Test Corporation\"",
		":type \"CORPORATE\"",
		":domicile \"US\"",
		":source \"Referral\"",
		")",
	}

	for _, element := range expectedElements {
		if !contains(dslText, element) {
			t.Errorf("Expected DSL to contain '%s', but it didn't. DSL:\n%s", element, dslText)
		}
	}
}

func TestGenerateHedgeFundDSLWithUUID(t *testing.T) {
	investorID := uuid.New()
	fundID := uuid.New()

	operation := &HedgeFundDSLOperation{
		Verb: "investor.record-indication",
		Args: map[string]interface{}{
			"investor": investorID,
			"fund":     fundID,
			"ticket":   "1000000.00",
			"currency": "USD",
		},
		Timestamp: time.Now(),
	}

	dslText := GenerateHedgeFundDSL(operation)

	// Test that UUIDs are properly formatted as strings
	expectedElements := []string{
		"(investor.record-indication",
		":investor \"" + investorID.String() + "\"",
		":fund \"" + fundID.String() + "\"",
		":ticket \"1000000.00\"",
		":currency \"USD\"",
		")",
	}

	for _, element := range expectedElements {
		if !contains(dslText, element) {
			t.Errorf("Expected DSL to contain '%s', but it didn't. DSL:\n%s", element, dslText)
		}
	}
}

func TestGenerateHedgeFundDSLPlan(t *testing.T) {
	planID := uuid.New()
	investorID := uuid.New()

	operation1 := HedgeFundDSLOperation{
		Verb: "investor.start-opportunity",
		Args: map[string]interface{}{
			"legal-name": "Test Corporation",
			"type":       "CORPORATE",
			"domicile":   "US",
		},
		Timestamp: time.Now(),
	}

	operation2 := HedgeFundDSLOperation{
		Verb: "kyc.begin",
		Args: map[string]interface{}{
			"investor": investorID.String(),
			"tier":     "STANDARD",
		},
		Timestamp: time.Now(),
	}

	plan := &HedgeFundDSLPlan{
		PlanID:      planID,
		InvestorID:  investorID,
		Description: "Complete investor onboarding",
		Operations:  []HedgeFundDSLOperation{operation1, operation2},
		Variables: map[string]interface{}{
			"investor-name": "Test Corporation",
			"target-amount": "1000000.00",
		},
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
	}

	dslText := GenerateHedgeFundDSLPlan(plan)

	// Test that plan DSL contains expected elements
	expectedElements := []string{
		"(hedge-fund.plan",
		":plan-id \"" + planID.String() + "\"",
		":investor-id \"" + investorID.String() + "\"",
		":description \"Complete investor onboarding\"",
		"(variables",
		":investor-name \"Test Corporation\"",
		":target-amount \"1000000.00\"",
		"(operations",
		"(investor.start-opportunity",
		"(kyc.begin",
		")",
	}

	for _, element := range expectedElements {
		if !contains(dslText, element) {
			t.Errorf("Expected plan DSL to contain '%s', but it didn't. DSL:\n%s", element, dslText)
		}
	}
}

func TestParseHedgeFundInvestorVars(t *testing.T) {
	dslText := `
	(investor.start-opportunity
	  :legal-name ?INV.LEGAL_NAME
	  :type ?INV.TYPE
	  :domicile ?INV.DOMICILE)

	(kyc.approve
	  :investor ?INV.ID
	  :risk ?KYC.RISK
	  :refresh-due ?KYC.REFRESH_DUE)

	(tax.capture
	  :investor ?INV.ID
	  :fatca ?TAX.FATCA_CLASS
	  :crs ?TAX.CRS_CLASS)

	(bank.set-instruction
	  :investor ?INV.ID
	  :currency "USD"
	  :iban ?BANK[USD].IBAN
	  :swift ?BANK[USD].SWIFT)

	(subscribe.request
	  :investor ?INV.ID
	  :amount ?TRADE.AMOUNT
	  :trade-date ?DATE.TRADE)
	`

	vars := ParseHedgeFundInvestorVars(dslText)

	expectedVars := map[string]string{
		"?INV.LEGAL_NAME":  "LEGAL_NAME",
		"?INV.TYPE":        "TYPE",
		"?INV.DOMICILE":    "DOMICILE",
		"?INV.ID":          "ID",
		"?KYC.RISK":        "RISK",
		"?KYC.REFRESH_DUE": "REFRESH_DUE",
		"?TAX.FATCA_CLASS": "FATCA_CLASS",
		"?TAX.CRS_CLASS":   "CRS_CLASS",
		"?BANK[USD].IBAN":  "USD.IBAN",
		"?BANK[USD].SWIFT": "USD.SWIFT",
		"?TRADE.AMOUNT":    "AMOUNT",
		"?DATE.TRADE":      "TRADE",
	}

	if len(vars) != len(expectedVars) {
		t.Errorf("Expected %d variables, got %d", len(expectedVars), len(vars))
	}

	for expectedVar, expectedValue := range expectedVars {
		if actualValue, exists := vars[expectedVar]; !exists {
			t.Errorf("Expected variable %s not found", expectedVar)
		} else if actualValue != expectedValue {
			t.Errorf("Expected variable %s to have value %s, got %s", expectedVar, expectedValue, actualValue)
		}
	}
}

func TestValidateHedgeFundDSLOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation *HedgeFundDSLOperation
		expectErr bool
		errorMsg  string
	}{
		{
			name: "Valid investor.start-opportunity",
			operation: &HedgeFundDSLOperation{
				Verb: "investor.start-opportunity",
				Args: map[string]interface{}{
					"legal-name": "Test Corporation",
					"type":       "CORPORATE",
					"domicile":   "US",
				},
				Timestamp: time.Now(),
			},
			expectErr: false,
		},
		{
			name: "Missing required argument",
			operation: &HedgeFundDSLOperation{
				Verb: "investor.start-opportunity",
				Args: map[string]interface{}{
					"legal-name": "Test Corporation",
					// Missing required "type" and "domicile"
				},
				Timestamp: time.Now(),
			},
			expectErr: true,
			errorMsg:  "required argument",
		},
		{
			name: "Invalid enum value",
			operation: &HedgeFundDSLOperation{
				Verb: "investor.start-opportunity",
				Args: map[string]interface{}{
					"legal-name": "Test Corporation",
					"type":       "INVALID_TYPE",
					"domicile":   "US",
				},
				Timestamp: time.Now(),
			},
			expectErr: true,
			errorMsg:  "invalid enum value",
		},
		{
			name: "Invalid UUID format",
			operation: &HedgeFundDSLOperation{
				Verb: "kyc.begin",
				Args: map[string]interface{}{
					"investor": "not-a-uuid",
					"tier":     "STANDARD",
				},
				Timestamp: time.Now(),
			},
			expectErr: true,
			errorMsg:  "invalid UUID format",
		},
		{
			name: "Valid UUID",
			operation: &HedgeFundDSLOperation{
				Verb: "kyc.begin",
				Args: map[string]interface{}{
					"investor": uuid.New().String(),
					"tier":     "STANDARD",
				},
				Timestamp: time.Now(),
			},
			expectErr: false,
		},
		{
			name: "Unknown verb",
			operation: &HedgeFundDSLOperation{
				Verb:      "unknown.verb",
				Args:      map[string]interface{}{},
				Timestamp: time.Now(),
			},
			expectErr: true,
			errorMsg:  "unknown hedge fund DSL verb",
		},
		{
			name: "Invalid date format",
			operation: &HedgeFundDSLOperation{
				Verb: "kyc.approve",
				Args: map[string]interface{}{
					"investor":    uuid.New().String(),
					"risk":        "LOW",
					"refresh-due": "invalid-date",
					"approved-by": "Test User",
				},
				Timestamp: time.Now(),
			},
			expectErr: true,
			errorMsg:  "invalid date format",
		},
		{
			name: "Valid date format",
			operation: &HedgeFundDSLOperation{
				Verb: "kyc.approve",
				Args: map[string]interface{}{
					"investor":    uuid.New().String(),
					"risk":        "LOW",
					"refresh-due": "2024-12-31",
					"approved-by": "Test User",
				},
				Timestamp: time.Now(),
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHedgeFundDSLOperation(tt.operation)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !contains(err.Error(), tt.errorMsg) {
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

func TestHedgeFundDSLContext(t *testing.T) {
	ctx := &HedgeFundDSLContext{
		InvestorID: uuid.New(),
		FundID:     func() *uuid.UUID { id := uuid.New(); return &id }(),
		Variables: map[string]interface{}{
			"investor-name": "Test Corporation",
			"target-amount": 1000000.00,
		},
	}

	// Test that context can be serialized to JSON
	jsonData, err := json.Marshal(ctx)
	if err != nil {
		t.Errorf("Failed to marshal context to JSON: %s", err.Error())
	}

	// Test that context can be deserialized from JSON
	var deserializedCtx HedgeFundDSLContext
	err = json.Unmarshal(jsonData, &deserializedCtx)
	if err != nil {
		t.Errorf("Failed to unmarshal context from JSON: %s", err.Error())
	}

	// Test that values are preserved
	if deserializedCtx.InvestorID != ctx.InvestorID {
		t.Errorf("InvestorID not preserved through JSON serialization")
	}

	if *deserializedCtx.FundID != *ctx.FundID {
		t.Errorf("FundID not preserved through JSON serialization")
	}

	if deserializedCtx.Variables["investor-name"] != ctx.Variables["investor-name"] {
		t.Errorf("Variables not preserved through JSON serialization")
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
