package hedgefundinvestor

import (
	"context"
	"strings"
	"testing"
	"time"

	registry "dsl-ob-poc/internal/domain-registry"
)

func TestHedgeFundDomain_IntegrationWithRegistry(t *testing.T) {
	// Create domain registry
	domainRegistry := registry.NewRegistry()
	defer domainRegistry.Shutdown()

	// Create and register hedge fund domain
	hfDomain := NewDomain()
	err := domainRegistry.Register(hfDomain)
	if err != nil {
		t.Fatalf("Failed to register hedge fund domain: %v", err)
	}

	// Test domain registration
	retrievedDomain, err := domainRegistry.Get("hedge-fund-investor")
	if err != nil {
		t.Fatalf("Failed to retrieve registered domain: %v", err)
	}

	if retrievedDomain.Name() != "hedge-fund-investor" {
		t.Errorf("Expected domain name 'hedge-fund-investor', got %s", retrievedDomain.Name())
	}

	// Test vocabulary retrieval through registry
	vocab, err := domainRegistry.GetVocabulary("hedge-fund-investor")
	if err != nil {
		t.Fatalf("Failed to get vocabulary: %v", err)
	}

	if len(vocab.Verbs) != 17 {
		t.Errorf("Expected 17 verbs in vocabulary, got %d", len(vocab.Verbs))
	}

	// Test finding domains by verb
	domains := domainRegistry.FindDomainsByVerb("investor.start-opportunity")
	if len(domains) != 1 || domains[0] != "hedge-fund-investor" {
		t.Errorf("Expected to find hedge-fund-investor domain for verb, got %v", domains)
	}

	// Test finding domains by category
	domains = domainRegistry.FindDomainsByCategory("kyc")
	if len(domains) != 1 || domains[0] != "hedge-fund-investor" {
		t.Errorf("Expected to find hedge-fund-investor domain for kyc category, got %v", domains)
	}
}

func TestHedgeFundDomain_IntegrationWithRouter(t *testing.T) {
	// Create domain registry and router
	domainRegistry := registry.NewRegistry()
	defer domainRegistry.Shutdown()

	router := registry.NewRouter(domainRegistry)

	// Register hedge fund domain
	hfDomain := NewDomain()
	err := domainRegistry.Register(hfDomain)
	if err != nil {
		t.Fatalf("Failed to register hedge fund domain: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name           string
		request        *registry.RoutingRequest
		expectedDomain string
		expectedStrat  registry.RoutingStrategy
	}{
		{
			name: "Route by explicit switch",
			request: &registry.RoutingRequest{
				Message:   "switch to hedge fund investor domain",
				SessionID: "test-session",
				Timestamp: time.Now(),
			},
			expectedDomain: "hedge-fund-investor",
			expectedStrat:  registry.StrategyExplicit,
		},
		{
			name: "Route by context (investor_id)",
			request: &registry.RoutingRequest{
				Message:   "start KYC process",
				SessionID: "test-session",
				Context: map[string]interface{}{
					"investor_id": "uuid-123",
				},
				Timestamp: time.Now(),
			},
			expectedDomain: "hedge-fund-investor",
			expectedStrat:  registry.StrategyContext,
		},
		{
			name: "Route by DSL verb",
			request: &registry.RoutingRequest{
				Message:   "(kyc.begin :investor \"uuid-123\" :tier \"STANDARD\")",
				SessionID: "test-session",
				Timestamp: time.Now(),
			},
			expectedDomain: "hedge-fund-investor",
			expectedStrat:  registry.StrategyVerb,
		},
		{
			name: "Route by keyword",
			request: &registry.RoutingRequest{
				Message:   "process investor subscription",
				SessionID: "test-session",
				Timestamp: time.Now(),
			},
			expectedDomain: "hedge-fund-investor",
			expectedStrat:  registry.StrategyKeyword,
		},
		{
			name: "Route by hedge fund state",
			request: &registry.RoutingRequest{
				Message:   "continue process",
				SessionID: "test-session",
				Context: map[string]interface{}{
					"current_state": "KYC_PENDING",
				},
				Timestamp: time.Now(),
			},
			expectedDomain: "hedge-fund-investor",
			expectedStrat:  registry.StrategyContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := router.Route(ctx, tt.request)
			if err != nil {
				t.Errorf("Router.Route() failed: %v", err)
				return
			}

			if response == nil {
				t.Fatal("Expected routing response, got nil")
			}

			if response.DomainName != tt.expectedDomain {
				t.Errorf("Expected domain %s, got %s", tt.expectedDomain, response.DomainName)
			}

			if response.Strategy != tt.expectedStrat {
				t.Errorf("Expected strategy %s, got %s", tt.expectedStrat, response.Strategy)
			}

			if response.Domain == nil {
				t.Error("Expected domain instance, got nil")
			}

			if response.Confidence <= 0 {
				t.Error("Expected positive confidence")
			}
		})
	}
}

