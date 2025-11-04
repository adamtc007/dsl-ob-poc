# Migration Quick Start Guide

**Multi-Domain Architecture Migration - Quick Reference**

## 🎯 Goal

Migrate the proven hedge fund investor chat framework to support multiple DSL domains (onboarding, hedge fund, KYC, product onboarding) with shared infrastructure.

---

## 🏗️ Core Architecture Principle

### Shared Across All Domains
- ✅ **Data Dictionary** - Universal attribute definitions (AttributeID-as-Type)
- ✅ **EBNF Grammar** - S-expression syntax rules
- ✅ **DSL Parser** - Parses any domain's DSL
- ✅ **Session Management** - Stateful conversation tracking
- ✅ **UUID Resolution** - Resolves placeholders to actual UUIDs

### Domain-Specific
- 🔹 **Verb Vocabularies** - Each domain defines its own verbs
- 🔹 **Domain Agents** - AI agents specialized in generating domain DSL
- 🔹 **Verb Validators** - Domain-specific approved verb lists
- 🔹 **State Machines** - Domain-specific lifecycle states

---

## 📋 6-Phase Migration Plan

### Phase 1: Extract Shared Infrastructure (Week 1)
**Create**:
```
internal/shared-dsl/
├── parser/          # S-expression parser (domain-agnostic)
├── validator/       # EBNF syntax validation
├── dictionary/      # Attribute dictionary service
├── session/         # Chat session management
└── resolver/        # UUID resolution
```

**Tests**: 65 new test cases

### Phase 2: Create Domain Registry (Week 2)
**Create**:
```
internal/domain-registry/
├── registry.go      # Register and lookup domains
├── domain.go        # Domain interface definition
└── router.go        # Route requests to domains
```

**Tests**: 33 new test cases

### Phase 3: Migrate Hedge Fund Domain (Week 3)
**Create**:
```
internal/domains/hedge-fund-investor/
├── domain.go        # Domain interface implementation
├── agent.go         # HF AI agent (migrated)
├── vocab.go         # 17 HF verbs
└── validator.go     # HF verb validator
```

**Tests**: Migrate existing + add 15 new

### Phase 4: Create Onboarding Domain (Week 4)
**Create**:
```
internal/domains/onboarding/
├── domain.go        # Domain interface implementation
├── agent.go         # Onboarding AI agent
├── vocab.go         # 68 onboarding verbs
├── validator.go     # Onboarding verb validator
└── orchestrator.go  # Cross-domain orchestration
```

**Tests**: Migrate 36 existing + add 47 new

### Phase 5: Update Web Server (Week 5)
**Modify**: `hedge-fund-investor-source/web/server.go`
- Add domain registry
- Update chat handler for routing
- Add domain switching endpoint
- Update frontend with domain selector

**Tests**: 15 integration tests

### Phase 6: Testing & Documentation (Week 6)
- 50+ integration tests
- Performance benchmarks
- Complete documentation
- Migration guide

---

## 🧪 Testing Strategy

### Test Coverage Goals
| Component | Existing | New | Total |
|-----------|----------|-----|-------|
| Shared Infrastructure | 0 | 65 | 65 |
| Domain Registry | 0 | 33 | 33 |
| Onboarding Domain | 36 | 47 | 83 |
| Cross-Domain | 0 | 50 | 50 |
| **TOTAL** | **36** | **195** | **231** |

### Critical Test Categories

#### 1. Regression Tests (Zero Functional Change)
```bash
# Test that onboarding DSL output matches legacy
go test ./internal/integration -run TestE2E_Onboarding_OutputMatches_Legacy
```

#### 2. Shared Infrastructure Tests
```bash
# Parser handles all domain DSLs
go test ./internal/shared-dsl/parser -v

# Dictionary shared across domains
go test ./internal/shared-dsl/dictionary -v

# Session DSL accumulation
go test ./internal/shared-dsl/session -v
```

#### 3. Domain Isolation Tests
```bash
# Onboarding verbs validated
go test ./internal/domains/onboarding -run TestValidator

# Hedge fund verbs validated
go test ./internal/domains/hedge-fund-investor -run TestValidator
```

#### 4. Cross-Domain Tests
```bash
# Onboarding orchestrates hedge fund
go test ./internal/integration -run TestCrossDomain
```

#### 5. Performance Tests
```bash
# No performance degradation
go test ./internal/benchmarks -bench=. -benchmem
```

---

## ✅ Pre-Migration Checklist

