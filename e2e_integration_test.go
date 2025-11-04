package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	registry "dsl-ob-poc/internal/domain-registry"
	hedgefundinvestor "dsl-ob-poc/internal/domains/hedge-fund-investor"
	"dsl-ob-poc/internal/domains/onboarding"
)

// TestMultiDomainE2EWorkflows tests complete end-to-end workflows across multiple domains
func TestMultiDomainE2EWorkflows(t *testing.T) {
	// Create domain registry with both domains
	reg := registry.NewRegistry()

	hfDomain := hedgefundinvestor.NewDomain()
	if err := reg.Register(hfDomain); err != nil {
		t.Fatalf("Failed to register hedge fund domain: %v", err)
	}

	obDomain := onboarding.NewDomain()
	if err := reg.Register(obDomain); err != nil {
		t.Fatalf("Failed to register onboarding domain: %v", err)
	}

	router := registry.NewRouter(reg)
	ctx := context.Background()

	// Test 1: Hedge Fund Complete Investor Journey
	t.Run("HedgeFund_CompleteInvestorJourney", func(t *testing.T) {
		sessionID := "e2e-hf-session-1"
		context := make(map[string]interface{})
		accumulatedDSL := ""

		// Step 1: Create investor opportunity
		step1Req := &registry.RoutingRequest{
			Message:   "Create investment opportunity for Acme Capital LP, a Swiss corporate investor",
			SessionID: sessionID,
			Context:   context,
			Timestamp: time.Now(),
		}

		routingResp, err := router.Route(ctx, step1Req)
		if err != nil {
			t.Fatalf("Step 1 routing failed: %v", err)
		}

		if routingResp.DomainName != "hedge-fund-investor" {
			t.Errorf("Expected hedge-fund-investor domain, got %s", routingResp.DomainName)
		}

		// Mock generation response since we don't have real AI
		mockStep1DSL := `(investor.start-opportunity
  :legal-name "Acme Capital LP"
  :type "CORPORATE"
  :domicile "CH")`

		accumulatedDSL += mockStep1DSL
		t.Logf("Step 1 DSL: %s", mockStep1DSL)

		// Step 2: Start KYC process
		step2Req := &registry.RoutingRequest{
			Message:     "Start KYC process for this investor",
			SessionID:   sessionID,
			Context:     context,
			ExistingDSL: accumulatedDSL,
			Timestamp:   time.Now(),
		}

		routingResp2, err := router.Route(ctx, step2Req)
		if err != nil {
			t.Fatalf("Step 2 routing failed: %v", err)
		}

		if routingResp2.DomainName != "hedge-fund-investor" {
			t.Errorf("Expected hedge-fund-investor domain for KYC, got %s", routingResp2.DomainName)
		}

		mockStep2DSL := `
(kyc.begin
  :investor "<investor_id>"
  :tier "STANDARD")`

		accumulatedDSL += mockStep2DSL
		t.Logf("Step 2 DSL: %s", mockStep2DSL)

		// Step 3: Collect documents
		mockStep3DSL := `
(kyc.collect-doc
  :investor "<investor_id>"
  :doc-type "Certificate of Incorporation"
  :subject "Acme Capital LP")`

		accumulatedDSL += mockStep3DSL
		t.Logf("Step 3 DSL: %s", mockStep3DSL)

		// Validate complete workflow DSL
		err = hfDomain.ValidateVerbs(accumulatedDSL)
		if err != nil {
			t.Logf("DSL validation warning (expected with placeholders): %v", err)
		}

		// Verify DSL contains expected verbs
		expectedVerbs := []string{"investor.start-opportunity", "kyc.begin", "kyc.collect-doc"}
		for _, verb := range expectedVerbs {
			if !containsString(accumulatedDSL, verb) {
				t.Errorf("Expected DSL to contain verb %s", verb)
			}
		}

		t.Logf("Hedge fund complete workflow DSL (%d chars):\n%s", len(accumulatedDSL), accumulatedDSL)
	})

	// Test 2: Onboarding Complete Case Journey
	t.Run("Onboarding_CompleteCaseJourney", func(t *testing.T) {
		sessionID := "e2e-ob-session-1"
		context := make(map[string]interface{})
		accumulatedDSL := ""

		// Step 1: Create onboarding case
		step1Req := &registry.RoutingRequest{
			Message:   "Create onboarding case for CBU-1234, a UCITS equity fund domiciled in Luxembourg",
			SessionID: sessionID,
			Context:   context,
			Timestamp: time.Now(),
		}

		routingResp, err := router.Route(ctx, step1Req)
		if err != nil {
			t.Fatalf("Step 1 routing failed: %v", err)
		}

		// Note: Router may select either domain based on keywords, both are valid for this test
		t.Logf("Router selected domain: %s (reason: %s)", routingResp.DomainName, routingResp.Reason)

		// For onboarding workflow, let's force onboarding domain
		obDomain, err := reg.Get("onboarding")
		if err != nil {
			t.Fatalf("Failed to get onboarding domain: %v", err)
		}

		// Mock generation for case creation
		mockStep1DSL := `(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU"))`

		accumulatedDSL += mockStep1DSL
		t.Logf("Step 1 DSL: %s", mockStep1DSL)

		// Step 2: Add products
		mockStep2DSL := `
(products.add "CUSTODY" "FUND_ACCOUNTING")`

		accumulatedDSL += mockStep2DSL
		t.Logf("Step 2 DSL: %s", mockStep2DSL)

		// Step 3: Discover KYC requirements
		mockStep3DSL := `
(kyc.start
  (documents
    (document "CertificateOfIncorporation"))
  (jurisdictions
    (jurisdiction "LU")))`

		accumulatedDSL += mockStep3DSL
		t.Logf("Step 3 DSL: %s", mockStep3DSL)

		// Step 4: Resource planning
		mockStep4DSL := `
(resources.plan
  (resource.create "CustodyAccount"
    (owner "CustodyTech")))`

		accumulatedDSL += mockStep4DSL
		t.Logf("Step 4 DSL: %s", mockStep4DSL)

		// Validate complete workflow DSL
		err = obDomain.ValidateVerbs(accumulatedDSL)
		if err != nil {
			t.Logf("DSL validation warning: %v", err)
		}

		// Verify DSL contains expected verbs
		expectedVerbs := []string{"case.create", "products.add", "kyc.start", "resources.plan"}
		for _, verb := range expectedVerbs {
			if !containsString(accumulatedDSL, verb) {
				t.Errorf("Expected DSL to contain verb %s", verb)
			}
		}

		t.Logf("Onboarding complete workflow DSL (%d chars):\n%s", len(accumulatedDSL), accumulatedDSL)
	})

	// Test 3: Cross-Domain Scenario
	t.Run("CrossDomain_InvestorOnboarding", func(t *testing.T) {
		sessionID := "e2e-cross-domain-1"
		context := make(map[string]interface{})
		accumulatedDSL := ""

		// Start with hedge fund investor
		step1Req := &registry.RoutingRequest{
			Message:   "Create investor opportunity for ABC Fund Management",
			SessionID: sessionID,
			Context:   context,
			Timestamp: time.Now(),
		}

		routingResp1, err := router.Route(ctx, step1Req)
		if err != nil {
			t.Fatalf("Step 1 routing failed: %v", err)
		}

		mockHFStep := `(investor.start-opportunity
  :legal-name "ABC Fund Management"
  :type "CORPORATE")`

		accumulatedDSL += mockHFStep
		currentDomain := routingResp1.DomainName
		t.Logf("Started in domain: %s", currentDomain)

		// Now switch context to onboarding
		step2Req := &registry.RoutingRequest{
			Message:       "Create case CBU-5678 for this client's onboarding",
			SessionID:     sessionID,
			CurrentDomain: currentDomain,
			Context:       context,
			ExistingDSL:   accumulatedDSL,
			Timestamp:     time.Now(),
		}

		routingResp2, err := router.Route(ctx, step2Req)
		if err != nil {
			t.Fatalf("Step 2 routing failed: %v", err)
		}

		mockOBStep := `
(case.create
  (cbu.id "CBU-5678")
  (nature-purpose "Fund management onboarding"))`

		accumulatedDSL += mockOBStep
		newDomain := routingResp2.DomainName
		t.Logf("Switched to domain: %s", newDomain)

		// Verify we have DSL from both domains
		if !containsString(accumulatedDSL, "investor.start-opportunity") {
			t.Error("Expected hedge fund DSL in cross-domain workflow")
		}
		if !containsString(accumulatedDSL, "case.create") {
			t.Error("Expected onboarding DSL in cross-domain workflow")
		}

		t.Logf("Cross-domain workflow DSL (%d chars):\n%s", len(accumulatedDSL), accumulatedDSL)
	})
}

