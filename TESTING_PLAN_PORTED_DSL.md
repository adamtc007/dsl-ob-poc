# Testing Plan for Ported DSL (Multi-Domain Migration)

## Executive Summary

This document outlines the comprehensive testing strategy for porting the existing onboarding DSL into the new multi-domain architecture. The goal is to ensure **zero functional regression** while validating the new shared infrastructure and domain isolation patterns.

---

## Testing Philosophy

### Core Principles
1. **Preserve Existing Functionality** - All existing onboarding DSL operations must work identically
2. **Validate Shared Infrastructure** - Parser, dictionary, session management must be domain-agnostic
3. **Verify Domain Isolation** - Each domain's verbs and state machines are independent
4. **Test Cross-Domain Composition** - Onboarding can orchestrate hedge fund operations
5. **Performance Parity** - No performance degradation from single-domain implementation

---

## Test Coverage Matrix

| Test Category | Existing Tests | New Tests Needed | Priority |
|--------------|----------------|------------------|----------|
| DSL Vocabulary | ✅ 8 tests | 🆕 12 tests | P0 |
| DSL Generation | ✅ 15 tests | 🆕 20 tests | P0 |
| Shared Parser | ❌ None | 🆕 25 tests | P0 |
| Dictionary Service | ✅ 6 tests | 🆕 10 tests | P0 |
| Session Management | ❌ None | 🆕 15 tests | P0 |
| Domain Registry | ❌ None | 🆕 18 tests | P0 |
| Verb Validation | ✅ 7 tests | 🆕 15 tests | P0 |
| Cross-Domain | ❌ None | 🆕 30 tests | P1 |
| Integration | ❌ None | 🆕 40 tests | P1 |
| Performance | ❌ None | 🆕 10 tests | P2 |

**Total**: 36 existing tests → 201 total tests after migration

---

## Phase 1: Shared Infrastructure Tests (Week 1)

### 1.1 DSL Parser Tests (`internal/shared-dsl/parser/parser_test.go`)

**New Test File** - 25 test cases

```go
package parser

import "testing"

// Basic Parsing Tests (5 tests)
func TestParse_SimpleVerb(t *testing.T)
func TestParse_NestedExpressions(t *testing.T)
func TestParse_MultipleTopLevel(t *testing.T)
func TestParse_EmptyDSL(t *testing.T)
func TestParse_MalformedSyntax(t *testing.T)

// Onboarding DSL Tests (10 tests)
func TestParse_OnboardingCaseCreate(t *testing.T)
func TestParse_OnboardingProductsAdd(t *testing.T)
func TestParse_OnboardingKYCStart(t *testing.T)
func TestParse_OnboardingServicesDiscover(t *testing.T)
func TestParse_OnboardingResourcesPlan(t *testing.T)
func TestParse_OnboardingValuesBinds(t *testing.T)
func TestParse_OnboardingCompleteWorkflow(t *testing.T)
func TestParse_OnboardingWithAttributes(t *testing.T)
func TestParse_OnboardingMultiProduct(t *testing.T)
func TestParse_OnboardingNestedResources(t *testing.T)

// Hedge Fund DSL Tests (5 tests)
func TestParse_HedgeFundInvestorStart(t *testing.T)
func TestParse_HedgeFundKYCBegin(t *testing.T)
func TestParse_HedgeFundSubscription(t *testing.T)
func TestParse_HedgeFundRedemption(t *testing.T)
func TestParse_HedgeFundCompleteWorkflow(t *testing.T)

// Cross-Domain Parsing (3 tests)
func TestParse_MixedDomainDSL(t *testing.T)
func TestParse_OnboardingCallsHedgeFund(t *testing.T)
func TestParse_LargeDSLDocument(t *testing.T)

// AST Validation (2 tests)
func TestAST_NodeTraversal(t *testing.T)
func TestAST_VerbExtraction(t *testing.T)
```

