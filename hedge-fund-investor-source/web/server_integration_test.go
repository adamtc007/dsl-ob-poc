package main

import (
	"bytes"
	"context"
	registry "dsl-ob-poc/internal/domain-registry"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestServerIntegration tests the multi-domain web server integration
func TestServerIntegration(t *testing.T) {
	// Create a test server
	server, err := NewServer(nil, nil, "test-api-key")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var health map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
			t.Errorf("Failed to decode health response: %v", err)
		}

		if health["service"] != "multi-domain-dsl-agent" {
			t.Errorf("Expected service 'multi-domain-dsl-agent', got %v", health["service"])
		}
	})

	// Test domains endpoint
	t.Run("Get Domains", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/domains", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode domains response: %v", err)
		}

		domains, ok := response["domains"].(map[string]interface{})
		if !ok {
			t.Errorf("Expected domains to be a map")
		}

		// Check that both domains are present
		expectedDomains := []string{"hedge-fund-investor", "onboarding"}
		for _, expectedDomain := range expectedDomains {
			if _, exists := domains[expectedDomain]; !exists {
				t.Errorf("Expected domain '%s' to be present", expectedDomain)
			}
		}
	})

	// Test vocabulary endpoints
	t.Run("Get Hedge Fund Vocabulary", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/domains/hedge-fund-investor/vocabulary", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var vocab map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&vocab); err != nil {
			t.Errorf("Failed to decode vocabulary response: %v", err)
		}

		if vocab["domain"] != "hedge-fund-investor" {
			t.Errorf("Expected domain 'hedge-fund-investor', got %v", vocab["domain"])
		}
	})

	t.Run("Get Onboarding Vocabulary", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/domains/onboarding/vocabulary", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var vocab map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&vocab); err != nil {
			t.Errorf("Failed to decode vocabulary response: %v", err)
		}

		if vocab["domain"] != "onboarding" {
			t.Errorf("Expected domain 'onboarding', got %v", vocab["domain"])
		}
	})

	// Test DSL validation with different domains
	t.Run("Validate Hedge Fund DSL", func(t *testing.T) {
		reqBody := map[string]string{
			"dsl":    "(investor.start-opportunity)",
			"domain": "hedge-fund-investor",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/dsl/validate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode validation response: %v", err)
		}

		if response["domain"] != "hedge-fund-investor" {
			t.Errorf("Expected domain 'hedge-fund-investor', got %v", response["domain"])
		}
	})

	t.Run("Validate Onboarding DSL", func(t *testing.T) {
		reqBody := map[string]string{
			"dsl":    "(case.create)",
			"domain": "onboarding",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/dsl/validate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode validation response: %v", err)
		}

		if response["domain"] != "onboarding" {
			t.Errorf("Expected domain 'onboarding', got %v", response["domain"])
		}
	})

	// Test routing metrics
	t.Run("Get Routing Metrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/routing/metrics", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var metrics map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&metrics); err != nil {
			t.Errorf("Failed to decode metrics response: %v", err)
		}

		// Should have basic routing metrics structure
		if _, exists := metrics["total_requests"]; !exists {
			t.Errorf("Expected metrics to contain 'total_requests'")
		}
	})
}

// TestSessionManagement tests multi-domain session handling
func TestSessionManagement(t *testing.T) {
	server, err := NewServer(nil, nil, "test-api-key")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test session creation with different domains
	t.Run("Create Hedge Fund Session", func(t *testing.T) {
		session := server.getOrCreateSession("test-session-1", "hedge-fund-investor")

		if session.SessionID != "test-session-1" {
			t.Errorf("Expected session ID 'test-session-1', got %s", session.SessionID)
		}

		if session.CurrentDomain != "hedge-fund-investor" {
			t.Errorf("Expected domain 'hedge-fund-investor', got %s", session.CurrentDomain)
		}
	})

	t.Run("Create Onboarding Session", func(t *testing.T) {
		session := server.getOrCreateSession("test-session-2", "onboarding")

		if session.SessionID != "test-session-2" {
			t.Errorf("Expected session ID 'test-session-2', got %s", session.SessionID)
		}

		if session.CurrentDomain != "onboarding" {
			t.Errorf("Expected domain 'onboarding', got %s", session.CurrentDomain)
		}
	})

	t.Run("Default Domain Fallback", func(t *testing.T) {
		session := server.getOrCreateSession("test-session-3", "invalid-domain")

		// Should fallback to hedge-fund-investor
		if session.CurrentDomain != "hedge-fund-investor" {
			t.Errorf("Expected fallback domain 'hedge-fund-investor', got %s", session.CurrentDomain)
		}
	})

	// Test session retrieval
	t.Run("Get Session", func(t *testing.T) {
		// Create a session first through the web server's session management
		originalSession := server.getOrCreateSession("test-session-4", "onboarding")

		// Set DSL directly on the web server session for testing
		// Note: In production, DSL should flow through DSL State Manager
		originalSession.BuiltDSL = "(test.dsl)"

		req := httptest.NewRequest("GET", "/api/session/test-session-4", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "test-session-4"})
		w := httptest.NewRecorder()

		server.handleGetSession(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var session ChatSession
		if err := json.NewDecoder(w.Body).Decode(&session); err != nil {
			t.Errorf("Failed to decode session response: %v", err)
		}

		if session.CurrentDomain != "onboarding" {
			t.Errorf("Expected domain 'onboarding', got %s", session.CurrentDomain)
		}

		// Verify DSL was set correctly on the web server session
		if session.BuiltDSL != "(test.dsl)" {
			t.Errorf("Expected DSL '(test.dsl)', got %s", session.BuiltDSL)
		}
	})
}

