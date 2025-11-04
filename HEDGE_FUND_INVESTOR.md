# Hedge Fund Investor Register & DSL

A comprehensive hedge fund investor onboarding and lifecycle management system with Domain-Specific Language (DSL) support, event sourcing, and operational reporting.

## 🎯 Overview

This system implements a complete hedge fund investor register with:

- **Event-Sourced Architecture**: Immutable audit trails for compliance
- **State Machine Management**: 11-state investor lifecycle with guard conditions
- **DSL Integration**: 17-verb Domain-Specific Language for workflow automation
- **Operational Reporting**: Position tracking, pipeline analytics, KYC monitoring
- **JSON Schema Validation**: Enterprise-grade type safety and validation
- **CLI Tools**: Production-ready command-line interface for operations

## 🏗️ Architecture

### Core Components

```
├── internal/hf-investor/
│   ├── domain/          # Core business entities and state machine
│   ├── dsl/             # DSL types, validators, and JSON schema
│   ├── state/           # State transition engine with guards
│   ├── store/           # Data access layer and interfaces
│   └── mocks/           # Test data generators
├── internal/cli/        # 20+ CLI commands for operations
├── sql/                 # Database migrations (Goose-compatible)
├── cmd/hf-cli/         # Standalone DSL validator CLI
└── examples/           # Sample runbooks and usage examples
```

### Event Sourcing Pattern

All state changes are captured as immutable events with complete audit trails:

```sql
-- Register Events: The source of truth for position tracking
CREATE TABLE hf_register_events (
  event_key        text NOT NULL UNIQUE,
  delta_units      numeric(24,8) NOT NULL,
  value_date       date NOT NULL,
  correlation_id   text,
  causation_id     text
);

-- Lots: Projected aggregates for fast queries
CREATE TABLE hf_register_lots (
  units            numeric(24,8) NOT NULL DEFAULT 0,
  last_activity_at timestamptz
);
```

## 🚀 Quick Start

### 1. Database Setup

```bash
# Apply hedge fund schema
psql "$DB_URL" -f sql/migration_hedge_fund_investor.sql

# Or using Goose
goose -dir sql postgres $DB_URL up
```

### 2. Create Investor

```bash
# Command line
go run . hf-create-investor \
  --code="INV-001" \
  --legal-name="Acme LP" \
  --type="CORPORATE" \
  --domicile="US"

# Generated DSL output:
# (investor.start-opportunity
#   :legal-name "Acme LP"
#   :type "CORPORATE"
#   :domicile "US")
```

### 3. Validate DSL Runbooks

```bash
# Via Makefile
make validate
make validate FILE=examples/runbook.sample.json

# Direct CLI
cat examples/runbook.sample.json | go run ./cmd/hf-cli dsl-validate -pretty
```

### 4. Query Operations

```bash
# Position as-of queries
go run . hf-positions --as-of="2024-12-31" --output="json"

# Pipeline funnel analytics
go run . hf-pipeline --output="table"

# Outstanding KYC requirements
go run . hf-outstanding-kyc --overdue --sort="due_date"
```

## 📊 DSL Vocabulary

### Investor Lifecycle (17 Verbs)

| Domain | Verbs | Purpose |
|--------|-------|---------|
| **investor** | `start-opportunity`, `record-indication` | Lead generation and interest tracking |
| **kyc** | `begin`, `collect-doc`, `screen`, `approve`, `refresh-schedule` | KYC/AML compliance workflow |
| **screen** | `continuous` | Ongoing monitoring setup |
| **tax** | `capture` | Tax form collection and validation |
| **bank** | `set-instruction` | Banking details management |
| **subscribe** | `request`, `issue` | Subscription processing |
| **cash** | `confirm` | Cash receipt confirmation |
| **deal** | `nav` | NAV dealing and pricing |
| **redeem** | `request`, `settle` | Redemption processing |
| **offboard** | `close` | Investor closure |

### Sample DSL Runbook

