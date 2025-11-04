# Migration Answer: Backing Onboarding Domain into Hedge Fund Framework

## Your Question

> "I want to implement the approach we have for hedge fund investor and use it for the onboarding Domain DSL - onboarding is the overall orchestration domain - it will call hedge fund investor - it will call KYC it will call Product onboarding. What's the least messy way of backing the OB domain into the hedge framework? - I suspect it would be drop the DSL / Verbs for onboarding into our verbs lookup and then have this hedge investor implementation made 'multi-tenanted - second tenant == onboarding - does this seem a good approach - can you plan a migration plan we can then work through"

## Answer

**Yes, your intuition is correct, but with one critical refinement**: Instead of "multi-tenant" think **"multi-domain with shared infrastructure"**.

### Why Not "Multi-Tenant"?

Tenants typically share the same functionality with data isolation. What you have are different **domains** with:
- Different DSL vocabularies (verbs)
- Different state machines
- Different business logic
- **Shared** data dictionary (AttributeID-as-Type system)
- **Shared** infrastructure (parser, session management, etc.)

---

## The Approach: Multi-Domain Architecture

### What's Shared (Universal Infrastructure)

✅ **Data Dictionary** - All domains reference the same attribute UUIDs
- `entity.legal_name` with UUID-001 is the **same** across onboarding, hedge fund, and KYC
- This is your universal semantic type system

✅ **DSL Parser & EBNF Validator** - Domain-agnostic S-expression parsing
- Parses any domain's DSL (syntax validation only)
- Doesn't care about verb semantics

✅ **Session Management** - Stateful DSL accumulation
- Maintains `BuiltDSL` across all operations
- Tracks context (investor_id, fund_id, cbu_id, etc.)

✅ **UUID Resolver** - Replaces placeholders with actual UUIDs
- Works across all domains

### What's Domain-Specific

🔹 **Verb Vocabularies**
- **Onboarding**: 68 verbs (case.*, products.*, services.*, resources.*, etc.)
- **Hedge Fund**: 17 verbs (investor.*, subscription.*, redemption.*, etc.)
- **KYC**: TBD verbs (kyc.*, compliance.*, screening.*, etc.)
- **Product Onboarding**: TBD verbs

🔹 **Verb Validators**
- Each domain validates its own approved verbs
- Prevents AI from hallucinating operations

🔹 **Domain Agents**
- Each domain has an AI agent specialized in generating that domain's DSL

🔹 **State Machines**
- Each domain has its own lifecycle states

---

## The Migration Plan (6 Phases)

### Phase 1: Extract Shared Infrastructure (Week 1)

Create domain-agnostic packages:
```
internal/shared-dsl/
├── parser/          # S-expression parser (NEW)
├── validator/       # EBNF syntax validation (NEW)
├── dictionary/      # Universal attribute dictionary (existing, formalize interface)
├── session/         # Chat session management (extract from HF web server)
└── resolver/        # UUID resolution (extract from HF web server)
```

**Key Actions**:
- Extract DSL parser from existing code (check if in `internal/dsl/` or `hedge-fund-investor-source/`)
- Extract EBNF validator
- Formalize dictionary service interface
- Extract session management from `hedge-fund-investor-source/web/server.go`
- Extract UUID resolver from `hedge-fund-investor-source/web/internal/resolver/`

**Tests**: 65 new test cases

### Phase 2: Create Domain Registry (Week 2)

Create domain management system:
```
internal/domain-registry/
├── registry.go      # Register and lookup domains
├── domain.go        # Domain interface definition
└── router.go        # Route requests to appropriate domain
```

**Domain Interface**:
```go
type Domain interface {
    Name() string                    // "onboarding", "hedge-fund-investor"
    GetVocabulary() *Vocabulary      // Domain-specific verbs
    ValidateVerbs(dsl string) error  // Domain-specific validation
    GenerateDSL(ctx, req) (*resp, error)  // Domain-specific AI agent
}
```

**Routing Strategies**:
1. Context-based: If `investor_id` exists → hedge fund domain
2. Keyword-based: "onboard" → onboarding, "subscribe" → hedge fund
3. Verb-based: Parse DSL, check which domain owns the verb
4. Default: Use session's current domain

**Tests**: 33 new test cases

### Phase 3: Migrate Hedge Fund Domain (Week 3)