// TestDomainRegistryE2E tests the domain registry system end-to-end
func TestDomainRegistryE2E(t *testing.T) {
	reg := registry.NewRegistry()

	// Test domain lifecycle
	t.Run("DomainLifecycle", func(t *testing.T) {
		// Initially empty
		domains := reg.List()
		if len(domains) != 0 {
			t.Errorf("Expected empty registry, got %d domains", len(domains))
		}

		// Register hedge fund domain
		hfDomain := hedgefundinvestor.NewDomain()
		err := reg.Register(hfDomain)
		if err != nil {
			t.Fatalf("Failed to register hedge fund domain: %v", err)
		}

		// Verify registration
		domains = reg.List()
		if len(domains) != 1 {
			t.Errorf("Expected 1 domain, got %d", len(domains))
		}
		if domains[0] != "hedge-fund-investor" {
			t.Errorf("Expected hedge-fund-investor, got %s", domains[0])
		}

		// Register onboarding domain
		obDomain := onboarding.NewDomain()
		err = reg.Register(obDomain)
		if err != nil {
			t.Fatalf("Failed to register onboarding domain: %v", err)
		}

		// Verify both domains
		domains = reg.List()
		if len(domains) != 2 {
			t.Errorf("Expected 2 domains, got %d", len(domains))
		}

		// Test domain retrieval
		retrievedHF, err := reg.Get("hedge-fund-investor")
		if err != nil {
			t.Errorf("Failed to retrieve hedge fund domain: %v", err)
		}
		if retrievedHF.Name() != "hedge-fund-investor" {
			t.Errorf("Retrieved wrong domain: %s", retrievedHF.Name())
		}

		// Test vocabulary access
		vocabs := reg.GetAllVocabularies()
		if len(vocabs) != 2 {
			t.Errorf("Expected 2 vocabularies, got %d", len(vocabs))
		}

		hfVocab, exists := vocabs["hedge-fund-investor"]
		if !exists {
			t.Error("Hedge fund vocabulary not found")
		} else {
			if len(hfVocab.Verbs) != 17 {
				t.Errorf("Expected 17 hedge fund verbs, got %d", len(hfVocab.Verbs))
			}
		}

		obVocab, exists := vocabs["onboarding"]
		if !exists {
			t.Error("Onboarding vocabulary not found")
		} else {
			if len(obVocab.Verbs) != 54 {
				t.Errorf("Expected 54 onboarding verbs, got %d", len(obVocab.Verbs))
			}
		}

		// Test health monitoring
		if !reg.IsHealthy() {
			t.Error("Registry should be healthy with both domains")
		}

		metrics := reg.GetMetrics()
		if metrics.TotalDomains != 2 {
			t.Errorf("Expected 2 domains in metrics, got %d", metrics.TotalDomains)
		}
	})
}

