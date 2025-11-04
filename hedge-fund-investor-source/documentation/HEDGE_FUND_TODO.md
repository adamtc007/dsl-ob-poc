# Hedge Fund Investor Register - Implementation Todo

## Current Status: Design Complete ✅

### ✅ Completed Foundation
1. **Create Postgres migration for Register of Investors schema** - Complete SQL DDL with event sourcing
2. **Implement JSON Schema for Investor Lifecycle DSL** - Full IR validation schema
3. **Design hedge fund investor module implementation plan** - 5-phase roadmap
4. **Create tiny example IR that validates against the schema** - Corporate subscription workflow
5. **Create SQL DDL with constraints, indexes, and views** - Production-ready database schema
6. **Implement Go IR type stubs and validators (stdlib only)** - Type-safe IR parsing/validation

### 🎯 Next Phase: Internal Library Implementation

## Library Architecture Plan

### Core Libraries Structure
```
internal/
├── hf-investor/           # Hedge Fund Investor Register (main lib)
│   ├── domain/           # Core business entities
│   ├── events/           # Event sourcing infrastructure
│   ├── state/            # Lifecycle state machine
│   ├── compliance/       # KYC/AML workflows
│   └── register/         # Register management
│
├── dsl-shared/           # Shared DSL Infrastructure
│   ├── vocab/            # Domain-tagged vocabulary system
│   ├── parser/           # S-expression parsing
│   ├── executor/         # Execution engine
│   └── ir/               # IR types and validation (existing)
│
└── event-sourcing/       # Shared Event Sourcing Library
    ├── store/            # Event store interface
    ├── projection/       # Projection engine
    └── replay/           # Event replay system
```

### Domain-Tagged Vocabulary System
```go
// DSL vocabulary tagged with hedge fund investor domain
type HedgeFundInvestorVocab struct {
    Domain   string                    `json:"domain"`   // "hedge-fund-investor"
    Version  string                    `json:"version"`  // "1.0.0"
    Verbs    map[string]VerbDefinition `json:"verbs"`
}

// Each verb tagged with domain context
type VerbDefinition struct {
    Name        string              `json:"name"`        // "investor.start-opportunity"
    Domain      string              `json:"domain"`      // "hedge-fund-investor"
    Category    string              `json:"category"`    // "investor-lifecycle"
    Args        map[string]ArgSpec  `json:"args"`
    StateChange *StateTransition    `json:"state_change,omitempty"`
}
```

## Bring-Up Kit (Pre-Implementation Validation)

### 🔧 IR Validator CLI
**File**: `cmd/ir-validate/main.go`
- Standalone CLI tool for validating IR plans
- Reads from stdin or file (`-in examples/ir/sub_issue.json`)
- Build: `go build -o bin/ir-validate ./cmd/ir-validate`
- Test: `./bin/ir-validate -in examples/ir/sub_issue.json`

### 🔑 Idempotency Helper
**File**: `internal/ir/idempotency.go`
- `DefaultIdempotencyKey(s Step)` - deterministic SHA-256 over {op,args}
- `ApplyIdempotency()` - fills key if nil, for retry safety
- Normalizes nested maps for consistent hashing

### 🧪 Validation Tests
**File**: `internal/ir/validate_test.go`
- `TestPlanOK()` - validates good IR plans
- `TestPlanBadUUID()` - tests error conditions
- `TestIdempotencyKeyGeneration()` - deterministic key testing
- Run: `go test ./internal/ir -v`

### 📊 Minimal Database Seed
**File**: `db/seed/seed_minimal.sql`
- Sample fund, share class, investor, register lot
- Quick smoke test data for register views
- Load: `psql "$DATABASE_URL" -f db/seed/seed_minimal.sql`
- Test: `SELECT * FROM register_of_investors_v;`

### 📋 Example IR Files
**File**: `examples/ir/sub_issue.json`
- 3-step workflow: opportunity → subscribe → issue
- Uses seed data UUIDs for consistency
- Validates against full IR schema

### 🗄️ Database Commands
```bash
# Up migration
goose -dir db/migrations postgres "$DATABASE_URL" up

# Seed minimal data
psql "$DATABASE_URL" -f db/seed/seed_minimal.sql

# Down migration (rollback)
goose -dir db/migrations postgres "$DATABASE_URL" down
```

### ⚠️ Common Gotchas
- **encoding/json**: Stick with stdlib, avoid json/v2
- **Decimals**: float64 for validation, decimal lib for booking
- **Dates**: YYYY-MM-DD format, keep UTC server-side
- **TRGM**: pg_trgm extension must exist before GIN index
- **Idempotency**: Apply `Step.ApplyIdempotency()` before persisting

## Critical Non-Functional Requirements

### 🚨 Rollback and Re-Run Capability (Priority 1)
**All hedge fund investor implementation changes MUST be able to be backed out and potentially re-run.**

**Technical Approaches:**
- **Database Migrations**: Complete up/down migration pairs with dependency-aware rollback
- **Feature Flags**: Runtime toggles for hedge fund investor functionality
- **Interface Isolation**: Clean abstraction boundaries allowing hot-swapping implementations
- **State Isolation**: Separate schema/tables that can be cleanly dropped without affecting core system
- **Event Sourcing**: Immutable event streams enable point-in-time rollback and replay
- **Library Versioning**: Semantic versioning with backward compatibility guarantees