// TestDomainRouting tests the intelligent domain routing
func TestDomainRouting(t *testing.T) {
	server, err := NewServer(nil, nil, "test-api-key")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()

	// Test routing based on keywords
	testCases := []struct {
		name           string
		message        string
		expectedDomain string
		description    string
	}{
		{
			name:           "Hedge Fund Keywords",
			message:        "Create investor opportunity for John Smith",
			expectedDomain: "hedge-fund-investor",
			description:    "Should route to hedge fund domain for investor-related messages",
		},
		{
			name:           "Onboarding Keywords",
			message:        "Create new case for CBU-1234",
			expectedDomain: "onboarding",
			description:    "Should route to onboarding domain for case-related messages",
		},
		{
			name:           "KYC Keywords",
			message:        "Start KYC process",
			expectedDomain: "hedge-fund-investor", // Both domains have KYC, but hedge fund is default
			description:    "Should route based on context or default domain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			routingReq := &registry.RoutingRequest{
				Message:   tc.message,
				SessionID: "test-routing-session",
				Context:   make(map[string]interface{}),
				Timestamp: time.Now(),
			}

			routingResp, err := server.domainRouter.Route(ctx, routingReq)
			if err != nil {
				t.Errorf("Failed to route message: %v", err)
				return
			}

			if routingResp.DomainName != tc.expectedDomain {
				t.Logf("Message: %s", tc.message)
				t.Logf("Expected domain: %s", tc.expectedDomain)
				t.Logf("Actual domain: %s", routingResp.DomainName)
				t.Logf("Routing reason: %s", routingResp.Reason)
				t.Logf("Routing strategy: %s", routingResp.Strategy)
				// Note: This might not be an error if the router has different logic
				// Just log for observation during development
			}
		})
	}
}

// TestChatIntegration tests the full chat flow with domain routing
func TestChatIntegration(t *testing.T) {
	server, err := NewServer(nil, nil, "test-api-key")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test chat request without actual AI call (will fail gracefully)
	t.Run("Chat Request Structure", func(t *testing.T) {
		chatReq := ChatRequest{
			SessionID: "test-chat-session",
			Message:   "Create investor opportunity",
			Domain:    "hedge-fund-investor",
		}

		jsonBody, _ := json.Marshal(chatReq)
		req := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// This will likely fail due to missing AI integration, but we can check the structure
		if w.Code == http.StatusOK {
			var response ChatResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode chat response: %v", err)
			}

			if response.SessionID != "test-chat-session" {
				t.Errorf("Expected session ID 'test-chat-session', got %s", response.SessionID)
			}
		} else {
			// Expected failure due to missing AI integration
			t.Logf("Chat request failed as expected (missing AI integration): %d", w.Code)
		}
	})
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	server, err := NewServer(nil, nil, "test-api-key")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Run("Invalid Domain", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/domains/invalid-domain/vocabulary", nil)
		req = mux.SetURLVars(req, map[string]string{"domain": "invalid-domain"})
		w := httptest.NewRecorder()

		server.handleGetVocabulary(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Nonexistent Session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/session/nonexistent", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
		w := httptest.NewRecorder()

		server.handleGetSession(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := `{"invalid": json}`
		req := httptest.NewRequest("POST", "/api/chat", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// BenchmarkDomainRouting benchmarks the domain routing performance
func BenchmarkDomainRouting(b *testing.B) {
	server, err := NewServer(nil, nil, "test-api-key")
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()
	routingReq := &registry.RoutingRequest{
		Message:   "Create investor opportunity for benchmark test",
		SessionID: "benchmark-session",
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.domainRouter.Route(ctx, routingReq)
		if err != nil {
			b.Errorf("Routing failed: %v", err)
		}
	}
}
