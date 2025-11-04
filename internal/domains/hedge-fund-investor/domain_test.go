package hedgefundinvestor

import (
	"context"
	"strings"
	"testing"
	"time"

	registry "dsl-ob-poc/internal/domain-registry"
)

func TestNewDomain(t *testing.T) {
	domain := NewDomain()

	if domain == nil {
		t.Fatal("NewDomain() returned nil")
	}

	if domain.Name() != "hedge-fund-investor" {
		t.Errorf("Expected name 'hedge-fund-investor', got %s", domain.Name())
	}

	if domain.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", domain.Version())
	}

	if !domain.IsHealthy() {
		t.Error("Expected domain to be healthy")
	}

	vocab := domain.GetVocabulary()
	if vocab == nil {
		t.Fatal("Vocabulary should not be nil")
	}

	if len(vocab.Verbs) != 17 {
		t.Errorf("Expected 17 verbs, got %d", len(vocab.Verbs))
	}

	if len(vocab.Categories) != 7 {
		t.Errorf("Expected 7 categories, got %d", len(vocab.Categories))
	}

	expectedStates := []string{
		"OPPORTUNITY", "PRECHECKS", "KYC_PENDING", "KYC_APPROVED",
		"SUB_PENDING_CASH", "FUNDED_PENDING_NAV", "ISSUED", "ACTIVE",
		"REDEEM_PENDING", "REDEEMED", "OFFBOARDED",
	}

	states := domain.GetValidStates()
	if len(states) != len(expectedStates) {
		t.Errorf("Expected %d states, got %d", len(expectedStates), len(states))
	}

	for i, expected := range expectedStates {
		if states[i] != expected {
			t.Errorf("Expected state %s at position %d, got %s", expected, i, states[i])
		}
	}

	if domain.GetInitialState() != "OPPORTUNITY" {
		t.Errorf("Expected initial state 'OPPORTUNITY', got %s", domain.GetInitialState())
	}
}