**Sample Test Implementation**:
```go
func TestParse_OnboardingCaseCreate(t *testing.T) {
    dsl := `(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund")
)`
    
    ast, err := Parse(dsl)
    if err != nil {
        t.Fatalf("Parse failed: %v", err)
    }
    
    if ast.Root == nil {
        t.Fatal("Expected non-nil root node")
    }
    
    if ast.Root.Type != VerbNode {
        t.Errorf("Expected VerbNode, got %v", ast.Root.Type)
    }
    
    if ast.Root.Value != "case.create" {
        t.Errorf("Expected verb 'case.create', got '%s'", ast.Root.Value)
    }
    
    if len(ast.Root.Children) != 2 {
        t.Errorf("Expected 2 children (cbu.id, nature-purpose), got %d", len(ast.Root.Children))
    }
}
```

### 1.2 EBNF Validator Tests (`internal/shared-dsl/validator/validator_test.go`)

**New Test File** - 15 test cases

```go
// Syntax Validation Tests
func TestValidateSyntax_ValidSExpression(t *testing.T)
func TestValidateSyntax_MissingClosingParen(t *testing.T)
func TestValidateSyntax_MissingOpeningParen(t *testing.T)
func TestValidateSyntax_InvalidVerbFormat(t *testing.T)
func TestValidateSyntax_ValidNestedExpressions(t *testing.T)

// Onboarding DSL Syntax Tests
func TestValidateSyntax_OnboardingVerbs(t *testing.T)
func TestValidateSyntax_OnboardingAttributes(t *testing.T)
func TestValidateSyntax_OnboardingResources(t *testing.T)

// Hedge Fund DSL Syntax Tests
func TestValidateSyntax_HedgeFundVerbs(t *testing.T)
func TestValidateSyntax_HedgeFundKYC(t *testing.T)

// Edge Cases
func TestValidateSyntax_EmptyDSL(t *testing.T)
func TestValidateSyntax_WhitespaceOnly(t *testing.T)
func TestValidateSyntax_CommentsIgnored(t *testing.T)
func TestValidateSyntax_UnicodeCharacters(t *testing.T)
func TestValidateSyntax_VeryLongDSL(t *testing.T)
```

### 1.3 Dictionary Service Tests (`internal/shared-dsl/dictionary/service_test.go`)

**Extend Existing Tests** - 10 new test cases

```go
// Cross-Domain Attribute Tests (NEW)
func TestDictionary_AttributeSharedAcrossDomains(t *testing.T)
func TestDictionary_OnboardingAttributes(t *testing.T)
func TestDictionary_HedgeFundAttributes(t *testing.T)
func TestDictionary_AttributeNameUniqueness(t *testing.T)

// Metadata Tests (NEW)
func TestDictionary_SourceMetadata(t *testing.T)
func TestDictionary_SinkMetadata(t *testing.T)
func TestDictionary_PrivacyFlags(t *testing.T)

// Query Performance Tests (NEW)
func TestDictionary_ConcurrentLookups(t *testing.T)
func TestDictionary_BulkAttributeRetrieval(t *testing.T)
func TestDictionary_CachingBehavior(t *testing.T)
```

**Sample Test**:
```go
func TestDictionary_AttributeSharedAcrossDomains(t *testing.T) {
    // Attribute "entity.legal_name" used by BOTH onboarding and hedge fund
    dict := NewMockDictionary()
    
    attrID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
    
    attr, err := dict.GetAttribute(context.Background(), attrID)
    if err != nil {
        t.Fatalf("Failed to get attribute: %v", err)
    }
    
    if attr.Name != "entity.legal_name" {
        t.Errorf("Expected 'entity.legal_name', got '%s'", attr.Name)
    }
    
    // Verify it's referenced in both domain vocabularies
    onboardingVocab := onboarding.GetVocabulary()
    hedgeFundVocab := hedgefund.GetVocabulary()
    
    // Both should reference the same attribute UUID
    // (Implementation specific to how vocabularies reference attributes)
}
```

### 1.4 Session Management Tests (`internal/shared-dsl/session/manager_test.go`)