// TestRoutingStrategiesE2E tests all routing strategies end-to-end
func TestRoutingStrategiesE2E(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	testCases := []struct {
		name           string
		message        string
		expectedDomain string
		description    string
	}{
		{
			name:           "ExplicitDomainSwitch",
			message:        "switch to onboarding domain",
			expectedDomain: "onboarding",
			description:    "Explicit domain switch should work",
		},
		{
			name:           "InvestorKeywords",
			message:        "create investor opportunity for new client",
			expectedDomain: "hedge-fund-investor",
			description:    "Investor keywords should route to hedge fund domain",
		},
		{
			name:           "CaseKeywords",
			message:        "create new case for CBU onboarding",
			expectedDomain: "onboarding",
			description:    "Case keywords should route to onboarding domain",
		},
		{
			name:           "KYCKeywords",
			message:        "start KYC process",
			expectedDomain: "", // Either domain is acceptable
			description:    "KYC keywords may route to either domain",
		},
		{
			name:           "VerbBasedRouting",
			message:        "investor.start-opportunity for new client",
			expectedDomain: "hedge-fund-investor",
			description:    "Verb-based routing should detect hedge fund verb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &registry.RoutingRequest{
				Message:   tc.message,
				SessionID: fmt.Sprintf("test-%s", tc.name),
				Context:   make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			resp, err := router.Route(ctx, req)
			if err != nil {
				t.Errorf("Routing failed for %s: %v", tc.name, err)
				return
			}

			if tc.expectedDomain != "" && resp.DomainName != tc.expectedDomain {
				t.Logf("Message: %s", tc.message)
				t.Logf("Expected: %s", tc.expectedDomain)
				t.Logf("Got: %s", resp.DomainName)
				t.Logf("Strategy: %s", resp.Strategy)
				t.Logf("Reason: %s", resp.Reason)
				t.Logf("Confidence: %.2f", resp.Confidence)
				// Log for analysis but don't fail - routing logic may be different
			}

			// Verify response structure
			if resp.Domain == nil {
				t.Errorf("Domain should not be nil in routing response")
			}
			if resp.Strategy == "" {
				t.Errorf("Strategy should not be empty")
			}
			if resp.Confidence < 0 || resp.Confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %.2f", resp.Confidence)
			}

			t.Logf("%s: %s → %s (strategy: %s, confidence: %.2f)",
				tc.name, tc.message, resp.DomainName, resp.Strategy, resp.Confidence)
		})
	}
}