func TestDomain_ValidateVerbs(t *testing.T) {
	domain := NewDomain()

	tests := []struct {
		name    string
		dsl     string
		wantErr bool
	}{
		{
			name:    "Valid opportunity verb",
			dsl:     "(investor.start-opportunity :legal-name \"Test Corp\" :type \"CORPORATE\")",
			wantErr: false,
		},
		{
			name:    "Valid KYC verb",
			dsl:     "(kyc.begin :investor \"uuid-123\" :tier \"STANDARD\")",
			wantErr: false,
		},
		{
			name:    "Valid subscription verb",
			dsl:     "(subscribe.request :investor \"uuid\" :fund \"fund-uuid\" :class \"class-uuid\" :amount 1000000.00 :currency \"USD\" :trade-date \"2024-01-15\" :value-date \"2024-01-15\")",
			wantErr: false,
		},
		{
			name:    "Multiple valid verbs",
			dsl:     "(investor.start-opportunity :legal-name \"Test\" :type \"INDIVIDUAL\")\n(kyc.begin :investor \"uuid-123\" :tier \"STANDARD\")",
			wantErr: false,
		},
		{
			name:    "Invalid verb",
			dsl:     "(invalid.verb :arg \"value\")",
			wantErr: true,
		},
		{
			name:    "Empty DSL",
			dsl:     "",
			wantErr: true,
		},
		{
			name:    "Whitespace only",
			dsl:     "   \n  \t  ",
			wantErr: true,
		},
		{
			name:    "No verbs",
			dsl:     "just some text without verbs",
			wantErr: true,
		},
		{
			name:    "Mixed valid and invalid verbs",
			dsl:     "(investor.start-opportunity :legal-name \"Test\" :type \"INDIVIDUAL\")\n(invalid.verb :arg \"value\")",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateVerbs(tt.dsl)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVerbs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDomain_ValidateStateTransition(t *testing.T) {
	domain := NewDomain()

	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		// Valid transitions
		{
			name:    "OPPORTUNITY to PRECHECKS",
			from:    "OPPORTUNITY",
			to:      "PRECHECKS",
			wantErr: false,
		},
		{
			name:    "PRECHECKS to KYC_PENDING",
			from:    "PRECHECKS",
			to:      "KYC_PENDING",
			wantErr: false,
		},
		{
			name:    "KYC_PENDING to KYC_APPROVED",
			from:    "KYC_PENDING",
			to:      "KYC_APPROVED",
			wantErr: false,
		},
		{
			name:    "KYC_APPROVED to SUB_PENDING_CASH",
			from:    "KYC_APPROVED",
			to:      "SUB_PENDING_CASH",
			wantErr: false,
		},
		{
			name:    "SUB_PENDING_CASH to FUNDED_PENDING_NAV",
			from:    "SUB_PENDING_CASH",
			to:      "FUNDED_PENDING_NAV",
			wantErr: false,
		},
		{
			name:    "FUNDED_PENDING_NAV to ISSUED",
			from:    "FUNDED_PENDING_NAV",
			to:      "ISSUED",
			wantErr: false,
		},
		{
			name:    "ISSUED to ACTIVE",
			from:    "ISSUED",
			to:      "ACTIVE",
			wantErr: false,
		},
		{
			name:    "ACTIVE to REDEEM_PENDING",
			from:    "ACTIVE",
			to:      "REDEEM_PENDING",
			wantErr: false,
		},
		{
			name:    "REDEEM_PENDING to REDEEMED",
			from:    "REDEEM_PENDING",
			to:      "REDEEMED",
			wantErr: false,
		},
		{
			name:    "REDEEMED to OFFBOARDED",
			from:    "REDEEMED",
			to:      "OFFBOARDED",
			wantErr: false,
		},
		// Invalid transitions
		{
			name:    "OPPORTUNITY to KYC_PENDING (skip PRECHECKS)",
			from:    "OPPORTUNITY",
			to:      "KYC_PENDING",
			wantErr: true,
		},
		{
			name:    "ACTIVE to OPPORTUNITY (backward)",
			from:    "ACTIVE",
			to:      "OPPORTUNITY",
			wantErr: true,
		},
		{
			name:    "OFFBOARDED to ACTIVE (from terminal)",
			from:    "OFFBOARDED",
			to:      "ACTIVE",
			wantErr: true,
		},
		{
			name:    "Invalid source state",
			from:    "INVALID_STATE",
			to:      "ACTIVE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateStateTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStateTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDomain_GenerateDSL(t *testing.T) {
	domain := NewDomain()
	ctx := context.Background()

	tests := []struct {
		name         string
		instruction  string
		context      map[string]interface{}
		expectedVerb string
		expectedDSL  string
		wantErr      bool
	}{
		{
			name:         "Create investor opportunity",
			instruction:  "start opportunity for John Smith",
			context:      nil,
			expectedVerb: "investor.start-opportunity",
			expectedDSL:  "investor.start-opportunity",
			wantErr:      false,
		},
		{
			name:        "Begin KYC with investor ID",
			instruction: "begin kyc",
			context: map[string]interface{}{
				"investor_id": "uuid-123",
			},
			expectedVerb: "kyc.begin",
			expectedDSL:  "kyc.begin",
			wantErr:      false,
		},
		{
			name:        "Begin KYC without investor ID",
			instruction: "start kyc",
			context:     map[string]interface{}{},
			wantErr:     true,
		},
		{
			name:        "Approve KYC with investor ID",
			instruction: "approve kyc",
			context: map[string]interface{}{
				"investor_id": "uuid-456",
			},
			expectedVerb: "kyc.approve",
			expectedDSL:  "kyc.approve",
			wantErr:      false,
		},
		{
			name:        "Submit subscription",
			instruction: "subscribe to fund",
			context: map[string]interface{}{
				"investor_id": "uuid-789",
			},
			expectedVerb: "subscribe.request",
			expectedDSL:  "subscribe.request",
			wantErr:      false,
		},
		{
			name:        "Unsupported instruction",
			instruction: "do something unknown",
			context:     nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &registry.GenerationRequest{
				Instruction: tt.instruction,
				SessionID:   "test-session",
				Context:     tt.context,
				Timestamp:   time.Now(),
			}

			resp, err := domain.GenerateDSL(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDSL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp == nil {
					t.Fatal("Expected response, got nil")
				}

				if resp.Verb != tt.expectedVerb {
					t.Errorf("Expected verb %s, got %s", tt.expectedVerb, resp.Verb)
				}

				if !strings.Contains(resp.DSL, tt.expectedDSL) {
					t.Errorf("Expected DSL to contain %s, got %s", tt.expectedDSL, resp.DSL)
				}

				if !resp.IsValid {
					t.Error("Expected valid response")
				}

				if resp.Confidence <= 0 {
					t.Error("Expected positive confidence")
				}

				if resp.Explanation == "" {
					t.Error("Expected non-empty explanation")
				}
			}
		})
	}
}

func TestDomain_GenerateDSL_NilRequest(t *testing.T) {
	domain := NewDomain()
	ctx := context.Background()

	_, err := domain.GenerateDSL(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil request")
	}
}

func TestDomain_GetCurrentState(t *testing.T) {
	domain := NewDomain()

	tests := []struct {
		name          string
		context       map[string]interface{}
		expectedState string
		wantErr       bool
	}{
		{
			name:          "Nil context",
			context:       nil,
			expectedState: "OPPORTUNITY",
			wantErr:       false,
		},
		{
			name:          "Empty context",
			context:       map[string]interface{}{},
			expectedState: "OPPORTUNITY",
			wantErr:       false,
		},
		{
			name: "Valid state in context",
			context: map[string]interface{}{
				"current_state": "KYC_PENDING",
			},
			expectedState: "KYC_PENDING",
			wantErr:       false,
		},
		{
			name: "Another valid state",
			context: map[string]interface{}{
				"current_state": "ACTIVE",
			},
			expectedState: "ACTIVE",
			wantErr:       false,
		},
		{
			name: "Invalid state in context",
			context: map[string]interface{}{
				"current_state": "INVALID_STATE",
			},
			expectedState: "",
			wantErr:       true,
		},
		{
			name: "Non-string state in context",
			context: map[string]interface{}{
				"current_state": 123,
			},
			expectedState: "OPPORTUNITY",
			wantErr:       false,
		},
		{
			name: "Context with other keys",
			context: map[string]interface{}{
				"investor_id": "uuid-123",
				"fund_id":     "fund-uuid",
			},
			expectedState: "OPPORTUNITY",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := domain.GetCurrentState(tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if state != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, state)
			}
		})
	}
}

func TestDomain_ExtractContext(t *testing.T) {
	domain := NewDomain()

	tests := []struct {
		name          string
		dsl           string
		expectedKey   string
		expectedVal   interface{}
		checkState    bool
		expectedState string
	}{
		{
			name:        "Empty DSL",
			dsl:         "",
			expectedKey: "",
			expectedVal: nil,
		},
		{
			name:        "Whitespace DSL",
			dsl:         "   \n  \t  ",
			expectedKey: "",
			expectedVal: nil,
		},
		{
			name:          "Start opportunity DSL",
			dsl:           "(investor.start-opportunity :legal-name \"John Smith\" :type \"INDIVIDUAL\")",
			checkState:    true,
			expectedState: "OPPORTUNITY",
		},
		{
			name:          "Record indication DSL",
			dsl:           "(investor.record-indication :investor \"uuid-123\" :fund \"fund-uuid\")",
			expectedKey:   "investor_id",
			expectedVal:   "uuid-123",
			checkState:    true,
			expectedState: "PRECHECKS",
		},
		{
			name:          "KYC begin DSL",
			dsl:           "(kyc.begin :investor \"uuid-456\" :tier \"STANDARD\")",
			expectedKey:   "investor_id",
			expectedVal:   "uuid-456",
			checkState:    true,
			expectedState: "KYC_PENDING",
		},
		{
			name:          "KYC approve DSL",
			dsl:           "(kyc.approve :investor \"uuid-789\" :risk \"MEDIUM\")",
			expectedKey:   "investor_id",
			expectedVal:   "uuid-789",
			checkState:    true,
			expectedState: "KYC_APPROVED",
		},
		{
			name:          "Subscribe request DSL",
			dsl:           "(subscribe.request :investor \"uuid-abc\" :fund \"fund-uuid\")",
			expectedKey:   "investor_id",
			expectedVal:   "uuid-abc",
			checkState:    true,
			expectedState: "SUB_PENDING_CASH",
		},
		{
			name: "Multi-line DSL",
			dsl: `(investor.start-opportunity
  :legal-name "Acme Capital LP"
  :type "CORPORATE")
(kyc.begin
  :investor "uuid-def"
  :tier "STANDARD")`,
			expectedKey:   "investor_id",
			expectedVal:   "uuid-def",
			checkState:    true,
			expectedState: "KYC_PENDING",
		},
		{
			name:        "DSL with placeholder",
			dsl:         "(kyc.begin :investor \"<investor_id>\" :tier \"STANDARD\")",
			expectedKey: "",
			expectedVal: nil,
		},
		{
			name:        "DSL without investor argument",
			dsl:         "(deal.nav :fund \"fund-uuid\" :class \"class-uuid\" :nav-date \"2024-01-15\" :nav 100.50)",
			expectedKey: "",
			expectedVal: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, err := domain.ExtractContext(tt.dsl)
			if err != nil {
				t.Errorf("ExtractContext() error = %v", err)
				return
			}

			if context == nil {
				t.Fatal("Expected context map, got nil")
			}

			if tt.expectedKey != "" {
				if val, exists := context[tt.expectedKey]; !exists {
					t.Errorf("Expected context key %s not found", tt.expectedKey)
				} else if val != tt.expectedVal {
					t.Errorf("Expected %v for key %s, got %v", tt.expectedVal, tt.expectedKey, val)
				}
			}

			if tt.checkState {
				if state, exists := context["current_state"]; !exists {
					t.Error("Expected current_state to be set")
				} else if state != tt.expectedState {
					t.Errorf("Expected state %s, got %s", tt.expectedState, state)
				}
			}
		})
	}
}