func TestHedgeFundDomain_EndToEndWorkflow(t *testing.T) {
	// Create complete system
	domainRegistry := registry.NewRegistry()
	defer domainRegistry.Shutdown()

	router := registry.NewRouter(domainRegistry)

	// Register hedge fund domain
	hfDomain := NewDomain()
	err := domainRegistry.Register(hfDomain)
	if err != nil {
		t.Fatalf("Failed to register hedge fund domain: %v", err)
	}

	ctx := context.Background()

	// Simulate complete hedge fund investor workflow
	workflow := []struct {
		name         string
		message      string
		context      map[string]interface{}
		expectedVerb string
	}{
		{
			name:         "1. Create opportunity",
			message:      "start opportunity for John Smith",
			context:      map[string]interface{}{},
			expectedVerb: "investor.start-opportunity",
		},
		{
			name:    "2. Begin KYC",
			message: "begin kyc process",
			context: map[string]interface{}{
				"investor_id": "uuid-123",
			},
			expectedVerb: "kyc.begin",
		},
		{
			name:    "3. Approve KYC",
			message: "approve kyc for investor",
			context: map[string]interface{}{
				"investor_id":   "uuid-123",
				"current_state": "KYC_PENDING",
			},
			expectedVerb: "kyc.approve",
		},
		{
			name:    "4. Submit subscription",
			message: "submit subscription request",
			context: map[string]interface{}{
				"investor_id":   "uuid-123",
				"current_state": "KYC_APPROVED",
			},
			expectedVerb: "subscribe.request",
		},
	}

	accumulatedDSL := ""
	currentContext := make(map[string]interface{})

	for _, step := range workflow {
		t.Run(step.name, func(t *testing.T) {
			// Merge step context with accumulated context
			requestContext := make(map[string]interface{})
			for k, v := range currentContext {
				requestContext[k] = v
			}
			for k, v := range step.context {
				requestContext[k] = v
			}

			// Route the request
			routingRequest := &registry.RoutingRequest{
				Message:     step.message,
				SessionID:   "workflow-session",
				Context:     requestContext,
				ExistingDSL: accumulatedDSL,
				Timestamp:   time.Now(),
			}

			routingResponse, err := router.Route(ctx, routingRequest)
			if err != nil {
				t.Fatalf("Failed to route request: %v", err)
			}

			if routingResponse.DomainName != "hedge-fund-investor" {
				t.Errorf("Expected hedge-fund-investor domain, got %s", routingResponse.DomainName)
			}

			// Generate DSL using the routed domain
			genRequest := &registry.GenerationRequest{
				Instruction: step.message,
				SessionID:   "workflow-session",
				Context:     requestContext,
				ExistingDSL: accumulatedDSL,
				Timestamp:   time.Now(),
			}

			genResponse, err := routingResponse.Domain.GenerateDSL(ctx, genRequest)
			if err != nil {
				t.Fatalf("Failed to generate DSL: %v", err)
			}

			if !strings.Contains(genResponse.DSL, step.expectedVerb) {
				t.Errorf("Expected DSL to contain verb %s, got: %s", step.expectedVerb, genResponse.DSL)
			}

			if !genResponse.IsValid {
				t.Error("Expected generated DSL to be valid")
			}

			// Validate the DSL
			err = hfDomain.ValidateVerbs(genResponse.DSL)
			if err != nil {
				t.Errorf("Generated DSL failed validation: %v", err)
			}

			// Extract context from generated DSL
			extractedContext, err := hfDomain.ExtractContext(genResponse.DSL)
			if err != nil {
				t.Errorf("Failed to extract context from DSL: %v", err)
			}

			// Accumulate DSL and update context for next step
			if accumulatedDSL != "" {
				accumulatedDSL += "\n\n"
			}
			accumulatedDSL += genResponse.DSL

			// Update current context
			for k, v := range extractedContext {
				currentContext[k] = v
			}

			t.Logf("Step %s completed:", step.name)
			t.Logf("  Generated DSL: %s", genResponse.DSL)
			t.Logf("  Confidence: %.2f", genResponse.Confidence)
			t.Logf("  New State: %s", genResponse.ToState)
		})
	}

	// Verify final accumulated DSL
	if accumulatedDSL == "" {
		t.Error("Expected accumulated DSL to be non-empty")
	}

	// Count verbs in accumulated DSL
	verbCount := 0
	for _, verb := range []string{
		"investor.start-opportunity",
		"kyc.begin",
		"kyc.approve",
		"subscribe.request",
	} {
		if strings.Contains(accumulatedDSL, verb) {
			verbCount++
		}
	}

	if verbCount != 4 {
		t.Errorf("Expected 4 verbs in accumulated DSL, found %d", verbCount)
	}

	t.Logf("Final accumulated DSL:\n%s", accumulatedDSL)
}

