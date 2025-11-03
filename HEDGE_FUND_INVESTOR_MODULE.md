# Hedge Fund Investor Module - Design Specification

## Overview

A production-ready hedge fund Register of Investors system built on event sourcing principles with a Domain-Specific Language (DSL) for investor lifecycle management. This module handles the complete investor journey from opportunity identification through offboarding, with full regulatory compliance and audit capabilities.

## Core Architecture Principles

### Event Sourcing Foundation
- **Immutable Events**: All investor actions recorded as immutable `register_event` entries
- **Derived State**: Current positions maintained in `register_lot` table derived from events
- **Complete Auditability**: Full reconstruction of any historical state from event log
- **Regulatory Compliance**: Built-in audit trails for regulatory examination

### Lifecycle State Machine
- **Clear Progression**: Defined states from `OPPORTUNITY` to `OFFBOARDED`
- **Guard Conditions**: Explicit requirements for state transitions
- **Business Logic Enforcement**: KYC approval before funding, cash confirmation before allocation
- **Idempotent Operations**: Safe retry of any operation without side effects

### DSL-Driven Operations
- **Business Semantics**: Verbs map directly to fund administration operations
- **Atomic Effects**: Each verb performs single, well-defined business action
- **Variable Scoping**: Sophisticated variable system for multi-currency, multi-class scenarios
- **Provenance Tracking**: Every DSL operation generates audit events

## Domain Model

### Core Entities

#### Fund Structure
- **fund**: Legal entity, domicile, administration details
- **share_class**: Currency, dealing frequency, fee structure
- **series**: Optional equalisation grouping for tax efficiency

#### Investor Identity
- **investor**: Legal identity (individual, corporate, trust, nominee)
- **beneficial_owner**: Ultimate beneficial ownership with PEP/sanctions flags
- **kyc_profile**: Risk rating, screening results, refresh schedule
- **tax_profile**: FATCA/CRS classification, withholding rates

#### Operational Data
- **bank_instruction**: Multi-currency settlement instructions with versioning
- **trade**: Subscription/redemption orders with NAV allocation
- **register_lot**: Current unit holdings by investor/class/series
- **register_event**: Immutable record of all unit movements

#### Governance
- **lifecycle_state**: Current investor state with transition history
- **document**: Evidence storage with expiry tracking
- **audit_event**: Complete action log for compliance

### Entity Relationships

```
fund (1) ——→ (n) share_class (1) ——→ (n) series
                     ↓
investor (1) ——→ (n) register_lot ——← (1) share_class
    ↓               ↓
    ├── beneficial_owner (n)
    ├── kyc_profile (1)
    ├── tax_profile (1)
    ├── bank_instruction (n)
    ├── trade (n)
    └── document (n)

register_lot (1) ——→ (n) register_event ——← (1) trade
```

## Lifecycle State Machine

### States
```
OPPORTUNITY → PRECHECKS → KYC_PENDING → KYC_APPROVED →
SUB_PENDING_CASH → FUNDED_PENDING_NAV → ISSUED → ACTIVE →
REDEEM_PENDING → REDEEMED → OFFBOARDED
```

### State Transitions & Guards

| From State | To State | Trigger | Guard Conditions |
|------------|----------|---------|------------------|
| OPPORTUNITY | PRECHECKS | `investor.record-indication` | NDA signed, indication recorded |
| PRECHECKS | KYC_PENDING | `kyc.begin` | Initial docs submitted |
| KYC_PENDING | KYC_APPROVED | `kyc.approve` | All docs verified, screening passed |
| KYC_APPROVED | SUB_PENDING_CASH | `subscribe.request` | Valid subscription order |
| SUB_PENDING_CASH | FUNDED_PENDING_NAV | `cash.confirm` | Settlement funds received |
| FUNDED_PENDING_NAV | ISSUED | `subscribe.issue` | NAV struck, units allocated |
| ISSUED | ACTIVE | Automatic | First allocation complete |
| ACTIVE | REDEEM_PENDING | `redeem.request` | Valid redemption notice |
| REDEEM_PENDING | REDEEMED | `redeem.settle` | All units redeemed, cash paid |
| REDEEMED | OFFBOARDED | `offboard.close` | Final documentation complete |

## Data Dictionary

### Core Tables

#### investor
```sql
investor_id UUID PK
type TEXT CHECK IN ('INDIVIDUAL','CORPORATE','TRUST','FOHF','NOMINEE')
legal_name TEXT NOT NULL
lei TEXT NULL
registration_number TEXT NULL
domicile TEXT NOT NULL
address_line1..4 TEXT, city TEXT, postal_code TEXT, country TEXT
status TEXT CHECK IN ('OPPORTUNITY','KYC_PENDING','APPROVED','ACTIVE','REDEEMED','OFFBOARDED')
created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
```

