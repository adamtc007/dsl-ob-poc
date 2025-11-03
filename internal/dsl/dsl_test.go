package dsl

import (
	"strings"
	"testing"

	"dsl-ob-poc/internal/dictionary"
	"dsl-ob-poc/internal/store"
)

func TestCreateCase(t *testing.T) {
	cbuID := "CBU-1234"
	naturePurpose := "UCITS equity fund"

	result := CreateCase(cbuID, naturePurpose)

	if !strings.Contains(result, "(case.create") {
		t.Errorf("Expected DSL to contain '(case.create', got: %s", result)
	}

	if !strings.Contains(result, cbuID) {
		t.Errorf("Expected DSL to contain CBU ID '%s', got: %s", cbuID, result)
	}

	if !strings.Contains(result, naturePurpose) {
		t.Errorf("Expected DSL to contain nature-purpose '%s', got: %s", naturePurpose, result)
	}

	// Verify it's valid S-expression format
	if !strings.HasPrefix(result, "(") || !strings.HasSuffix(result, ")") {
		t.Errorf("Expected DSL to be wrapped in parentheses, got: %s", result)
	}
}

func TestAddProducts(t *testing.T) {
	currentDSL := `(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "Test fund")
)`

	products := []*store.Product{
		{ProductID: "p1", Name: "CUSTODY", Description: "Custody services"},
		{ProductID: "p2", Name: "FUND_ACCOUNTING", Description: "Fund accounting"},
	}

	result, err := AddProducts(currentDSL, products)
	if err != nil {
		t.Fatalf("AddProducts failed: %v", err)
	}

	if !strings.Contains(result, "(products.add") {
		t.Errorf("Expected DSL to contain '(products.add', got: %s", result)
	}

	if !strings.Contains(result, "CUSTODY") {
		t.Errorf("Expected DSL to contain 'CUSTODY', got: %s", result)
	}

	if !strings.Contains(result, "FUND_ACCOUNTING") {
		t.Errorf("Expected DSL to contain 'FUND_ACCOUNTING', got: %s", result)
	}

	// Original DSL should still be present
	if !strings.Contains(result, "(case.create") {
		t.Errorf("Expected DSL to preserve original case.create block, got: %s", result)
	}
}

func TestAddProductsEmpty(t *testing.T) {
	currentDSL := "(case.create)"
	var products []*store.Product

	result, err := AddProducts(currentDSL, products)
	if err != nil {
		t.Fatalf("AddProducts with empty list should not error: %v", err)
	}

	if result != currentDSL {
		t.Errorf("Expected DSL unchanged with empty products, got: %s", result)
	}
}

func TestParseProductNames(t *testing.T) {
	dsl := `(case.create
  (cbu.id "CBU-1234")
)

(products.add "CUSTODY" "FUND_ACCOUNTING" "TRANSFER_AGENCY")`

	names, err := ParseProductNames(dsl)
	if err != nil {
		t.Fatalf("ParseProductNames failed: %v", err)
	}

	expectedCount := 3
	if len(names) != expectedCount {
		t.Errorf("Expected %d product names, got %d: %v", expectedCount, len(names), names)
	}

	expectedNames := map[string]bool{
		"CUSTODY":         true,
		"FUND_ACCOUNTING": true,
		"TRANSFER_AGENCY": true,
	}

	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("Unexpected product name: %s", name)
		}
	}
}

func TestParseProductNamesNoBlock(t *testing.T) {
	dsl := "(case.create)"

	_, err := ParseProductNames(dsl)
	if err == nil {
		t.Error("Expected error when no products.add block found")
	}
}

func TestAddDiscoveredServices(t *testing.T) {
	currentDSL := `(case.create
  (cbu.id "CBU-1234")
)

(products.add "CUSTODY")`

	plan := ServiceDiscoveryPlan{
		ProductServices: map[string][]store.Service{
			"CUSTODY": {
				{ServiceID: "s1", Name: "CustodyService", Description: "Custody"},
				{ServiceID: "s2", Name: "SettlementService", Description: "Settlement"},
			},
		},
	}

	result, err := AddDiscoveredServices(currentDSL, plan)
	if err != nil {
		t.Fatalf("AddDiscoveredServices failed: %v", err)
	}

	if !strings.Contains(result, "(services.discover") {
		t.Errorf("Expected DSL to contain '(services.discover', got: %s", result)
	}

	if !strings.Contains(result, "(for.product \"CUSTODY\"") {
		t.Errorf("Expected DSL to contain product block, got: %s", result)
	}

	if !strings.Contains(result, "CustodyService") {
		t.Errorf("Expected DSL to contain 'CustodyService', got: %s", result)
	}

	if !strings.Contains(result, "SettlementService") {
		t.Errorf("Expected DSL to contain 'SettlementService', got: %s", result)
	}
}