func TestDomain_VocabularyStructure(t *testing.T) {
	domain := NewDomain()
	vocab := domain.GetVocabulary()

	// Test that all expected verbs are present
	expectedVerbs := []string{
		// Opportunity management
		"investor.start-opportunity",
		"investor.record-indication",
		// KYC/Compliance
		"kyc.begin",
		"kyc.collect-doc",
		"kyc.screen",
		"kyc.approve",
		"kyc.refresh-schedule",
		// Ongoing monitoring
		"screen.continuous",
		// Tax & banking
		"tax.capture",
		"bank.set-instruction",
		// Subscription workflow
		"subscribe.request",
		"cash.confirm",
		"deal.nav",
		"subscribe.issue",
		// Redemption & offboarding
		"redeem.request",
		"redeem.settle",
		"offboard.close",
	}

	if len(vocab.Verbs) != len(expectedVerbs) {
		t.Errorf("Expected %d verbs, got %d", len(expectedVerbs), len(vocab.Verbs))
	}

	for _, verbName := range expectedVerbs {
		if verb, exists := vocab.Verbs[verbName]; !exists {
			t.Errorf("Expected verb %s not found", verbName)
		} else {
			// Test verb structure
			if verb.Name != verbName {
				t.Errorf("Verb name mismatch: expected %s, got %s", verbName, verb.Name)
			}

			if verb.Category == "" {
				t.Errorf("Verb %s has empty category", verbName)
			}

			if verb.Version == "" {
				t.Errorf("Verb %s has empty version", verbName)
			}

			if verb.Description == "" {
				t.Errorf("Verb %s has empty description", verbName)
			}

			if verb.Arguments == nil {
				t.Errorf("Verb %s has nil arguments", verbName)
			}

			if len(verb.Examples) == 0 {
				t.Errorf("Verb %s has no examples", verbName)
			}
		}
	}

	// Test categories
	expectedCategories := []string{
		"opportunity", "kyc", "monitoring", "tax-banking",
		"subscription", "redemption", "offboarding",
	}

	if len(vocab.Categories) != len(expectedCategories) {
		t.Errorf("Expected %d categories, got %d", len(expectedCategories), len(vocab.Categories))
	}

	for _, categoryName := range expectedCategories {
		if category, exists := vocab.Categories[categoryName]; !exists {
			t.Errorf("Expected category %s not found", categoryName)
		} else {
			if category.Name != categoryName {
				t.Errorf("Category name mismatch: expected %s, got %s", categoryName, category.Name)
			}

			if category.Description == "" {
				t.Errorf("Category %s has empty description", categoryName)
			}

			if len(category.Verbs) == 0 {
				t.Errorf("Category %s has no verbs", categoryName)
			}

			if category.Color == "" {
				t.Errorf("Category %s has empty color", categoryName)
			}

			if category.Icon == "" {
				t.Errorf("Category %s has empty icon", categoryName)
			}
		}
	}
}

