# Migration Deprecation Tracker

**Status**: Migration In Progress  
**Started**: 2024-01-XX  
**Target Completion**: 2024-XX-XX (6 weeks)

---

## Purpose

This document tracks all code marked for deletion during the multi-domain migration. Files are kept for reference and comparison testing until migration is complete and all tests pass.

**DO NOT DELETE** any files marked below until:
1. ✅ All migration phases complete
2. ✅ All 231 tests passing
3. ✅ Zero regression verified
4. ✅ Performance benchmarks meet criteria
5. ✅ Documentation complete

---

## Deprecation Status Legend

- 🔴 **DEPRECATED** - Will be deleted, do not modify
- 🟡 **MIGRATING** - Being actively migrated
- 🟢 **MIGRATED** - New implementation complete, old code can be deleted after testing
- ✅ **DELETED** - Removed from codebase

---

## Phase 1: Onboarding Code to Migrate

### Core DSL Vocabulary

| File | Status | New Location | Notes |
|------|--------|--------------|-------|
| `internal/dsl/vocab.go` | 🔴 DEPRECATED | `internal/domains/onboarding/vocab.go` | 68 onboarding verbs |
| `internal/dsl/vocab_test.go` | 🔴 DEPRECATED | `internal/domains/onboarding/vocab_test.go` | 8 existing tests + 12 new |
| `internal/dsl/dsl.go` | 🔴 DEPRECATED | `internal/domains/onboarding/builder.go` | DSL builder functions |
| `internal/dsl/dsl_test.go` | 🔴 DEPRECATED | `internal/domains/onboarding/builder_test.go` | 15 DSL generation tests |
| `internal/dsl/executor.go` | 🔴 DEPRECATED | `internal/shared-dsl/executor/executor.go` | Move to shared (domain-agnostic) |
| `internal/dsl/executor_test.go` | 🔴 DEPRECATED | `internal/shared-dsl/executor/executor_test.go` | Executor tests |

### Agent & Verb Validation

| File | Status | New Location | Notes |
|------|--------|--------------|-------|
| `internal/agent/agent.go` | 🔴 DEPRECATED | `internal/domains/onboarding/agent.go` | Onboarding AI agent |
| `internal/agent/dsl_agent.go` | 🔴 DEPRECATED | `internal/domains/onboarding/agent.go` | DSL transformation agent |
| `internal/agent/dsl_agent_test.go` | 🔴 DEPRECATED | `internal/domains/onboarding/validator_test.go` | 7 verb validation tests |
| `internal/agent/kyc_agent.go` | 🟡 KEEP | - | KYC discovery agent (may become separate domain) |

### CLI Commands

| File | Status | New Location | Notes |
|------|--------|--------------|-------|
| `internal/cli/create.go` | 🟡 REFACTOR | Update to use domain registry | CLI entry point |
| `internal/cli/add_products.go` | 🟡 REFACTOR | Update to use domain registry | CLI entry point |
| `internal/cli/discover_kyc.go` | 🟡 REFACTOR | Update to use domain registry | CLI entry point |
| `internal/cli/discover_services.go` | 🟡 REFACTOR | Update to use domain registry | CLI entry point |
| `internal/cli/discover_resources.go` | 🟡 REFACTOR | Update to use domain registry | CLI entry point |
| `internal/cli/history.go` | 🟡 KEEP | - | Domain-agnostic, no changes needed |
| `internal/cli/history_test.go` | 🟡 KEEP | - | Keep existing tests |

### Store & Database (Keep - Shared Infrastructure)

| File | Status | Notes |
|------|--------|-------|
| `internal/store/*.go` | 🟢 KEEP | Database operations - shared across all domains |
| `internal/dictionary/*.go` | 🟢 KEEP | Dictionary service - shared across all domains |
| `sql/*.sql` | 🟢 KEEP | Database schema - universal |

---

## Phase 2: Code to Extract into Shared Infrastructure

### Session Management (from Hedge Fund)

| File | Status | New Location | Notes |
|------|--------|--------------|-------|
| `hedge-fund-investor-source/web/server.go` | 🟡 REFACTOR | Update for multi-domain | Keep, modify for registry |
| `hedge-fund-investor-source/web/internal/dslstate/manager.go` | 🔴 EXTRACT | `internal/shared-dsl/session/manager.go` | DSL state management |
| `hedge-fund-investor-source/web/internal/dslstate/manager_test.go` | 🔴 EXTRACT | `internal/shared-dsl/session/manager_test.go` | Session tests |
| `hedge-fund-investor-source/web/internal/context/context.go` | 🔴 EXTRACT | `internal/shared-dsl/session/context.go` | Context tracking |
| `hedge-fund-investor-source/web/internal/resolver/resolver.go` | 🔴 EXTRACT | `internal/shared-dsl/resolver/resolver.go` | UUID resolution |

### Parser (Determine Location)

| File | Status | New Location | Notes |
|------|--------|--------------|-------|
| `internal/dsl/parser.go` (if exists) | 🔴 EXTRACT | `internal/shared-dsl/parser/parser.go` | S-expression parser |
| OR: Create new parser | 🆕 NEW | `internal/shared-dsl/parser/parser.go` | If no existing parser |

---

## Phase 3: Hedge Fund Code to Migrate

### Hedge Fund Domain Implementation