**New Test File** - 15 test cases

```go
// Session Lifecycle Tests
func TestSession_CreateNew(t *testing.T)
func TestSession_GetExisting(t *testing.T)
func TestSession_GetOrCreate(t *testing.T)
func TestSession_Delete(t *testing.T)
func TestSession_Expiration(t *testing.T)

// DSL Accumulation Tests
func TestSession_AccumulateDSL_OnboardingOnly(t *testing.T)
func TestSession_AccumulateDSL_HedgeFundOnly(t *testing.T)
func TestSession_AccumulateDSL_CrossDomain(t *testing.T)
func TestSession_AccumulateDSL_EmptyAppend(t *testing.T)

// Context Tracking Tests
func TestSession_UpdateContext_InvestorID(t *testing.T)
func TestSession_UpdateContext_FundID(t *testing.T)
func TestSession_UpdateContext_MultipleEntities(t *testing.T)
func TestSession_UpdateContext_Override(t *testing.T)

// Concurrency Tests
func TestSession_ConcurrentAccess(t *testing.T)
func TestSession_ConcurrentDSLAccumulation(t *testing.T)
```

**Sample Test**:
```go
func TestSession_AccumulateDSL_CrossDomain(t *testing.T) {
    mgr := NewManager()
    
    // Create session starting in onboarding domain
    sess := mgr.GetOrCreate("test-session", "onboarding")
    
    // Accumulate onboarding DSL
    onboardingDSL := `(case.create (cbu.id "CBU-1234") (nature-purpose "Test"))`
    mgr.AccumulateDSL(sess.SessionID, onboardingDSL)
    
    // Switch to hedge fund domain
    sess.Domain = "hedge-fund-investor"
    
    // Accumulate hedge fund DSL
    hedgeFundDSL := `(investor.start-opportunity (legal-name "Acme Corp") (type "CORPORATE"))`
    mgr.AccumulateDSL(sess.SessionID, hedgeFundDSL)
    
    // Verify both DSLs are accumulated
    finalDSL := sess.BuiltDSL
    
    if !strings.Contains(finalDSL, "case.create") {
        t.Error("Expected onboarding DSL in accumulated result")
    }
    
    if !strings.Contains(finalDSL, "investor.start-opportunity") {
        t.Error("Expected hedge fund DSL in accumulated result")
    }
    
    // Verify order preservation
    caseIdx := strings.Index(finalDSL, "case.create")
    investorIdx := strings.Index(finalDSL, "investor.start-opportunity")
    
    if caseIdx > investorIdx {
        t.Error("Expected onboarding DSL before hedge fund DSL (chronological order)")
    }
}
```

### 1.5 UUID Resolver Tests (`internal/shared-dsl/resolver/resolver_test.go`)

**New Test File** - 10 test cases

```go
// Basic Resolution Tests
func TestResolve_SinglePlaceholder(t *testing.T)
func TestResolve_MultiplePlaceholders(t *testing.T)
func TestResolve_NoPlaceholders(t *testing.T)
func TestResolve_MissingContextValue(t *testing.T)

// Domain-Specific Tests
func TestResolve_OnboardingPlaceholders(t *testing.T)
func TestResolve_HedgeFundPlaceholders(t *testing.T)
func TestResolve_CrossDomainPlaceholders(t *testing.T)

// Edge Cases
func TestResolve_NestedPlaceholders(t *testing.T)
func TestResolve_InvalidPlaceholderSyntax(t *testing.T)
func TestResolve_PlaceholderInString(t *testing.T)
```

---

## Phase 2: Domain Registry Tests (Week 2)

### 2.1 Domain Interface Tests (`internal/domain-registry/domain_test.go`)

**New Test File** - 8 test cases