func TestDomain_ArgumentTypes(t *testing.T) {
	domain := NewDomain()
	vocab := domain.GetVocabulary()

	// Test specific argument types for key verbs
	tests := []struct {
		verbName string
		argName  string
		argType  registry.ArgumentType
		required bool
	}{
		{"investor.start-opportunity", "legal-name", registry.ArgumentTypeString, true},
		{"investor.start-opportunity", "type", registry.ArgumentTypeEnum, true},
		{"investor.start-opportunity", "domicile", registry.ArgumentTypeString, false},
		{"investor.record-indication", "investor", registry.ArgumentTypeUUID, true},
		{"investor.record-indication", "ticket", registry.ArgumentTypeDecimal, true},
		{"kyc.begin", "tier", registry.ArgumentTypeEnum, false},
		{"kyc.approve", "risk", registry.ArgumentTypeEnum, true},
		{"kyc.approve", "refresh-due", registry.ArgumentTypeDate, true},
		{"subscribe.request", "amount", registry.ArgumentTypeDecimal, true},
		{"subscribe.request", "trade-date", registry.ArgumentTypeDate, true},
		{"redeem.request", "percentage", registry.ArgumentTypeDecimal, false},
	}

	for _, tt := range tests {
		t.Run(tt.verbName+"_"+tt.argName, func(t *testing.T) {
			verb, exists := vocab.Verbs[tt.verbName]
			if !exists {
				t.Fatalf("Verb %s not found", tt.verbName)
			}

			arg, exists := verb.Arguments[tt.argName]
			if !exists {
				t.Fatalf("Argument %s not found in verb %s", tt.argName, tt.verbName)
			}

			if arg.Type != tt.argType {
				t.Errorf("Expected argument type %s, got %s", tt.argType, arg.Type)
			}

			if arg.Required != tt.required {
				t.Errorf("Expected required %t, got %t", tt.required, arg.Required)
			}
		})
	}
}