| File | Status | New Location | Notes |
|------|--------|--------------|-------|
| `hedge-fund-investor-source/hf-investor/dsl/hedge_fund_dsl.go` | 🔴 DEPRECATED | `internal/domains/hedge-fund-investor/vocab.go` | 17 HF verbs |
| `hedge-fund-investor-source/hf-investor/dsl/hedge_fund_dsl_test.go` | 🔴 DEPRECATED | `internal/domains/hedge-fund-investor/vocab_test.go` | HF vocab tests |
| `hedge-fund-investor-source/web/internal/hf-agent/hf_dsl_agent.go` | 🔴 DEPRECATED | `internal/domains/hedge-fund-investor/agent.go` | HF AI agent |
| `hedge-fund-investor-source/hf-investor/state/state_machine.go` | 🟢 KEEP/REFACTOR | `internal/domains/hedge-fund-investor/state.go` | HF state machine |
| `hedge-fund-investor-source/hf-investor/domain/*.go` | 🟢 KEEP | - | Domain models (keep as-is) |

---

## Migration Checklist by Phase

### Phase 1: Shared Infrastructure ⏳ IN PROGRESS

- [ ] Create `internal/shared-dsl/` package structure
- [ ] Extract parser (or create new)
- [ ] Extract EBNF validator
- [ ] Formalize dictionary service interface
- [ ] Extract session management from HF web server
- [ ] Extract UUID resolver
- [ ] Write 65 new tests
- [ ] Mark deprecated files with `// DEPRECATED:` comments

### Phase 2: Domain Registry

- [ ] Create `internal/domain-registry/` package
- [ ] Define Domain interface
- [ ] Implement Registry
- [ ] Implement Router
- [ ] Write 33 new tests

### Phase 3: Migrate Hedge Fund Domain

- [ ] Create `internal/domains/hedge-fund-investor/`
- [ ] Implement Domain interface
- [ ] Migrate agent to use shared infrastructure
- [ ] Migrate vocabulary
- [ ] Create validator
- [ ] Update tests
- [ ] Verify zero regression

### Phase 4: Create Onboarding Domain

- [ ] Create `internal/domains/onboarding/`
- [ ] Migrate vocabulary from `internal/dsl/vocab.go`
- [ ] Migrate agent from `internal/agent/dsl_agent.go`
- [ ] Migrate validator from `internal/agent/dsl_agent.go`
- [ ] Create orchestrator
- [ ] Migrate 36 existing tests
- [ ] Write 47 new tests
- [ ] **CRITICAL**: Run `TestE2E_Onboarding_OutputMatches_Legacy`

### Phase 5: Update Web Server

- [ ] Update `hedge-fund-investor-source/web/server.go` for multi-domain
- [ ] Add domain registry to server
- [ ] Update chat handler with routing
- [ ] Add `/api/domains` endpoint
- [ ] Add `/api/switch-domain` endpoint
- [ ] Update frontend with domain selector
- [ ] Write 15 integration tests

### Phase 6: Testing & Cleanup

- [ ] Run all 231 tests
- [ ] Verify zero regression
- [ ] Run performance benchmarks
- [ ] Complete documentation
- [ ] Code review
- [ ] **FINAL STEP**: Delete deprecated files (change status to ✅)

---

## Deprecation Markers in Code

All deprecated files will be marked with this header comment:

```go
// DEPRECATED: This file is marked for deletion as part of multi-domain migration.
// 
// Migration Status: Phase X
// New Location: internal/domains/onboarding/filename.go
// Deprecation Date: 2024-XX-XX
// Planned Deletion: After Phase 6 complete and all tests passing
// 
// DO NOT MODIFY THIS FILE - It is kept for reference only.
// See MIGRATION_DEPRECATION_TRACKER.md for details.

package oldpackage
```

---

## Regression Testing Requirements

Before deleting ANY deprecated file, verify:

1. ✅ **Output Comparison**: DSL output matches legacy byte-for-byte
2. ✅ **Database State**: Database state matches legacy
3. ✅ **API Responses**: All API responses identical
4. ✅ **Performance**: No degradation (within 10%)
5. ✅ **Test Coverage**: 80%+ coverage on new code

---

## Rollback Plan

If migration fails at any phase:

1. Revert all changes in current phase
2. Re-enable deprecated code (remove DEPRECATED markers temporarily)
3. Disable multi-domain via feature flag: `ENABLE_MULTI_DOMAIN=false`
4. Investigate issues
5. Resume migration after fixes

---

## Timeline

| Phase | Week | Status | Completion Date |
|-------|------|--------|-----------------|
| 1. Shared Infrastructure | Week 1 | ⏳ IN PROGRESS | TBD |
| 2. Domain Registry | Week 2 | ⏸️ PENDING | TBD |
| 3. Migrate Hedge Fund | Week 3 | ⏸️ PENDING | TBD |
| 4. Create Onboarding | Week 4 | ⏸️ PENDING | TBD |
| 5. Update Web Server | Week 5 | ⏸️ PENDING | TBD |
| 6. Testing & Cleanup | Week 6 | ⏸️ PENDING | TBD |

---

## Notes

- All deprecated files remain in git history even after deletion
- Use `git log --follow` to track file renames/moves
- Keep CHANGELOG.md updated with each phase completion
- Tag releases after each successful phase: `v2.0-phase1`, `v2.0-phase2`, etc.

---

## Questions & Issues

Track any issues or questions during migration:

| Date | Issue | Resolution | Status |
|------|-------|------------|--------|
| TBD | Example issue | Example resolution | ✅ Resolved |

---

**Last Updated**: 2024-XX-XX  
**Updated By**: Migration Team  
**Next Review**: After Phase 1 completion