```go
// Interface Compliance Tests
func TestDomain_OnboardingImplementsInterface(t *testing.T)
func TestDomain_HedgeFundImplementsInterface(t *testing.T)

// Vocabulary Tests
func TestDomain_GetVocabulary_Onboarding(t *testing.T)
func TestDomain_GetVocabulary_HedgeFund(t *testing.T)

// Verb Validation Tests
func TestDomain_ValidateVerbs_OnboardingValid(t *testing.T)
func TestDomain_ValidateVerbs_OnboardingInvalid(t *testing.T)
func TestDomain_ValidateVerbs_HedgeFundValid(t *testing.T)
func TestDomain_ValidateVerbs_HedgeFundInvalid(t *testing.T)
```

### 2.2 Registry Tests (`internal/domain-registry/registry_test.go`)

**New Test File** - 10 test cases

```go
// Registration Tests
func TestRegistry_RegisterDomain(t *testing.T)
func TestRegistry_RegisterDuplicateDomain(t *testing.T)
func TestRegistry_RegisterNilDomain(t *testing.T)

// Lookup Tests
func TestRegistry_GetDomain_Exists(t *testing.T)
func TestRegistry_GetDomain_NotFound(t *testing.T)
func TestRegistry_ListDomains(t *testing.T)

// Vocabulary Tests
func TestRegistry_GetVocabulary_Onboarding(t *testing.T)
func TestRegistry_GetVocabulary_HedgeFund(t *testing.T)

// Concurrency Tests
func TestRegistry_ConcurrentRegistration(t *testing.T)
func TestRegistry_ConcurrentLookup(t *testing.T)
```

### 2.3 Router Tests (`internal/domain-registry/router_test.go`)

**New Test File** - 15 test cases

```go
// Basic Routing Tests
func TestRouter_RouteToOnboarding(t *testing.T)
func TestRouter_RouteToHedgeFund(t *testing.T)
func TestRouter_RouteToDefaultDomain(t *testing.T)

// Context-Based Routing Tests
func TestRouter_RouteByInvestorID(t *testing.T)
func TestRouter_RouteByCBUID(t *testing.T)
func TestRouter_RouteByFundID(t *testing.T)

// Message Analysis Routing Tests
func TestRouter_RouteByKeyword_Onboarding(t *testing.T)
func TestRouter_RouteByKeyword_HedgeFund(t *testing.T)
func TestRouter_RouteByVerb_Onboarding(t *testing.T)
func TestRouter_RouteByVerb_HedgeFund(t *testing.T)

// Fallback Tests
func TestRouter_FallbackToCurrentDomain(t *testing.T)
func TestRouter_FallbackToDefaultDomain(t *testing.T)

// Edge Cases
func TestRouter_AmbiguousMessage(t *testing.T)
func TestRouter_EmptyMessage(t *testing.T)
func TestRouter_NoDomainAvailable(t *testing.T)
```

---

## Phase 3: Onboarding Domain Tests (Week 3)

### 3.1 Vocabulary Migration Tests (`internal/domains/onboarding/vocab_test.go`)

**Migrate + Extend Existing Tests** - 20 test cases

```go
// MIGRATED from internal/dsl/vocab_test.go (8 existing tests)
func TestVocab_CaseManagement(t *testing.T)
func TestVocab_ProductService(t *testing.T)
func TestVocab_KYCCompliance(t *testing.T)
func TestVocab_ResourceInfrastructure(t *testing.T)
func TestVocab_AttributeData(t *testing.T)
func TestVocab_CompleteWorkflow(t *testing.T)
func TestVocab_Permutations(t *testing.T)
func TestVocab_DSLBlockCombination(t *testing.T)

// NEW - Registry Integration Tests (12 new tests)
func TestVocab_RegisterWithRegistry(t *testing.T)
func TestVocab_AllVerbsHaveDefinitions(t *testing.T)
func TestVocab_VerbArgumentsComplete(t *testing.T)
func TestVocab_StateTransitionsValid(t *testing.T)
func TestVocab_NoVerbOverlap_WithHedgeFund(t *testing.T)
func TestVocab_VerbNamingConventions(t *testing.T)
func TestVocab_68VerbsRegistered(t *testing.T)
func TestVocab_CategoryCoverage(t *testing.T)
func TestVocab_AttributeReferences_ExistInDictionary(t *testing.T)
func TestVocab_VerbDescriptions_NotEmpty(t *testing.T)
func TestVocab_VersionFormat(t *testing.T)
func TestVocab_DomainName_Correct(t *testing.T)
```