#### register_lot (Current Holdings)
```sql
lot_id UUID PK
investor_id UUID FK → investor(investor_id)
class_id UUID FK → share_class(class_id)
series_id UUID FK → series(series_id) NULL
units NUMERIC(28,10) NOT NULL -- Current holdings
first_trade_date DATE NOT NULL
last_activity_at TIMESTAMPTZ
```

#### register_event (Immutable Unit Movements)
```sql
event_id UUID PK
lot_id UUID FK → register_lot(lot_id)
event_type TEXT CHECK IN ('ISSUE','REDEEM','TRANSFER_IN','TRANSFER_OUT','CORP_ACTION')
event_ts TIMESTAMPTZ NOT NULL
delta_units NUMERIC(28,10) NOT NULL -- +/- unit change
nav_per_share NUMERIC(18,8) NULL
source_trade_id UUID FK → trade(trade_id) NULL
note TEXT
```

#### trade (Orders & Allocations)
```sql
trade_id UUID PK
investor_id UUID FK → investor(investor_id)
class_id UUID FK → share_class(class_id)
type TEXT CHECK IN ('SUB','RED','TRANSFER_IN','TRANSFER_OUT','SWITCH_IN','SWITCH_OUT')
status TEXT CHECK IN ('PENDING','ALLOCATED','SETTLED','CANCELLED')
trade_date DATE NOT NULL
nav_date DATE NULL
nav_per_share NUMERIC(18,8) NULL
units NUMERIC(28,10) NULL
gross_amount NUMERIC(28,2) NULL
fees_amount NUMERIC(28,2) DEFAULT 0
currency TEXT NOT NULL
fx_rate NUMERIC(18,8) DEFAULT 1
bank_id UUID FK → bank_instruction(bank_id) NULL
idempotency_key TEXT UNIQUE -- Prevents duplicate trades
```

## DSL Specification

### Variable System

#### Investor Variables
- `?INV.LEGAL_NAME` → investor.legal_name
- `?INV.TYPE` → investor.type ('INDIVIDUAL', 'CORPORATE', etc.)
- `?INV.DOMICILE` → investor.domicile
- `?INV.ADDRESS.LINE1` → investor.address_line1
- `?INV.LEI` → investor.lei

#### KYC Variables
- `?KYC.RISK` → kyc_profile.risk_rating ('LOW', 'MEDIUM', 'HIGH')
- `?KYC.STATUS` → kyc_profile.status
- `?KYC.REFRESH_DUE` → kyc_profile.refresh_due_at

#### Tax Variables
- `?TAX.FATCA_CLASS` → tax_profile.fatca_class
- `?TAX.CRS_CLASS` → tax_profile.crs_class
- `?TAX.FORM_TYPE` → tax_profile.form_type

#### Banking Variables (Multi-Currency)
- `?BANK[USD].IBAN` → bank_instruction.iban WHERE currency='USD'
- `?BANK[EUR].SWIFT` → bank_instruction.swift_bic WHERE currency='EUR'

#### Trading Variables
- `?TRADE.AMOUNT` → trade.gross_amount
- `?TRADE.UNITS` → trade.units
- `?TRADE.NAV_DATE` → trade.nav_date
- `?TRADE.NAV_PERSHARE` → trade.nav_per_share

### DSL Verbs

#### Opportunity Management
```lisp
(investor.start-opportunity
  :legal-name ?INV.LEGAL_NAME
  :type ?INV.TYPE
  :domicile ?INV.DOMICILE)

(investor.record-indication
  :investor ?INV.ID
  :fund ?FUND.ID
  :class ?CLASS.ID
  :ticket ?TRADE.AMOUNT)
```

#### KYC/KYB Process
```lisp
(kyc.begin :investor ?INV.ID)

(kyc.collect-doc
  :investor ?INV.ID
  :doc-type "passport"
  :subject "primary_signatory")

(kyc.screen
  :investor ?INV.ID
  :provider "worldcheck")

(kyc.approve
  :investor ?INV.ID
  :risk ?KYC.RISK
  :refresh-due ?KYC.REFRESH_DUE)
```

#### Tax & Banking Setup
```lisp
(tax.capture
  :investor ?INV.ID
  :fatca ?TAX.FATCA_CLASS
  :crs ?TAX.CRS_CLASS
  :form ?TAX.FORM_TYPE)

(bank.set-instruction
  :investor ?INV.ID
  :currency "USD"
  :iban ?BANK[USD].IBAN
  :swift ?BANK[USD].SWIFT)
```

#### Subscription Workflow
```lisp
(subscribe.request
  :investor ?INV.ID
  :class ?CLASS.ID
  :amount ?TRADE.AMOUNT
  :trade-date ?DATE.TRADE
  :ccy "USD")

(cash.confirm
  :investor ?INV.ID
  :amount ?TRADE.AMOUNT
  :value-date ?DATE.TRADE
  :bank-currency "USD")

(deal.nav
  :fund ?FUND.ID
  :nav-date ?TRADE.NAV_DATE)

(subscribe.issue
  :investor ?INV.ID
  :class ?CLASS.ID
  :series ?SERIES.ID
  :nav-per-share ?TRADE.NAV_PERSHARE
  :units ?TRADE.UNITS)
```