```json
{
  "runbook_id": "11111111-2222-3333-4444-555555555555",
  "steps": [
    {
      "verb": "investor.start-opportunity",
      "params": {
        "legal_name": "Acme LP",
        "investor_type": "CORPORATE",
        "domicile": "US"
      }
    },
    {
      "verb": "kyc.begin",
      "params": {
        "investor_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
        "jurisdiction": "US"
      }
    },
    {
      "verb": "subscribe.request",
      "params": {
        "investor_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
        "fund_id": "bbbbbbbb-cccc-dddd-eeee-ffffffffffff",
        "share_class_id": "cccccccc-dddd-eeee-ffff-000000000000",
        "amount": 5000000,
        "currency": "USD"
      }
    }
  ]
}
```

## 🔄 State Machine

### Investor States (11 States)

```
OPPORTUNITY → PRECHECKS → KYC_PENDING → KYC_APPROVED
    ↓             ↓           ↓            ↓
OFFBOARDED ← REDEEMED ← REDEEM_PENDING ← ACTIVE
                                          ↑
                     ISSUED ← FUNDED_PENDING_NAV
                        ↑            ↑
                 SUB_PENDING_CASH ←──┘
```

### State Transitions with Guards

```go
func (i *HedgeFundInvestor) CanTransitionTo(targetStatus string) bool {
    validTransitions := map[string][]string{
        InvestorStatusOpportunity:      {InvestorStatusPrechecks},
        InvestorStatusKYCPending:       {InvestorStatusKYCApproved, InvestorStatusPrechecks},
        InvestorStatusActive:           {InvestorStatusRedeemPending, InvestorStatusSubPendingCash},
        // ... complete state machine logic
    }
    // Validation logic with guard conditions
}
```

## 📈 Operational Reporting

### Position As-Of Queries

Fast projection queries using event aggregation:

```sql
-- Units per investor/fund/class/series as of a date
SELECT l.investor_id, l.fund_id, l.class_id, l.series_id,
       SUM(e.delta_units) AS units
FROM "hf-investor".hf_register_events e
JOIN "hf-investor".hf_register_lots l ON l.lot_id = e.lot_id
WHERE e.value_date <= $1::date
GROUP BY 1,2,3,4;
```

### Pipeline Funnel Analytics

```sql
-- Investor status counts for ops dashboard
SELECT status, COUNT(*) AS investors
FROM "hf-investor".hf_investors
GROUP BY status
ORDER BY status;
```

### Outstanding KYC Requirements

```sql
-- Document requirements with fulfillment tracking
SELECT investor_id, doc_type, status, requested_at, due_at,
       CASE WHEN due_at < CURRENT_DATE THEN
            CURRENT_DATE - due_at::date
       END AS days_overdue
FROM "hf-investor".hf_document_requirements
WHERE status IN ('REQUESTED','OVERDUE')
ORDER BY due_at NULLS LAST;
```

## 🛠️ CLI Commands (20+ Operations)

### Investor Management
```bash
hf-create-investor        # Create new investor
hf-record-indication      # Record investment interest
hf-begin-kyc             # Start KYC process
hf-approve-kyc           # Approve KYC completion
hf-subscribe-request     # Create subscription
```

### Compliance & Reporting
```bash
hf-screen-investor       # Run compliance screening
hf-set-continuous-screening  # Setup ongoing monitoring
hf-capture-tax-info      # Collect tax documentation
hf-outstanding-kyc       # Query pending requirements
```

### Position Management
```bash
hf-positions             # Position as-of queries
hf-pipeline              # Pipeline funnel analytics
hf-trading               # Trade execution workflow
```

### Output Formats
- **Table**: Human-readable console output
- **JSON**: API integration and processing
- **CSV**: Excel/analytics export

## 🔒 Data Schema

### Core Tables (15+ Tables)

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `hf_investors` | Investor master data | Status tracking, domicile, type |
| `hf_register_events` | Position events | Event sourcing, immutable audit |
| `hf_register_lots` | Position aggregates | Fast queries, trigger-maintained |
| `hf_document_requirements` | KYC tracking | Due dates, fulfillment status |
| `hf_kyc_profiles` | Compliance data | Risk ratings, refresh cycles |
| `hf_trades` | Trade execution | Lifecycle tracking, settlement |
| `hf_lifecycle_states` | State history | Complete audit trail |