### Environment Setup
- [ ] Go 1.21+ installed (for greenteagc)
- [ ] PostgreSQL 15+ running
- [ ] `GEMINI_API_KEY` environment variable set
- [ ] All existing tests passing: `make test`

### Baseline Metrics
- [ ] Run existing tests and capture results
- [ ] Benchmark existing onboarding workflow
- [ ] Document current DSL output samples
- [ ] Export current database schema

---

## 🚀 Quick Migration Commands

### Phase 1: Extract Shared Infrastructure
```bash
# Create package structure
mkdir -p internal/shared-dsl/{parser,validator,dictionary,session,resolver}

# Extract parser from existing code
# TODO: Determine if parser exists in internal/dsl/ or hedge-fund-investor-source/

# Run tests
go test ./internal/shared-dsl/... -v -coverprofile=coverage-shared.out
go tool cover -func=coverage-shared.out
```

### Phase 2: Domain Registry
```bash
# Create registry structure
mkdir -p internal/domain-registry

# Implement domain interface
# Implement registry
# Implement router

# Run tests
go test ./internal/domain-registry/... -v
```

### Phase 3: Migrate Hedge Fund
```bash
# Create hedge fund domain
mkdir -p internal/domains/hedge-fund-investor

# Copy and refactor from hedge-fund-investor-source/
# Update to use shared infrastructure

# Run tests
go test ./internal/domains/hedge-fund-investor/... -v
```

### Phase 4: Create Onboarding Domain
```bash
# Create onboarding domain
mkdir -p internal/domains/onboarding

# Migrate vocabulary from internal/dsl/vocab.go
# Migrate agent from internal/agent/dsl_agent.go
# Create orchestrator

# Run tests
go test ./internal/domains/onboarding/... -v
```

### Phase 5: Update Web Server
```bash
# Update server.go with domain registry
# Add domain switching endpoint
# Update frontend with domain selector

# Test web server
go run hedge-fund-investor-source/web/server.go
# Visit http://localhost:8080
```

### Phase 6: Integration Testing
```bash
# Run all integration tests
go test ./internal/integration/... -v

# Run performance benchmarks
go test ./internal/benchmarks/... -bench=. -benchmem

# Compare with baseline
```

---

## 🔍 Testing Priorities

### P0 - Critical (Must Pass Before Merge)
1. ✅ All existing onboarding tests still pass
2. ✅ DSL output matches legacy implementation byte-for-byte
3. ✅ Shared parser handles all domain DSLs correctly
4. ✅ Dictionary lookups work across domains
5. ✅ Verb validation prevents unapproved verbs
6. ✅ Session DSL accumulation works correctly
7. ✅ Domain routing works for basic cases

### P1 - High Priority (Must Pass Before Production)
1. Cross-domain orchestration works
2. Domain switching mid-conversation works
3. Context tracking across domains works
4. All 231 tests passing
5. Performance within 10% of baseline
6. Complete documentation

### P2 - Nice to Have
1. Performance optimization (cache, etc.)
2. Advanced routing strategies
3. Domain-specific UI customizations
4. Metrics and observability

---

## 📊 Success Metrics

### Functional Metrics
- ✅ All 68 onboarding verbs working
- ✅ All 17 hedge fund verbs working
- ✅ Cross-domain orchestration working
- ✅ Shared dictionary queried by both domains
- ✅ Domain switching works mid-conversation

### Performance Metrics
- ✅ Domain routing < 10ms
- ✅ Dictionary lookup < 5ms
- ✅ DSL parsing < 50ms (for 100+ line DSL)
- ✅ Session manager supports 1000+ concurrent sessions
- ✅ No performance degradation vs single-domain

### Quality Metrics
- ✅ 80%+ code coverage across all packages
- ✅ Zero critical security vulnerabilities
- ✅ All linters passing (golangci-lint)
- ✅ Documentation complete for all public APIs

---

## 🔄 Rollback Strategy

### Feature Flags
```go
type Config struct {
    EnableMultiDomain bool
    EnabledDomains    []string  // ["onboarding", "hedge-fund-investor"]
}
```

### Rollback Checkpoints
1. **Phase 1 Complete**: Shared infra exists, hedge fund still standalone
2. **Phase 2 Complete**: Registry exists, no functional change yet
3. **Phase 3 Complete**: Hedge fund migrated, can disable via flag
4. **Phase 4 Complete**: Onboarding domain added, can disable via flag
5. **Phase 5 Complete**: Web server multi-domain, can fallback to single

### Emergency Rollback
```bash
# Disable multi-domain via environment variable
export ENABLE_MULTI_DOMAIN=false

# Or revert to previous git tag
git checkout v1.0-single-domain
make build
```