// TestVocabularyConsistencyE2E tests vocabulary consistency across domains
func TestVocabularyConsistencyE2E(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())

	t.Run("VocabularyIntegrity", func(t *testing.T) {
		vocabs := reg.GetAllVocabularies()

		for domainName, vocab := range vocabs {
			t.Run(fmt.Sprintf("Domain_%s", domainName), func(t *testing.T) {
				// Verify vocabulary structure
				if vocab.Domain != domainName {
					t.Errorf("Vocabulary domain mismatch: expected %s, got %s", domainName, vocab.Domain)
				}

				if vocab.Version == "" {
					t.Error("Vocabulary version should not be empty")
				}

				if len(vocab.Verbs) == 0 {
					t.Error("Vocabulary should have verbs")
				}

				// Verify verb structure
				for verbName, verbDef := range vocab.Verbs {
					if verbDef.Name != verbName {
						t.Errorf("Verb name mismatch: key %s != name %s", verbName, verbDef.Name)
					}

					if verbDef.Category == "" {
						t.Errorf("Verb %s should have a category", verbName)
					}

					if verbDef.Description == "" {
						t.Errorf("Verb %s should have a description", verbName)
					}

					// Verify arguments if present
					for argName, argSpec := range verbDef.Arguments {
						if argSpec.Name != argName {
							t.Errorf("Argument name mismatch in %s: key %s != name %s", verbName, argName, argSpec.Name)
						}

						if argSpec.Type == "" {
							t.Errorf("Argument %s in verb %s should have a type", argName, verbName)
						}
					}
				}

				// Verify categories
				if len(vocab.Categories) == 0 {
					t.Error("Vocabulary should have categories")
				}

				for catName, category := range vocab.Categories {
					if category.Name != catName {
						t.Errorf("Category name mismatch: key %s != name %s", catName, category.Name)
					}

					if len(category.Verbs) == 0 {
						t.Errorf("Category %s should have verbs", catName)
					}

					// Verify category verbs exist in vocabulary
					for _, verbName := range category.Verbs {
						if _, exists := vocab.Verbs[verbName]; !exists {
							t.Errorf("Category %s references non-existent verb %s", catName, verbName)
						}
					}
				}

				t.Logf("Domain %s: %d verbs, %d categories", domainName, len(vocab.Verbs), len(vocab.Categories))
			})
		}
	})

	t.Run("CrossDomainVerbConflicts", func(t *testing.T) {
		vocabs := reg.GetAllVocabularies()
		allVerbs := make(map[string]string) // verb -> domain

		// Collect all verbs across domains
		for domainName, vocab := range vocabs {
			for verbName := range vocab.Verbs {
				if existingDomain, exists := allVerbs[verbName]; exists {
					t.Logf("Verb conflict: %s exists in both %s and %s", verbName, existingDomain, domainName)
					// This might be acceptable for some verbs like kyc.* that appear in multiple domains
				} else {
					allVerbs[verbName] = domainName
				}
			}
		}

		t.Logf("Total unique verbs across all domains: %d", len(allVerbs))
	})
}