func TestDomain_StateTransitions(t *testing.T) {
	domain := NewDomain()
	vocab := domain.GetVocabulary()

	// Test specific state transitions
	tests := []struct {
		verbName  string
		fromState []string
		toState   string
	}{
		{"investor.start-opportunity", nil, "OPPORTUNITY"},
		{"investor.record-indication", []string{"OPPORTUNITY"}, "PRECHECKS"},
		{"kyc.begin", []string{"PRECHECKS"}, "KYC_PENDING"},
		{"kyc.approve", []string{"KYC_PENDING"}, "KYC_APPROVED"},
		{"subscribe.request", []string{"KYC_APPROVED"}, "SUB_PENDING_CASH"},
		{"cash.confirm", []string{"SUB_PENDING_CASH"}, "FUNDED_PENDING_NAV"},
		{"subscribe.issue", []string{"FUNDED_PENDING_NAV"}, "ACTIVE"},
		{"redeem.request", []string{"ACTIVE"}, "REDEEM_PENDING"},
		{"redeem.settle", []string{"REDEEM_PENDING"}, "REDEEMED"},
		{"offboard.close", []string{"REDEEMED"}, "OFFBOARDED"},
	}

	for _, tt := range tests {
		t.Run(tt.verbName, func(t *testing.T) {
			verb, exists := vocab.Verbs[tt.verbName]
			if !exists {
				t.Fatalf("Verb %s not found", tt.verbName)
			}

			if verb.StateTransition == nil {
				t.Fatalf("Verb %s has no state transition", tt.verbName)
			}

			transition := verb.StateTransition

			if tt.fromState == nil {
				// No specific from state requirement
				if len(transition.FromStates) != 0 && transition.FromStates[0] != "" {
					t.Errorf("Expected no from states, got %v", transition.FromStates)
				}
			} else {
				if len(transition.FromStates) != len(tt.fromState) {
					t.Errorf("Expected %d from states, got %d", len(tt.fromState), len(transition.FromStates))
				} else {
					for i, expected := range tt.fromState {
						if transition.FromStates[i] != expected {
							t.Errorf("Expected from state %s, got %s", expected, transition.FromStates[i])
						}
					}
				}
			}

			if transition.ToState != tt.toState {
				t.Errorf("Expected to state %s, got %s", tt.toState, transition.ToState)
			}
		})
	}
}

func TestDomain_Metrics(t *testing.T) {
	domain := NewDomain()
	metrics := domain.GetMetrics()

	if metrics == nil {
		t.Fatal("Expected metrics, got nil")
	}

	if metrics.TotalVerbs != 17 {
		t.Errorf("Expected 17 total verbs, got %d", metrics.TotalVerbs)
	}

	if metrics.ActiveVerbs != 17 {
		t.Errorf("Expected 17 active verbs, got %d", metrics.ActiveVerbs)
	}

	if !metrics.IsHealthy {
		t.Error("Expected healthy metrics")
	}

	if metrics.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", metrics.Version)
	}
}

func TestExtractName(t *testing.T) {
	tests := []struct {
		instruction string
		expected    string
	}{
		{"start opportunity for John Smith", "John Smith"},
		{"create investor Jane Doe", "Jane Doe"},
		{"begin process for Alice Johnson", "Alice Johnson"},
		{"start opportunity for Acme Corp", "Acme Corp"},
		{"create investor Bob", "Bob"},
		{"no name here", ""},
		{"for", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.instruction, func(t *testing.T) {
			result := extractName(tt.instruction)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