### Event Sourcing Benefits

✅ **Complete Audit Trail**: Every position change tracked with correlation IDs
✅ **Point-in-Time Queries**: Reconstruct positions at any historical date
✅ **Compliance Ready**: Immutable records for regulatory requirements
✅ **Scalable**: Event aggregation for performance, detailed events for accuracy

## 🔬 Validation & Type Safety

### JSON Schema Validation

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/hf-investor/hf_dsl.schema.json",
  "title": "Hedge Fund Investor DSL",
  "properties": {
    "steps": {
      "type": "array",
      "items": { "$ref": "#/$defs/Step" }
    }
  },
  "$defs": {
    "Step": {
      "properties": {
        "verb": { "enum": ["investor.start-opportunity", ...] },
        "params": { "type": "object" }
      }
    }
  }
}
```

### Go Type System

```go
type Runbook struct {
    RunbookID string    `json:"runbook_id,omitempty"`
    AsOf      time.Time `json:"as_of,omitempty"`
    Steps     []Step    `json:"steps"`
}

func (r *Runbook) Validate() error {
    // Comprehensive validation with business rules
    // - Required fields, format validation
    // - Cross-field dependencies
    // - Business logic constraints
}
```

## 🧪 Testing & Quality

### Test Coverage
- **Domain Logic**: State machine validation, business rules
- **DSL Parsing**: All 17 verbs with parameter validation
- **Database**: SQL operations with comprehensive mocking
- **CLI**: Command execution and flag parsing

### Code Quality
```bash
make lint      # golangci-lint with 20+ linters
make test      # Comprehensive test suite
make check     # Pre-commit quality checks
```

## 🚀 Production Deployment

### Performance Optimizations
- **Composite Indexes**: `(cbu_id, created_at DESC)` for fast latest lookups
- **Trigger-Based Aggregation**: Real-time lot unit calculations
- **Event Partitioning**: Date-based partitioning for large volumes

### Monitoring & Observability
- **Correlation IDs**: Request tracing across operations
- **Audit Events**: Complete operational history
- **Health Checks**: Database connectivity and schema validation

### Scalability Patterns
- **Read Replicas**: Position queries against replicas
- **Event Streaming**: Kafka integration for real-time processing
- **API Gateway**: Rate limiting and authentication

## 📚 Examples & Usage

### Complete Investor Lifecycle

See `examples/runbook.sample.json` for a comprehensive runbook showcasing:

1. **Opportunity Creation** → Legal entity setup
2. **KYC Process** → Compliance workflow
3. **Tax Documentation** → W-8BEN-E collection
4. **Banking Setup** → Wire instruction capture
5. **Subscription** → Investment request processing
6. **Cash Confirmation** → Receipt validation
7. **NAV Dealing** → Pricing and allocation
8. **Unit Issuance** → Position creation
9. **Redemption** → Exit processing
10. **Offboarding** → Relationship closure

### Integration Patterns

```bash
# Workflow automation
cat investor_batch.json | go run ./cmd/hf-cli dsl-validate
psql "$DB_URL" -f generated_trades.sql

# API integration
curl -X POST /api/investors -d @runbook.json
curl -X GET /api/positions?as-of=2024-12-31

# Batch processing
make validate FILE=daily_subscriptions.json
./dsl-poc hf-positions --as-of="$(date +%Y-%m-%d)" --output=csv > positions.csv
```

## 🤝 Contributing

### Development Workflow
1. **Feature Branch**: Create from `main`
2. **Implementation**: Follow existing patterns
3. **Testing**: Add comprehensive tests
4. **Documentation**: Update relevant .md files
5. **Quality**: `make check` before commit

### Code Standards
- **Go Style**: Follow effective Go patterns
- **SQL**: PostgreSQL-specific optimizations
- **Documentation**: Comprehensive inline comments
- **Testing**: Business logic and edge cases

---

**Version**: v9 - Hedge Fund Investor Register
**Last Updated**: November 2025
**Production Ready**: ✅ Event sourcing, state machines, CLI tools, comprehensive testing