func TestParseServiceNames(t *testing.T) {
	dsl := `(services.discover
  (for.product "CUSTODY"
    (service "CustodyService")
    (service "SettlementService")
  )
  (for.product "FUND_ACCOUNTING"
    (service "FundAccountingService")
  )
)`

	names, err := ParseServiceNames(dsl)
	if err != nil {
		t.Fatalf("ParseServiceNames failed: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Expected 3 service names, got %d: %v", len(names), names)
	}

	expectedNames := map[string]bool{
		"CustodyService":        true,
		"SettlementService":     true,
		"FundAccountingService": true,
	}

	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("Unexpected service name: %s", name)
		}
	}
}

func TestParseServiceNamesNoBlock(t *testing.T) {
	dsl := "(case.create)"

	_, err := ParseServiceNames(dsl)
	if err == nil {
		t.Error("Expected error when no service blocks found")
	}
}

func TestAddDiscoveredResources(t *testing.T) {
	currentDSL := `(case.create)

(services.discover
  (for.product "CUSTODY"
    (service "CustodyService")
  )
)`

	plan := ResourceDiscoveryPlan{
		ServiceResources: map[string][]store.ProdResource{
			"CustodyService": {
				{
					ResourceID:      "r1",
					Name:            "CustodyAccount",
					Owner:           "CustodyTech",
					DictionaryGroup: "CustodyAccount",
				},
			},
		},
		ResourceAttributes: map[string][]dictionary.Attribute{
			"CustodyAccount": {
				{
					AttributeID:     "a1",
					Name:            "custody.account_number",
					LongDescription: "Custody account identifier",
					GroupID:         "CustodyAccount",
					Mask:            "string",
					Domain:          "Custody",
				},
			},
		},
	}

	result, err := AddDiscoveredResources(currentDSL, plan)
	if err != nil {
		t.Fatalf("AddDiscoveredResources failed: %v", err)
	}

	if !strings.Contains(result, "(resources.plan") {
		t.Errorf("Expected DSL to contain '(resources.plan', got: %s", result)
	}

	if !strings.Contains(result, "(resource.create \"CustodyAccount\"") {
		t.Errorf("Expected DSL to contain resource block, got: %s", result)
	}

	if !strings.Contains(result, "(owner \"CustodyTech\")") {
		t.Errorf("Expected DSL to contain owner, got: %s", result)
	}

	if !strings.Contains(result, "(attr.\"custody.account_number\")") {
		t.Errorf("Expected DSL to contain attribute, got: %s", result)
	}
}

func TestAddDiscoveredResourcesMultiple(t *testing.T) {
	currentDSL := "(case.create)"

	plan := ResourceDiscoveryPlan{
		ServiceResources: map[string][]store.ProdResource{
			"CustodyService": {
				{ResourceID: "r1", Name: "CustodyAccount", Owner: "CustodyTech", DictionaryGroup: "CustodyAccount"},
			},
			"AccountingService": {
				{ResourceID: "r2", Name: "AccountingRecord", Owner: "AcctTech", DictionaryGroup: "FundAccounting"},
			},
		},
		ResourceAttributes: map[string][]dictionary.Attribute{
			"CustodyAccount": {{AttributeID: "a1", Name: "custody.account_number", GroupID: "CustodyAccount"}},
			"FundAccounting": {{AttributeID: "a2", Name: "accounting.nav_value", GroupID: "FundAccounting"}},
		},
	}

	result, err := AddDiscoveredResources(currentDSL, plan)
	if err != nil {
		t.Fatalf("AddDiscoveredResources failed: %v", err)
	}

	// Should contain both resources
	if !strings.Contains(result, "CustodyAccount") {
		t.Errorf("Expected DSL to contain 'CustodyAccount', got: %s", result)
	}

	if !strings.Contains(result, "AccountingRecord") {
		t.Errorf("Expected DSL to contain 'AccountingRecord', got: %s", result)
	}

	if !strings.Contains(result, "account_number") {
		t.Errorf("Expected DSL to contain 'account_number', got: %s", result)
	}

	if !strings.Contains(result, "nav_value") {
		t.Errorf("Expected DSL to contain 'nav_value', got: %s", result)
	}
}