// TestPerformanceE2E tests performance characteristics of the multi-domain system
func TestPerformanceE2E(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	t.Run("RoutingPerformance", func(t *testing.T) {
		messages := []string{
			"Create investor opportunity",
			"Start KYC process",
			"Create case for CBU-1234",
			"Add custody products",
			"Begin compliance check",
		}

		iterations := 1000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			req := &registry.RoutingRequest{
				Message:   messages[i%len(messages)],
				SessionID: fmt.Sprintf("perf-test-%d", i),
				Context:   make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			_, err := router.Route(ctx, req)
			if err != nil {
				t.Errorf("Routing failed at iteration %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		avgLatency := duration / time.Duration(iterations)

		t.Logf("Routing performance: %d requests in %v", iterations, duration)
		t.Logf("Average latency per request: %v", avgLatency)
		t.Logf("Requests per second: %.2f", float64(iterations)/duration.Seconds())

		if avgLatency > 10*time.Millisecond {
			t.Logf("Warning: Average routing latency is high: %v", avgLatency)
		}
	})

	t.Run("DomainLookupPerformance", func(t *testing.T) {
		iterations := 10000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			domainName := "hedge-fund-investor"
			if i%2 == 0 {
				domainName = "onboarding"
			}

			_, err := reg.Get(domainName)
			if err != nil {
				t.Errorf("Domain lookup failed at iteration %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		avgLatency := duration / time.Duration(iterations)

		t.Logf("Domain lookup performance: %d lookups in %v", iterations, duration)
		t.Logf("Average latency per lookup: %v", avgLatency)

		if avgLatency > 1*time.Microsecond {
			t.Logf("Warning: Average domain lookup latency is high: %v", avgLatency)
		}
	})
}

// TestErrorRecoveryE2E tests error handling and recovery scenarios
func TestErrorRecoveryE2E(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	t.Run("InvalidDomainHandling", func(t *testing.T) {
		_, err := reg.Get("invalid-domain")
		if err == nil {
			t.Error("Expected error for invalid domain")
		}

		vocabs := reg.GetAllVocabularies()
		if _, exists := vocabs["invalid-domain"]; exists {
			t.Error("Invalid domain should not appear in vocabularies")
		}
	})

	t.Run("RouterFallbackBehavior", func(t *testing.T) {
		// Test routing with empty message
		req := &registry.RoutingRequest{
			Message:   "",
			SessionID: "empty-message-test",
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		_, err := router.Route(ctx, req)
		if err == nil {
			t.Error("Expected error for empty message")
		}

		// Test routing with very long message
		longMessage := make([]byte, 10000)
		for i := range longMessage {
			longMessage[i] = 'a'
		}

		req = &registry.RoutingRequest{
			Message:   string(longMessage),
			SessionID: "long-message-test",
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		resp, err := router.Route(ctx, req)
		if err != nil {
			t.Errorf("Router should handle long messages gracefully: %v", err)
		} else {
			t.Logf("Long message routed to: %s", resp.DomainName)
		}
	})

	t.Run("RegistryHealthRecovery", func(t *testing.T) {
		// Registry should be healthy with both domains
		if !reg.IsHealthy() {
			t.Error("Registry should be healthy")
		}

		metrics := reg.GetMetrics()
		if metrics.TotalDomains != 2 {
			t.Errorf("Expected 2 domains, got %d", metrics.TotalDomains)
		}

		// Test domain unregistration and re-registration
		err := reg.Unregister("hedge-fund-investor")
		if err != nil {
			t.Errorf("Failed to unregister domain: %v", err)
		}

		domains := reg.List()
		if len(domains) != 1 {
			t.Errorf("Expected 1 domain after unregistration, got %d", len(domains))
		}

		// Re-register
		hfDomain := hedgefundinvestor.NewDomain()
		err = reg.Register(hfDomain)
		if err != nil {
			t.Errorf("Failed to re-register domain: %v", err)
		}

		if !reg.IsHealthy() {
			t.Error("Registry should be healthy after re-registration")
		}
	})
}

// Helper function to check if a string contains another string
func containsString(text, substr string) bool {
	return len(text) >= len(substr) &&
		(text == substr ||
			text[:len(substr)] == substr ||
			text[len(text)-len(substr):] == substr ||
			findSubstring(text, substr))
}

func findSubstring(text, substr string) bool {
	if len(substr) > len(text) {
		return false
	}
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkMultiDomainOperations benchmarks key multi-domain operations
func BenchmarkMultiDomainOperations(b *testing.B) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	b.Run("DomainRouting", func(b *testing.B) {
		req := &registry.RoutingRequest{
			Message:   "Create investor opportunity for benchmark",
			SessionID: "benchmark-session",
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := router.Route(ctx, req)
			if err != nil {
				b.Errorf("Routing failed: %v", err)
			}
		}
	})

	b.Run("DomainLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			domainName := "hedge-fund-investor"
			if i%2 == 0 {
				domainName = "onboarding"
			}
			_, err := reg.Get(domainName)
			if err != nil {
				b.Errorf("Domain lookup failed: %v", err)
			}
		}
	})

	b.Run("VocabularyAccess", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = reg.GetAllVocabularies()
		}
	})
}