func TestHedgeFundDomain_StateTransitionValidation(t *testing.T) {
	domain := NewDomain()

	// Test the complete state machine progression
	states := []string{
		"OPPORTUNITY", "PRECHECKS", "KYC_PENDING", "KYC_APPROVED",
		"SUB_PENDING_CASH", "FUNDED_PENDING_NAV", "ISSUED", "ACTIVE",
		"REDEEM_PENDING", "REDEEMED", "OFFBOARDED",
	}

	// Test valid sequential transitions
	for i := 0; i < len(states)-1; i++ {
		from := states[i]
		to := states[i+1]

		err := domain.ValidateStateTransition(from, to)
		if err != nil {
			t.Errorf("Expected valid transition from %s to %s, got error: %v", from, to, err)
		}
	}

	// Test that we can't skip states
	err := domain.ValidateStateTransition("OPPORTUNITY", "KYC_PENDING")
	if err == nil {
		t.Error("Expected error when skipping PRECHECKS state")
	}

	// Test that we can't go backwards
	err = domain.ValidateStateTransition("ACTIVE", "KYC_PENDING")
	if err == nil {
		t.Error("Expected error when going backwards in state machine")
	}

	// Test terminal state
	err = domain.ValidateStateTransition("OFFBOARDED", "ACTIVE")
	if err == nil {
		t.Error("Expected error when transitioning from terminal state")
	}
}

func TestHedgeFundDomain_VerbCategorization(t *testing.T) {
	domain := NewDomain()
	vocab := domain.GetVocabulary()

	// Test that each category contains the expected verbs
	expectedCategoryVerbs := map[string][]string{
		"opportunity": {
			"investor.start-opportunity",
			"investor.record-indication",
		},
		"kyc": {
			"kyc.begin",
			"kyc.collect-doc",
			"kyc.screen",
			"kyc.approve",
			"kyc.refresh-schedule",
		},
		"monitoring": {
			"screen.continuous",
		},
		"tax-banking": {
			"tax.capture",
			"bank.set-instruction",
		},
		"subscription": {
			"subscribe.request",
			"cash.confirm",
			"deal.nav",
			"subscribe.issue",
		},
		"redemption": {
			"redeem.request",
			"redeem.settle",
		},
		"offboarding": {
			"offboard.close",
		},
	}

	for categoryName, expectedVerbs := range expectedCategoryVerbs {
		category, exists := vocab.Categories[categoryName]
		if !exists {
			t.Errorf("Expected category %s not found", categoryName)
			continue
		}

		if len(category.Verbs) != len(expectedVerbs) {
			t.Errorf("Category %s: expected %d verbs, got %d", categoryName, len(expectedVerbs), len(category.Verbs))
			continue
		}

		for _, expectedVerb := range expectedVerbs {
			found := false
			for _, actualVerb := range category.Verbs {
				if actualVerb == expectedVerb {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Category %s missing expected verb %s", categoryName, expectedVerb)
			}
		}

		// Verify each verb in the category actually exists in vocabulary
		for _, verbName := range category.Verbs {
			if _, exists := vocab.Verbs[verbName]; !exists {
				t.Errorf("Category %s references non-existent verb %s", categoryName, verbName)
			}
		}
	}
}

func TestHedgeFundDomain_ArgumentValidation(t *testing.T) {
	domain := NewDomain()
	vocab := domain.GetVocabulary()

	// Test key verbs have proper argument specifications
	tests := []struct {
		verbName     string
		argName      string
		shouldExist  bool
		expectedType registry.ArgumentType
	}{
		{"investor.start-opportunity", "legal-name", true, registry.ArgumentTypeString},
		{"investor.start-opportunity", "type", true, registry.ArgumentTypeEnum},
		{"investor.start-opportunity", "nonexistent", false, ""},
		{"kyc.begin", "investor", true, registry.ArgumentTypeUUID},
		{"kyc.begin", "tier", true, registry.ArgumentTypeEnum},
		{"subscribe.request", "amount", true, registry.ArgumentTypeDecimal},
		{"subscribe.request", "trade-date", true, registry.ArgumentTypeDate},
		{"redeem.request", "percentage", true, registry.ArgumentTypeDecimal},
		{"redeem.request", "units", true, registry.ArgumentTypeDecimal},
	}

	for _, tt := range tests {
		t.Run(tt.verbName+"_"+tt.argName, func(t *testing.T) {
			verb, verbExists := vocab.Verbs[tt.verbName]
			if !verbExists {
				t.Fatalf("Verb %s does not exist", tt.verbName)
			}

			arg, argExists := verb.Arguments[tt.argName]
			if argExists != tt.shouldExist {
				t.Errorf("Argument %s existence: expected %t, got %t", tt.argName, tt.shouldExist, argExists)
				return
			}

			if tt.shouldExist && arg.Type != tt.expectedType {
				t.Errorf("Argument %s type: expected %s, got %s", tt.argName, tt.expectedType, arg.Type)
			}
		})
	}
}