**Rollback Testing:**
- Automated rollback tests for each migration
- Integration test suites that verify core system integrity after hedge fund module removal
- Performance baseline tests to ensure rollback doesn't degrade existing functionality

## Implementation Todo (Staged Approach)

### 🟡 Stage 1: Core Infrastructure (Week 1-2)
- [ ] **Extract DSL infrastructure to `internal/dsl-shared`**
  - [ ] Move existing IR types and validation
  - [ ] Create domain-tagged vocabulary system
  - [ ] Implement S-expression parser (extend existing)
  - [ ] Create execution engine interface

- [ ] **Create event sourcing library `internal/event-sourcing`**
  - [ ] Event store interface and PostgreSQL implementation
  - [ ] Projection engine for derived state
  - [ ] Event replay capabilities

- [ ] **Establish `internal/hf-investor` foundation**
  - [ ] Core domain entities (Investor, Trade, etc.)
  - [ ] Repository interfaces
  - [ ] Basic CRUD operations

### 🟡 Stage 2: Hedge Fund Domain Implementation (Week 3-4)
- [ ] **Implement hedge fund investor vocabulary**
  - [ ] Domain-tagged verbs (all 18 operations)
  - [ ] Hedge fund specific validation rules
  - [ ] Business logic integration

- [ ] **Build lifecycle state machine**
  - [ ] State transition engine
  - [ ] Guard conditions for state changes
  - [ ] State persistence and journaling

- [ ] **Event sourcing for register management**
  - [ ] Register event types (ISSUE, REDEEM, etc.)
  - [ ] Register lot projections
  - [ ] Event-driven register updates

### 🟡 Stage 3: Compliance and KYC (Week 5-6)
- [ ] **KYC workflow implementation**
  - [ ] Document management system
  - [ ] Screening integration interfaces
  - [ ] Approval workflow automation

- [ ] **Compliance features**
  - [ ] FATCA/CRS classification
  - [ ] Beneficial ownership tracking
  - [ ] Regulatory reporting views

- [ ] **Tax and banking integration**
  - [ ] Multi-currency banking instructions
  - [ ] Tax profile management
  - [ ] Withholding calculations

### 🟡 Stage 4: Integration and APIs (Week 7-8)
- [ ] **DSL execution integration**
  - [ ] Hedge fund IR executor
  - [ ] AttrRef resolution system
  - [ ] Idempotency handling

- [ ] **API layer**
  - [ ] REST APIs for all operations
  - [ ] Register reporting endpoints
  - [ ] Bulk operation support

- [ ] **CLI integration**
  - [ ] Hedge fund specific commands
  - [ ] IR file execution
  - [ ] Register management tools

## Library Design Principles

### 1. Domain Separation
- **Hedge Fund Investor**: Specific business logic, entities, workflows
- **DSL Shared**: Reusable across domains (onboarding, trading, etc.)
- **Event Sourcing**: Infrastructure library for any domain

### 2. Interface-Driven Design
```go
// Shared interfaces across domains
type EventStore interface {
    Append(ctx context.Context, streamID string, events []Event) error
    Load(ctx context.Context, streamID string) ([]Event, error)
}

type DSLExecutor interface {
    Execute(ctx context.Context, plan *ir.Plan) error
    ValidatePlan(plan *ir.Plan) error
}
```

### 3. Domain Tags for Vocabulary
```json
{
  "domain": "hedge-fund-investor",
  "version": "1.0.0",
  "verbs": {
    "investor.start-opportunity": {
      "domain": "hedge-fund-investor",
      "category": "investor-lifecycle",
      "state_transitions": ["OPPORTUNITY"]
    }
  }
}
```

### 4. Pluggable Architecture
- Event stores can be swapped (PostgreSQL, EventStore, etc.)
- DSL vocabularies are domain-specific but share infrastructure
- State machines are configurable per domain

## Integration Strategy

### Internal Library Dependencies
```
hf-investor
├── depends on: dsl-shared (vocab, parser, executor)
├── depends on: event-sourcing (store, projections)
└── depends on: existing datastore interfaces

dsl-shared
├── depends on: ir (types, validation)
└── standalone vocabulary system

event-sourcing
└── standalone infrastructure library
```

### External API
```go
// Public API for hedge fund investor register
package hfinvestor

type RegisterService struct {
    executor   dsl.Executor
    eventStore eventsourcing.EventStore
    repo       InvestorRepository
}

func (s *RegisterService) ExecuteWorkflow(ctx context.Context, irData []byte) error
func (s *RegisterService) GetRegister(ctx context.Context, fundID string) (*Register, error)
func (s *RegisterService) GetInvestor(ctx context.Context, investorID string) (*Investor, error)
```

## Testing Strategy

### Library-Level Testing
- **Unit tests**: Each library independently tested
- **Integration tests**: Cross-library interaction
- **End-to-end tests**: Full workflow validation

### Domain Testing
- **Hedge fund scenarios**: Complete investor lifecycles
- **DSL validation**: All 18 verbs with edge cases
- **Event sourcing**: Replay and projection testing

## Deployment Strategy

### Library Versioning
- **Semantic versioning** for each internal library
- **Domain vocabulary versioning** separate from infrastructure
- **Backward compatibility** for IR format changes

### Staged Rollout
1. **Internal libraries** - Complete and tested
2. **CLI integration** - Command-line tools
3. **API services** - REST endpoints
4. **UI integration** - Web interfaces

This approach creates reusable, domain-specific libraries while maintaining clean separation of concerns and enabling future expansion to other domains (trading, risk management, etc.).