# Multi-Domain Migration - Continuation Prompt for New Thread

## Current Status: Phase 1 COMPLETE ✅

**Date Completed**: 2024-01-XX  
**Phase Progress**: Phase 1 of 6 (Extract Shared Infrastructure) - **100% COMPLETE**

---

## ✅ What Has Been Completed

### Phase 1: Shared Infrastructure (COMPLETE)

All domain-agnostic infrastructure has been extracted and fully tested:

#### 1. ✅ DSL Parser (COMPLETE)
- **Location**: `internal/shared-dsl/parser/`
- **Files**: `parser.go` (520 lines), `parser_test.go` (1,148 lines)
- **Tests**: 31 tests, **88.5% coverage**
- **Performance**: 11.3 μs for 100-line DSL (4,400x faster than 50ms target)
- **Verified**: Parses onboarding DSL, hedge fund DSL, and cross-domain DSL

#### 2. ✅ Session Manager (COMPLETE)
- **Location**: `internal/shared-dsl/session/`
- **Files**: `manager.go` (381 lines), `manager_test.go` (816 lines)
- **Tests**: 30 tests, **96.9% coverage**
- **Features**: DSL accumulation, context tracking, domain switching, thread-safe
- **Verified**: Concurrent access tested (100 goroutines)

#### 3. ✅ UUID Resolver (COMPLETE)
- **Location**: `internal/shared-dsl/session/resolver/`
- **Files**: `resolver.go` (284 lines), `resolver_test.go` (841 lines)
- **Tests**: 40 tests, **95.3% coverage**
- **Features**: Placeholder resolution, alternate forms (camelCase, snake_case), context extraction
- **Verified**: Resolves placeholders in onboarding and hedge fund DSL

#### 4. ✅ Deprecated Files Marked
- `internal/dsl/vocab.go` - Marked DEPRECATED
- `internal/dsl/dsl.go` - Marked DEPRECATED
- `internal/agent/dsl_agent.go` - Marked DEPRECATED

#### 5. ✅ Documentation Created
- `MULTI_DOMAIN_MIGRATION_PLAN.md` - Complete 6-phase plan
- `TESTING_PLAN_PORTED_DSL.md` - Comprehensive testing strategy
- `MIGRATION_QUICK_START.md` - Quick reference guide
- `MIGRATION_DEPRECATION_TRACKER.md` - Tracks deprecated code
- `PARSER_COMPLETE_SUMMARY.md` - Parser completion details
- `PHASE_1_STATUS.md` - Phase 1 progress tracker

---

## 🎯 Next Steps: Phase 2 - Domain Registry

### Phase 2 Overview (Week 2)

Create the domain registry system to enable multiple domains to coexist with dynamic routing.

### Tasks for Phase 2:

#### 2.1 Define Domain Interface (`internal/domain-registry/domain.go`)
Create the interface that all domains must implement:

```go
type Domain interface {
    Name() string                    // "onboarding", "hedge-fund-investor"
    Version() string                 // "1.0.0"
    GetVocabulary() *Vocabulary      // Domain-specific verbs
    ValidateVerbs(dsl string) error  // Domain-specific validation
    GenerateDSL(ctx, req) (*resp, error)  // Domain-specific AI agent
    GetCurrentState(context) string  // State machine
    ValidateTransition(from, to string) error
}
```

#### 2.2 Create Domain Registry (`internal/domain-registry/registry.go`)
- Register multiple domains
- Lookup domains by name
- Thread-safe operations
- List all registered domains

#### 2.3 Create Domain Router (`internal/domain-registry/router.go`)
Routing strategies:
1. Context-based: If `investor_id` exists → hedge fund domain
2. Keyword-based: "onboard" → onboarding, "subscribe" → hedge fund
3. Verb-based: Parse DSL, check which domain owns the verb
4. Default: Use session's current domain

#### 2.4 Write Tests (33 tests needed)
- Domain interface compliance tests
- Registry tests (register, lookup, thread-safety)
- Router tests (all routing strategies)

---

## 📊 Overall Migration Progress

| Phase | Status | Progress |
|-------|--------|----------|
| 1. Extract Shared Infrastructure | ✅ COMPLETE | 100% |
| 2. Create Domain Registry | ⏳ NEXT | 0% |
| 3. Migrate Hedge Fund Domain | ⏸️ TODO | 0% |
| 4. Create Onboarding Domain | ⏸️ TODO | 0% |
| 5. Update Web Server | ⏸️ TODO | 0% |
| 6. Testing & Documentation | ⏸️ TODO | 0% |

**Overall**: 1 of 6 phases complete (17%)

---

## 🔑 Key Architectural Insights

### Shared Infrastructure (Universal)
1. **Data Dictionary** - Universal attribute definitions (AttributeID-as-Type)
2. **DSL Parser** - Parses S-expressions (syntax only)
3. **Session Manager** - Stateful DSL accumulation
4. **UUID Resolver** - Resolves placeholders

### Domain-Specific
1. **Verb Vocabularies** - Each domain has its own verbs
   - Onboarding: 68 verbs (case.*, products.*, services.*, resources.*)
   - Hedge Fund: 17 verbs (investor.*, subscription.*, redemption.*)