Refactor existing hedge fund implementation:
```
internal/domains/hedge-fund-investor/
├── domain.go        # Implement Domain interface
├── agent.go         # Migrate from hedge-fund-investor-source/web/internal/hf-agent/
├── vocab.go         # 17 verbs from hedge-fund-investor-source/hf-investor/dsl/
└── validator.go     # Verb validation (already exists, formalize)
```

**Key Changes**:
- Use shared parser instead of custom parsing
- Use shared dictionary service
- Use shared session management
- Return standardized responses

**Tests**: Migrate existing + 15 new

### Phase 4: Create Onboarding Domain (Week 4)

Port existing onboarding implementation to domain pattern:
```
internal/domains/onboarding/
├── domain.go        # Implement Domain interface
├── agent.go         # Migrate from internal/agent/dsl_agent.go
├── vocab.go         # 68 verbs from internal/dsl/vocab.go
├── validator.go     # Migrate from internal/agent/dsl_agent.go (validateDSLVerbs)
└── orchestrator.go  # NEW - Cross-domain orchestration logic
```

**Orchestrator Pattern**:
```go
// User says: "Onboard Acme Capital as a hedge fund investor"

// Step 1: Onboarding domain creates case
(case.create (cbu.id "CBU-1234") ...)

// Step 2: Orchestrator detects "hedge fund investor" keyword
// Step 3: Routes to hedge fund domain
(investor.start-opportunity (legal-name "Acme Capital") ...)

// Step 4: Both DSLs accumulated in session.BuiltDSL
```

**Tests**: Migrate 36 existing + 47 new

### Phase 5: Update Web Server (Week 5)

Refactor web server for multi-domain support:
```go
type Server struct {
    registry   *registry.Registry     // NEW: Domain registry
    sessionMgr *session.Manager       // NEW: Shared session manager
    dictionary dictionary.Service     // NEW: Shared dictionary
}

func NewServer(...) {
    reg := registry.NewRegistry()
    
    // Register onboarding domain
    obDomain := onboarding.NewOnboardingDomain(apiKey, dict, reg)
    reg.Register(obDomain)
    
    // Register hedge fund domain
    hfDomain := hedgefund.NewHedgeFundDomain(apiKey, dict)
    reg.Register(hfDomain)
}
```

**New Endpoints**:
- `/api/domains` - List available domains
- `/api/switch-domain` - Switch active domain
- `/api/chat` - Updated to use domain routing

**Frontend Updates**:
- Domain selector dropdown
- Display current domain
- Color-code DSL by domain

**Tests**: 15 integration tests

### Phase 6: Testing & Documentation (Week 6)

**Critical Tests**:
```go
// Zero regression test
func TestE2E_Onboarding_OutputMatches_Legacy(t *testing.T) {
    legacyDSL := runLegacyOnboarding(...)
    newDSL := runMultiDomainOnboarding(...)
    assert.Equal(legacyDSL, newDSL)  // Must be identical
}
```

**Documentation**:
- Architecture guide
- Domain developer guide (how to add new domains)
- API reference
- Migration guide

**Tests**: 50+ integration tests, 10 performance benchmarks

---

## Critical Design Decisions

### 1. Dictionary is Universal (NOT Domain-Specific)

**Correct**:
```sql
-- ONE dictionary table for ALL domains
CREATE TABLE "dsl-ob-poc".dictionary (
    attribute_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,  -- "entity.legal_name"
    ...
);

-- Both domains reference the SAME attribute
Onboarding DSL: (var (attr-id "uuid-001"))  → entity.legal_name
Hedge Fund DSL: (var (attr-id "uuid-001"))  → entity.legal_name
```

**Why**: AttributeID is the universal semantic contract across all domains.

### 2. Parser is Domain-Agnostic (Syntax Only)

**Parser** (shared): Validates S-expression syntax
```go
ast, err := parser.Parse(dsl)  // Just checks syntax
```

**Validator** (domain-specific): Validates approved verbs
```go
err := onboardingDomain.ValidateVerbs(dsl)  // Checks semantics
```

### 3. Verb Validation is Domain-Specific

**Onboarding Domain**:
```go
approvedVerbs := map[string]bool{
    "case.create": true,
    "products.add": true,
    // ... 68 onboarding verbs
}
```

**Hedge Fund Domain**:
```go
approvedVerbs := map[string]bool{
    "investor.start-opportunity": true,
    "subscription.submit": true,
    // ... 17 hedge fund verbs
}
```