### 3.2 Agent Tests (`internal/domains/onboarding/agent_test.go`)

**Migrate + Extend Existing Tests** - 15 test cases

```go
// MIGRATED from internal/agent/dsl_agent_test.go (7 existing tests)
func TestAgent_ValidateVerbs_AllApproved(t *testing.T)
func TestAgent_ValidateVerbs_Unapproved(t *testing.T)
func TestAgent_ValidateVerbs_IgnoresNonVerbs(t *testing.T)
// ... (other existing verb validation tests)

// NEW - Domain Interface Tests (8 new tests)
func TestAgent_GenerateDSL_CaseCreate(t *testing.T)
func TestAgent_GenerateDSL_ProductsAdd(t *testing.T)
func TestAgent_GenerateDSL_KYCStart(t *testing.T)
func TestAgent_GenerateDSL_CompleteWorkflow(t *testing.T)
func TestAgent_UsesSharedDictionary(t *testing.T)
func TestAgent_UsesSharedParser(t *testing.T)
func TestAgent_ReturnsStandardizedResponse(t *testing.T)
func TestAgent_ErrorHandling(t *testing.T)
```

### 3.3 Validator Tests (`internal/domains/onboarding/validator_test.go`)

**Migrate Existing Tests** - 10 test cases

```go
// MIGRATED from internal/agent/dsl_agent_test.go
func TestValidator_AllOnboardingVerbs(t *testing.T)
func TestValidator_InvalidVerb(t *testing.T)
func TestValidator_MixedValid Invalid(t *testing.T)
func TestValidator_EmptyDSL(t *testing.T)
func TestValidator_IgnoresNonVerbConstructs(t *testing.T)

// NEW - Shared Parser Integration (5 new tests)
func TestValidator_UsesSharedParser(t *testing.T)
func TestValidator_ParseError_ReturnsError(t *testing.T)
func TestValidator_AllCategories_Covered(t *testing.T)
func TestValidator_VerbCount_68Verbs(t *testing.T)
func TestValidator_PerformanceBenchmark(t *testing.T)
```

### 3.4 Orchestrator Tests (`internal/domains/onboarding/orchestrator_test.go`)

**New Test File** - 12 test cases

```go
// Cross-Domain Orchestration Tests
func TestOrchestrator_CallHedgeFundDomain(t *testing.T)
func TestOrchestrator_CallKYCDomain(t *testing.T)
func TestOrchestrator_CallProductDomain(t *testing.T)

// DSL Composition Tests
func TestOrchestrator_ComposeDSL_OnboardingAndHedgeFund(t *testing.T)
func TestOrchestrator_ComposeDSL_MultipleSubdomains(t *testing.T)
func TestOrchestrator_ComposeDSL_PreserveOrder(t *testing.T)

// Context Propagation Tests
func TestOrchestrator_PropagateContext_ToHedgeFund(t *testing.T)
func TestOrchestrator_PropagateContext_BidirectionalSync(t *testing.T)

// Error Handling Tests
func TestOrchestrator_SubdomainNotAvailable(t *testing.T)
func TestOrchestrator_SubdomainVerbValidationFails(t *testing.T)
func TestOrchestrator_PartialFailure_Rollback(t *testing.T)
func TestOrchestrator_CircularDependencyDetection(t *testing.T)
```

---

## Phase 4: Integration Tests (Week 4)

### 4.1 End-to-End Onboarding Tests (`internal/integration/onboarding_e2e_test.go`)

**New Test File** - 20 test cases

