package mocks

import (
	"context"
	"testing"
)

func TestMockStore_DisconnectedOperation(t *testing.T) {
	// Create mock store pointing to our JSON mock data
	mockStore := NewMockStore("../../data/mocks")
	defer mockStore.Close()

	ctx := context.Background()

	// Test CBU operations
	cbus, err := mockStore.ListCBUs(ctx)
	if err != nil {
		t.Fatalf("Failed to list CBUs: %v", err)
	}

	if len(cbus) == 0 {
		t.Error("Expected mock CBUs, got none")
	}

	// Test getting a specific CBU
	cbu, err := mockStore.GetCBUByName(ctx, "CBU-1234")
	if err != nil {
		t.Fatalf("Failed to get CBU by name: %v", err)
	}

	if cbu.Name != "CBU-1234" {
		t.Errorf("Expected CBU name 'CBU-1234', got '%s'", cbu.Name)
	}

	// Test role operations
	roles, err := mockStore.ListRoles(ctx)
	if err != nil {
		t.Fatalf("Failed to list roles: %v", err)
	}

	if len(roles) == 0 {
		t.Error("Expected mock roles, got none")
	}

	// Test product operations
	products, err := mockStore.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	if len(products) == 0 {
		t.Error("Expected mock products, got none")
	}

	// Test service discovery
	services, err := mockStore.GetServicesForProducts(ctx, []string{"CUSTODY", "FUND_ACCOUNTING"})
	if err != nil {
		t.Fatalf("Failed to get services for products: %v", err)
	}

	if len(services) == 0 {
		t.Error("Expected services for products, got none")
	}

	// Test dictionary operations
	attr, err := mockStore.GetDictionaryAttributeByName(ctx, "onboard.cbu_id")
	if err != nil {
		t.Fatalf("Failed to get dictionary attribute: %v", err)
	}

	if attr.Name != "onboard.cbu_id" {
		t.Errorf("Expected attribute name 'onboard.cbu_id', got '%s'", attr.Name)
	}

	// Test DSL operations
	dsl, err := mockStore.GetLatestDSL(ctx, "CBU-1234")
	if err != nil {
		t.Fatalf("Failed to get latest DSL: %v", err)
	}

	if dsl == "" {
		t.Error("Expected DSL content, got empty string")
	}

	// Test attribute value resolution
	value, source, state, err := mockStore.ResolveValueFor(ctx, "123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Failed to resolve value: %v", err)
	}

	if state != "resolved" && state != "pending" {
		t.Errorf("Expected state 'resolved' or 'pending', got '%s'", state)
	}

	if source == nil {
		t.Error("Expected source metadata, got nil")
	}

	if value == nil {
		t.Error("Expected value, got nil")
	}

	t.Logf("Mock store test completed successfully!")
	t.Logf("CBUs: %d, Roles: %d, Products: %d, Services: %d", len(cbus), len(roles), len(products), len(services))
	t.Logf("Latest DSL length: %d characters", len(dsl))
	t.Logf("Resolved value: %s, state: %s", string(value), state)
}