**No Overlap**: Each domain owns its verbs exclusively.

### 4. DSL Accumulation is Cross-Domain

```go
type Session struct {
    BuiltDSL string  // Accumulates ALL domains' DSL
    Context  map[string]interface{}  // Tracks all entity IDs
    Domain   string  // Current active domain
}

// Example accumulated DSL
session.BuiltDSL = `
(case.create (cbu.id "CBU-1234") ...)

(investor.start-opportunity (legal-name "Acme") ...)

(kyc.begin (investor "uuid-123") ...)
`
```

---

## Success Criteria

### Functional
✅ All 68 onboarding verbs working  
✅ All 17 hedge fund verbs working  
✅ Cross-domain orchestration working  
✅ Shared dictionary queried by both domains  
✅ Domain switching mid-conversation works  

### Quality
✅ **Zero functional regression** - Onboarding DSL output matches legacy exactly  
✅ 80%+ code coverage  
✅ All linters passing  
✅ Complete documentation  

### Performance
✅ Domain routing < 10ms  
✅ Dictionary lookup < 5ms  
✅ DSL parsing < 50ms (100+ line DSL)  
✅ **No performance degradation** vs single-domain  

---

## Rollback Strategy

Each phase has a rollback checkpoint:
1. **Phase 1**: Shared infra extracted, hedge fund still works standalone
2. **Phase 2**: Registry exists, no functional changes
3. **Phase 3**: Hedge fund migrated, can disable via feature flag
4. **Phase 4**: Onboarding added, can disable via feature flag
5. **Phase 5**: Web server multi-domain, can fallback to single-domain

**Emergency Rollback**:
```bash
export ENABLE_MULTI_DOMAIN=false
# Or: git checkout v1.0-single-domain
```

---

## Testing Strategy for Ported DSL

**Your TODO: "We need to write new tests for the ported DSL"**

### Existing Tests (36 total)
- `internal/dsl/dsl_test.go` - 15 DSL generation tests
- `internal/dsl/vocab_test.go` - 8 vocabulary tests
- `internal/agent/dsl_agent_test.go` - 7 verb validation tests
- Various store/CLI tests - 6 tests

### New Tests Needed (195 total)
- **Shared Infrastructure**: 65 tests (parser, validator, dictionary, session, resolver)
- **Domain Registry**: 33 tests (registry, router, domain interface)
- **Onboarding Domain**: 47 tests (vocabulary, agent, validator, orchestrator)
- **Cross-Domain**: 50 tests (integration, e2e, web server)

### Critical Test
```go
func TestE2E_Onboarding_OutputMatches_Legacy(t *testing.T) {
    // Run SAME workflow in both implementations
    // DSL output MUST be byte-for-byte identical
    assert.Equal(legacyDSL, newDSL)
}
```

**Goal**: 231 total tests (36 existing + 195 new) = comprehensive coverage with zero regression

---

## Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| 1 | Week 1 | Shared infrastructure extracted + 65 tests |
| 2 | Week 2 | Domain registry created + 33 tests |
| 3 | Week 3 | Hedge fund migrated + tests |
| 4 | Week 4 | Onboarding domain created + 83 tests |
| 5 | Week 5 | Web server updated + 15 tests |
| 6 | Week 6 | Testing complete + documentation |

**Total**: 6 weeks to production-ready multi-domain architecture

---

## Summary

**Your Approach**: ✅ Correct intuition
- Drop onboarding verbs into shared lookup
- Make framework support multiple "tenants" (domains)

**Refinement**: Multi-domain (not multi-tenant)
- **Shared**: Dictionary, parser, session management, UUID resolution
- **Domain-specific**: Verbs, validators, agents, state machines
- **Orchestration**: Onboarding can call hedge fund, KYC, product domains

**Least Messy Migration**:
1. Extract shared infrastructure first (no functional changes)
2. Create domain registry and interface
3. Migrate hedge fund to domain pattern (validate no regression)
4. Port onboarding to domain pattern (validate zero regression)
5. Update web server for multi-domain
6. Comprehensive testing

**Key Insight**: The data dictionary (AttributeID-as-Type) is the universal contract that makes multi-domain composition possible. Different domains have different verbs, but they speak the same "type language" through shared attribute UUIDs.

**Next Steps**: See `MULTI_DOMAIN_MIGRATION_PLAN.md` and `TESTING_PLAN_PORTED_DSL.md` for detailed implementation guides.