```go
// Complete Onboarding Workflows
func TestE2E_Onboarding_MinimalCase(t *testing.T)
func TestE2E_Onboarding_FullWorkflow(t *testing.T)
func TestE2E_Onboarding_MultiProduct(t *testing.T)
func TestE2E_Onboarding_ComplexKYC(t *testing.T)
func TestE2E_Onboarding_WithResources(t *testing.T)

// State Transitions
func TestE2E_Onboarding_StateProgression(t *testing.T)
func TestE2E_Onboarding_InvalidStateTransition(t *testing.T)

// DSL Accumulation Verification
func TestE2E_Onboarding_DSLAccumulation_Correctness(t *testing.T)
func TestE2E_Onboarding_DSLAccumulation_Order(t *testing.T)
func TestE2E_Onboarding_DSLAccumulation_Idempotency(t *testing.T)

// Dictionary Integration
func TestE2E_Onboarding_AttributeResolution(t *testing.T)
func TestE2E_Onboarding_AttributeBinding(t *testing.T)
func TestE2E_Onboarding_PrivacyFlags_Respected(t *testing.T)

// Error Scenarios
func TestE2E_Onboarding_InvalidDSL_Rejected(t *testing.T)
func TestE2E_Onboarding_MissingAttribute_Error(t *testing.T)
func TestE2E_Onboarding_VerbValidation_Failure(t *testing.T)

// Comparison with Legacy
func TestE2E_Onboarding_OutputMatches_Legacy(t *testing.T)
func TestE2E_Onboarding_StateMatches_Legacy(t *testing.T)
func TestE2E_Onboarding_PerformanceCompare_Legacy(t *testing.T)
func TestE2E_Onboarding_DatabaseCompare_Legacy(t *testing.T)
```

**Critical Test Example**:
```go
func TestE2E_Onboarding_OutputMatches_Legacy(t *testing.T) {
    // This test ensures ZERO REGRESSION
    // Compare new multi-domain implementation with old single-domain
    
    // Setup: Same inputs for both implementations
    cbuID := "CBU-TEST-001"
    naturePurpose := "UCITS equity fund domiciled in LU"
    products := []string{"CUSTODY", "FUND_ACCOUNTING"}
    
    // OLD IMPLEMENTATION (existing code)
    legacyDSL := runLegacyOnboarding(cbuID, naturePurpose, products)
    
    // NEW IMPLEMENTATION (multi-domain)
    newDSL := runMultiDomainOnboarding(cbuID, naturePurpose, products)
    
    // Compare DSL outputs (normalize whitespace)
    if normalizeDSL(legacyDSL) != normalizeDSL(newDSL) {
        t.Errorf("DSL output mismatch:\nLegacy:\n%s\n\nNew:\n%s", legacyDSL, newDSL)
    }
    
    // Compare database state
    legacyState := getLegacyDBState(cbuID)
    newState := getMultiDomainDBState(cbuID)
    
    if !reflect.DeepEqual(legacyState, newState) {
        t.Errorf("Database state mismatch:\nLegacy: %+v\nNew: %+v", legacyState, newState)
    }
}
```

### 4.2 Cross-Domain Integration Tests (`internal/integration/cross_domain_test.go`)

**New Test File** - 15 test cases

```go
// Onboarding → Hedge Fund
func TestCrossDomain_OnboardingCallsHedgeFund(t *testing.T)
func TestCrossDomain_OnboardingToHedgeFund_Context(t *testing.T)
func TestCrossDomain_OnboardingToHedgeFund_DSLAccumulation(t *testing.T)

// Domain Switching
func TestCrossDomain_SwitchFromOnboardingToHedgeFund(t *testing.T)
func TestCrossDomain_SwitchFromHedgeFundToOnboarding(t *testing.T)
func TestCrossDomain_MultipleSwitches(t *testing.T)

// Shared Dictionary
func TestCrossDomain_SharedAttribute_LegalName(t *testing.T)
func TestCrossDomain_SharedAttribute_Domicile(t *testing.T)
func TestCrossDomain_SharedAttribute_DocumentType(t *testing.T)

// Verb Isolation
func TestCrossDomain_OnboardingVerbs_NotInHedgeFund(t *testing.T)
func TestCrossDomain_HedgeFundVerbs_NotInOnboarding(t *testing.T)

// State Machine Isolation
func TestCrossDomain_OnboardingState_Independent(t *testing.T)
func TestCrossDomain_HedgeFundState_Independent(t *testing.T)

// Error Propagation
func TestCrossDomain_SubdomainError_Handled(t *testing.T)
func TestCrossDomain_ContextMismatch_Error(t *testing.T)
```