#### Ongoing Operations
```lisp
(kyc.refresh-schedule
  :investor ?INV.ID
  :frequency "ANNUAL"
  :next ?KYC.REFRESH_DUE)

(screen.continuous
  :investor ?INV.ID
  :frequency "WEEKLY")
```

#### Redemption & Offboarding
```lisp
(redeem.request
  :investor ?INV.ID
  :class ?CLASS.ID
  :units ?TRADE.UNITS
  :notice-date ?DATE.NOTICE)

(redeem.settle
  :investor ?INV.ID
  :amount ?TRADE.AMOUNT
  :settle-date ?DATE.SETTLE)

(offboard.close :investor ?INV.ID)
```

## Implementation Architecture

### Module Structure
```
hedge-fund-investor/
├── internal/
│   ├── domain/           # Core domain entities
│   ├── events/           # Event sourcing infrastructure
│   ├── dsl/             # DSL parser and executor
│   ├── state/           # Lifecycle state machine
│   ├── store/           # Data persistence layer
│   └── compliance/      # KYC/AML/Tax utilities
├── sql/
│   ├── migrations/      # Database schema
│   └── views/           # Register and reporting views
├── dsl/
│   ├── vocab/           # DSL verb definitions
│   └── examples/        # Sample workflows
└── docs/
    ├── api/             # API documentation
    └── compliance/      # Regulatory guidance
```

### Key Components

#### DSL Engine
- **Parser**: S-expression to AST conversion
- **Executor**: Verb execution with database effects
- **Variable Resolver**: Dynamic variable substitution
- **Idempotency Manager**: Duplicate operation prevention

#### Event Sourcing
- **Event Store**: Immutable event persistence
- **Projections**: Derived state calculations (register_lot)
- **Snapshots**: Performance optimization for large histories
- **Replay**: Historical state reconstruction

#### State Machine
- **State Manager**: Current state tracking
- **Transition Engine**: Guard condition validation
- **Journal**: State change history
- **Notification**: State change events

#### Compliance Engine
- **KYC Workflow**: Document collection and verification
- **Screening Integration**: PEP/sanctions checking
- **Tax Classification**: FATCA/CRS determination
- **Reporting**: Regulatory report generation

## Integration Points

### External Systems
- **Fund Administration Platform**: NAV imports, trade confirmations
- **KYC Providers**: Document verification, screening services
- **Banking**: SWIFT messaging, payment confirmations
- **Regulatory**: AIFMD/MiFID II reporting
- **Tax Authorities**: CRS/FATCA reporting

### API Design
- **REST APIs**: Standard CRUD operations
- **DSL Endpoint**: Direct S-expression execution
- **Webhook Support**: Event notifications
- **Bulk Operations**: Large dataset imports
- **Reporting APIs**: Register extracts and analytics

## Security & Compliance

### Data Protection
- **Encryption**: PII encrypted at rest and in transit
- **Access Control**: Role-based permissions
- **Audit Logging**: Complete action trail
- **Data Retention**: Configurable retention policies

### Regulatory Compliance
- **GDPR**: Right to erasure with event anonymization
- **AIFMD**: Investor reporting and due diligence
- **MiFID II**: Best execution and transaction reporting
- **AML**: Continuous monitoring and suspicious activity reporting

## Performance Considerations

### Scalability
- **Read Replicas**: Separate read/write workloads
- **Event Partitioning**: Shard by investor ID
- **Materialized Views**: Pre-computed register snapshots
- **Caching**: Frequently accessed data caching

### Optimization
- **Bulk Processing**: Batch event insertion
- **Async Operations**: Non-blocking KYC processing
- **Index Strategy**: Optimized query patterns
- **Archive Strategy**: Historical data management

## Future Enhancements

### Advanced Features
- **Multi-Fund Management**: Cross-fund transfers and allocations
- **Institutional Platforms**: Master-feeder structures
- **Digital Onboarding**: API-driven investor portals
- **Real-time Analytics**: Live position and performance monitoring

### Technology Evolution
- **Blockchain Integration**: Immutable audit trails
- **AI/ML**: Automated KYC processing
- **Cloud Native**: Kubernetes deployment
- **Event Streaming**: Kafka-based event distribution

---

## Next Steps

1. **SQL Schema Definition**: Complete DDL with all constraints and indexes
2. **JSON DSL Vocabulary**: Formal verb and variable specifications
3. **Implementation Plan**: Phased development approach
4. **Proof of Concept**: Core workflow demonstration
5. **Integration Design**: External system interfaces

This design provides a solid foundation for a production-grade hedge fund investor management system that can scale from startup funds to large institutional asset managers while maintaining full regulatory compliance and operational flexibility.