---

## 📚 Key Files Reference

### Existing Code (Don't Modify During Migration)
```
dsl-ob-poc/
├── internal/dsl/vocab.go                    # Onboarding vocabulary (SOURCE)
├── internal/agent/dsl_agent.go              # Onboarding agent (SOURCE)
├── internal/agent/dsl_agent_test.go         # Verb validation tests (MIGRATE)
├── internal/dsl/dsl_test.go                 # DSL tests (MIGRATE)
└── internal/dsl/vocab_test.go               # Vocabulary tests (MIGRATE)

hedge-fund-investor-source/
├── hf-investor/dsl/hedge_fund_dsl.go        # HF vocabulary (SOURCE)
├── web/internal/hf-agent/hf_dsl_agent.go    # HF agent (SOURCE)
└── web/server.go                            # Web server (WILL MODIFY)
```

### New Code (Create During Migration)
```
internal/
├── shared-dsl/                              # Phase 1
│   ├── parser/
│   ├── validator/
│   ├── dictionary/
│   ├── session/
│   └── resolver/
├── domain-registry/                         # Phase 2
│   ├── registry.go
│   ├── domain.go
│   └── router.go
└── domains/                                 # Phase 3-4
    ├── hedge-fund-investor/
    │   ├── domain.go
    │   ├── agent.go
    │   ├── vocab.go
    │   └── validator.go
    └── onboarding/
        ├── domain.go
        ├── agent.go
        ├── vocab.go
        ├── validator.go
        └── orchestrator.go
```

---

## 🐛 Common Issues & Solutions

### Issue: Tests fail after extracting shared infrastructure
**Solution**: Ensure all imports updated to use new package paths
```bash
# Find all files importing old paths
grep -r "internal/dsl" --include="*.go"

# Update imports
# Old: "dsl-ob-poc/internal/dsl"
# New: "dsl-ob-poc/internal/shared-dsl/parser"
```

### Issue: Parser doesn't handle domain-specific syntax
**Solution**: Parser should be syntax-only, domain validation is separate
```go
// Parser - domain-agnostic (handles S-expressions)
ast, err := parser.Parse(dsl)

// Validator - domain-specific (checks verbs)
err := domain.ValidateVerbs(dsl)
```

### Issue: Dictionary attributes not found
**Solution**: Ensure dictionary service initialized with correct connection
```go
dictService := dictionary.NewService(dbStore)
domain := onboarding.NewOnboardingDomain(apiKey, dictService, registry)
```

### Issue: Domain routing picks wrong domain
**Solution**: Check routing priority order (context > keyword > verb > default)
```go
// 1. Check context for entity IDs (investor_id → hedge fund)
// 2. Check message keywords ("onboard" → onboarding)
// 3. Parse DSL and check verb ownership
// 4. Use session's current domain as default
```

---

## 📖 Related Documentation

- **Full Migration Plan**: `MULTI_DOMAIN_MIGRATION_PLAN.md`
- **Testing Plan**: `TESTING_PLAN_PORTED_DSL.md`
- **Architecture**: `CLAUDE.md` (Core patterns: DSL-as-State, AttributeID-as-Type)
- **Hedge Fund Module**: `HEDGE_FUND_INVESTOR.md`
- **Dictionary Schema**: `SCHEMA_DOCUMENTATION.md`

---

## 🆘 Getting Help

### Review Checklist Before Asking
1. Read the full migration plan
2. Check test output for specific errors
3. Verify all imports are correct
4. Ensure database schema is up to date
5. Check environment variables are set

### Useful Debug Commands
```bash
# Check which tests are failing
go test ./... -v | grep FAIL

# Run specific test with verbose output
go test ./internal/domains/onboarding -v -run TestVocab_68VerbsRegistered

# Check coverage for specific package
go test ./internal/shared-dsl/parser -coverprofile=coverage.out
go tool cover -html=coverage.out

# Benchmark comparison
go test ./internal/benchmarks -bench=. -benchmem > new.txt
# Compare with baseline.txt
```

---

## ✨ Key Takeaways

1. **Shared Dictionary is Universal** - All domains reference the same attribute UUIDs
2. **Parser is Domain-Agnostic** - Validates syntax only, not semantics
3. **Verb Validation is Domain-Specific** - Each domain has its own approved verb list
4. **Zero Functional Regression** - Onboarding DSL output must match legacy exactly
5. **Test First** - Write tests before migrating code to catch regressions early

**Goal**: Multi-domain architecture with shared infrastructure, zero functional regression, complete test coverage.