### 4.3 Web Server Integration Tests (`internal/integration/web_server_test.go`)

**New Test File** - 15 test cases

```go
// Multi-Domain Endpoints
func TestWebServer_GetDomains(t *testing.T)
func TestWebServer_SwitchDomain_Onboarding(t *testing.T)
func TestWebServer_SwitchDomain_HedgeFund(t *testing.T)

// Chat Endpoint with Routing
func TestWebServer_Chat_OnboardingDomain(t *testing.T)
func TestWebServer_Chat_HedgeFundDomain(t *testing.T)
func TestWebServer_Chat_AutoRouting(t *testing.T)

// Session Management
func TestWebServer_Session_DSLAccumulation(t *testing.T)
func TestWebServer_Session_ContextTracking(t *testing.T)
func TestWebServer_Session_CrossDomain(t *testing.T)

// Vocabulary Endpoints
func TestWebServer_GetVocabulary_Onboarding(t *testing.T)
func TestWebServer_GetVocabulary_HedgeFund(t *testing.T)

// WebSocket Tests
func TestWebServer_WebSocket_MultiDomain(t *testing.T)
func TestWebServer_WebSocket_DomainSwitch(t *testing.T)

// Error Handling
func TestWebServer_InvalidDomain_404(t *testing.T)
func TestWebServer_UnapprovedVerb_400(t *testing.T)
```

---

## Phase 5: Performance Tests (Week 5)

### 5.1 Benchmark Tests (`internal/benchmarks/dsl_benchmarks_test.go`)

**New Test File** - 10 benchmark tests

```go
// Parser Benchmarks
func BenchmarkParser_OnboardingDSL_Small(b *testing.B)
func BenchmarkParser_OnboardingDSL_Large(b *testing.B)
func BenchmarkParser_HedgeFundDSL(b *testing.B)
func BenchmarkParser_MixedDSL(b *testing.B)

// Dictionary Benchmarks
func BenchmarkDictionary_LookupSingle(b *testing.B)
func BenchmarkDictionary_LookupBulk(b *testing.B)
func BenchmarkDictionary_ConcurrentLookups(b *testing.B)

// Domain Routing Benchmarks
func BenchmarkRouter_DomainSelection(b *testing.B)

// End-to-End Benchmarks
func BenchmarkE2E_OnboardingWorkflow(b *testing.B)
func BenchmarkE2E_CrossDomainWorkflow(b *testing.B)
```

**Performance Acceptance Criteria**:
- Parser: < 50ms for 100-line DSL
- Dictionary lookup: < 5ms per attribute
- Domain routing: < 10ms per request
- E2E onboarding workflow: < 500ms
- **No degradation from legacy implementation**

---

## Test Execution Strategy

### Continuous Integration

```yaml
# .github/workflows/test-ported-dsl.yml
name: Test Ported DSL

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Run Phase 1 - Shared Infrastructure
        run: go test ./internal/shared-dsl/... -v -coverprofile=coverage-shared.out
      
      - name: Run Phase 2 - Domain Registry
        run: go test ./internal/domain-registry/... -v -coverprofile=coverage-registry.out
      
      - name: Run Phase 3 - Onboarding Domain
        run: go test ./internal/domains/onboarding/... -v -coverprofile=coverage-onboarding.out
      
      - name: Coverage Report
        run: |
          go tool cover -func=coverage-shared.out
          go tool cover -func=coverage-registry.out
          go tool cover -func=coverage-onboarding.out

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          