2. **Verb Validators** - Domain-specific approved verb lists
3. **Domain Agents** - AI agents that generate domain DSL
4. **State Machines** - Domain-specific lifecycle states

### Critical Pattern: DSL-as-State
- The accumulated DSL document IS the state
- Each operation appends to the DSL
- State is reconstructed by parsing the DSL
- Provides complete audit trail for compliance

---

## 🚀 How to Continue in New Thread

### Prompt for Claude:

```
Continue the multi-domain migration for dsl-ob-poc project.

Current Status:
- Phase 1 (Extract Shared Infrastructure) is COMPLETE
- Parser, Session Manager, and UUID Resolver all implemented with 88-96% test coverage
- All shared infrastructure is in internal/shared-dsl/

Next Task: Phase 2 - Create Domain Registry System

Please:
1. Review PHASE_1_STATUS.md and MULTI_DOMAIN_MIGRATION_PLAN.md
2. Start Phase 2 by creating internal/domain-registry/ with:
   - domain.go (Domain interface)
   - registry.go (Registry implementation)
   - router.go (Domain routing logic)
3. Write comprehensive tests for each component
4. Follow the same quality standards as Phase 1 (80%+ coverage)

Key files to reference:
- MULTI_DOMAIN_MIGRATION_PLAN.md (complete plan)
- PHASE_1_STATUS.md (what's done)
- internal/shared-dsl/ (shared infrastructure to build upon)
```

---

## 📁 Key Files Reference

### Shared Infrastructure (Use These)
```
internal/shared-dsl/
├── parser/
│   ├── parser.go                    # S-expression parser
│   └── parser_test.go               # 31 tests, 88.5% coverage
├── session/
│   ├── manager.go                   # Session management
│   └── manager_test.go              # 30 tests, 96.9% coverage
└── resolver/
    ├── resolver.go                  # UUID resolver
    └── resolver_test.go             # 40 tests, 95.3% coverage
```

### Deprecated (Reference Only, Do Not Modify)
```
internal/dsl/vocab.go                # Onboarding vocabulary (TO MIGRATE)
internal/dsl/dsl.go                  # DSL builders (TO MIGRATE)
internal/agent/dsl_agent.go          # Onboarding agent (TO MIGRATE)
```

### Documentation
```
MULTI_DOMAIN_MIGRATION_PLAN.md       # Complete 6-phase plan
MIGRATION_QUICK_START.md             # Quick reference
TESTING_PLAN_PORTED_DSL.md           # Testing strategy
MIGRATION_DEPRECATION_TRACKER.md     # Deprecated files tracker
```

---

## ⚠️ Important Notes

### Do NOT Delete Yet
All deprecated files are marked but **kept for reference** until:
1. All 6 phases complete
2. All 231 tests passing
3. Zero regression verified
4. Performance benchmarks met

### Testing Standards
- Minimum 80% code coverage (we achieved 88-96%)
- Test both onboarding AND hedge fund DSL examples
- Include concurrency tests where applicable
- Verify zero regression with existing tests

### Architectural Principles
1. **Domain-agnostic shared infrastructure** - parser doesn't know about verbs
2. **Domain-specific validation** - each domain validates its own verbs
3. **AttributeID-as-Type** - dictionary is universal across domains
4. **DSL-as-State** - accumulated DSL is the complete state

---

## 🎯 Success Criteria for Phase 2

### Functional
- ✅ Domain interface defined with full documentation
- ✅ Registry manages multiple domains (register, lookup, list)
- ✅ Router correctly routes requests to appropriate domain
- ✅ Thread-safe concurrent access
- ✅ All routing strategies work (context, keyword, verb, default)

### Quality
- ✅ 33+ tests passing
- ✅ 80%+ code coverage
- ✅ Comprehensive API documentation
- ✅ Integration with Phase 1 shared infrastructure verified

### Performance
- ✅ Domain routing < 10ms
- ✅ Registry lookup < 1ms
- ✅ No performance degradation

---

## 📝 Commands to Run

### Verify Phase 1 Still Works
```bash
# Run all Phase 1 tests
go test ./internal/shared-dsl/... -v -coverprofile=coverage.out

# Should see:
# parser:   31 tests PASS, 88.5% coverage
# session:  30 tests PASS, 96.9% coverage  
# resolver: 40 tests PASS, 95.3% coverage
```

### Start Phase 2
```bash
# Create directory structure
mkdir -p internal/domain-registry

# Create files (see Phase 2 tasks above)
# - domain.go
# - registry.go
# - router.go
# - domain_test.go
# - registry_test.go
# - router_test.go
```

---

## 🔗 Related Documentation

- **CLAUDE.md** - Core architectural patterns (DSL-as-State, AttributeID-as-Type)
- **HEDGE_FUND_TODO.md** - Hedge fund module plans
- **MULTI_DOMAIN_MIGRATION_PLAN.md** - Complete 6-phase migration plan (959 lines)

---

**Phase 1 Complete!** Ready for Phase 2: Domain Registry System 🚀

**Total Progress**: 17% of overall migration (1 of 6 phases)
**Next Milestone**: Phase 2 complete (33%)