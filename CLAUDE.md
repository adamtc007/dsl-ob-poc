# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**DSL Onboarding POC** is a Go-based proof-of-concept for a client onboarding Domain-Specific Language (DSL) system. It implements an immutable, versioned state machine that tracks client onboarding progression through stages while generating S-expression DSL output.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ (for `greenteagc` garbage collector)
- PostgreSQL database
- (Optional) Google Gemini API key for AI-assisted KYC discovery

### Setup & First Run
```bash
# 1. Set database connection
export DB_CONN_STRING="postgres://localhost:5432/postgres?sslmode=disable"

# 2. Build the application
make build-greenteagc

# 3. Initialize database schema
make init-db

# 4. Seed with catalog data
./dsl-poc seed-catalog

# 5. Create your first onboarding case
./dsl-poc create --cbu="CBU-1234" --nature-purpose="UCITS equity fund"

# 6. Add products
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"

# 7. View the accumulated DSL history
./dsl-poc history --cbu="CBU-1234"
```

### Environment Variables
- **`DB_CONN_STRING`** (required) - PostgreSQL connection string
  - Format: `postgres://user:password@host:port/database?sslmode=disable`
  - Example: `postgres://localhost:5432/postgres?sslmode=disable`

- **`GEMINI_API_KEY`** (optional) - Google Gemini API key for AI-assisted operations
  - Required for: `discover-kyc`, `discover-services`, `discover-resources` commands
  - Get key from: https://ai.google.dev/
  - If not provided, commands gracefully skip AI generation

## 🏗️ Core Architectural Pattern: **DSL-as-State**

**This is the fundamental pattern that makes the entire system work.**

### What is DSL-as-State?

The DSL is not just a representation of state—**the DSL IS the state itself**. This is the key architectural insight:

- **State = Accumulated DSL Document**: Each onboarding case's current state is represented by its complete, accumulated DSL document
- **Immutable Event Sourcing**: Each command (create, add-products, discover-kyc, etc.) appends to the DSL, creating a new version
- **Executable Documentation**: The DSL is simultaneously:
  - Human-readable documentation of what happened
  - Machine-parseable structured data
  - Complete audit trail for compliance
  - State reconstruction mechanism
  - Executable workflow definition

### Why This Works for Onboarding/KYC/Investor Management

1. **Compositional State Building**: Each operation builds on previous DSL
   ```lisp
   ;; Version 1: Create case
   (case.create (cbu.id "CBU-1234") (nature-purpose "UCITS fund"))
   
   ;; Version 2: Add products (accumulated)
   (case.create (cbu.id "CBU-1234") (nature-purpose "UCITS fund"))
   (products.add "CUSTODY" "FUND_ACCOUNTING")
   
   ;; Version 3: Discover KYC (accumulated)
   (case.create (cbu.id "CBU-1234") (nature-purpose "UCITS fund"))
   (products.add "CUSTODY" "FUND_ACCOUNTING")
   (kyc.start (documents (document "CertificateOfIncorporation")) ...)
   ```

2. **Complete Audit Trail**: Required for regulatory compliance—every decision is captured in the DSL

3. **State Inspection at Any Point**: Parse the DSL to understand current onboarding stage

4. **Time Travel**: Access any historical version to see state at that moment

5. **Workflow Coordination**: Multiple systems can consume the same DSL to execute their parts

6. **Validation & Testing**: Can validate entire workflows by parsing accumulated DSL

### Implementation Examples

**Main Onboarding POC** (`internal/cli/`):
- `create` → Initial DSL version
- `add-products` → Appends to DSL
- `discover-kyc` → Appends KYC requirements
- `discover-services` → Appends service plan
- `discover-resources` → Appends resource plan
- `history` → Shows DSL evolution over time

Each command appends to the accumulated DSL document, creating new versions while maintaining complete audit trail of all onboarding decisions.

### Key Architectural Benefits

✅ **Immutability**: DSL versions never change, only accumulate
✅ **Traceability**: Complete history of all decisions
✅ **Composability**: Build complex state from simple operations
✅ **Declarative**: What to do, not how to do it
✅ **Testable**: Can verify workflows by parsing DSL
✅ **Extensible**: Add new verbs without breaking existing DSL
✅ **Human-Readable**: Business users can review the DSL
✅ **Machine-Executable**: Systems can parse and execute

### Critical Implementation Details

1. **Verb Validation**: Only approved DSL verbs are allowed (prevents AI hallucination)
2. **UUID Resolution**: Placeholders like `<investor_id>` resolved to actual UUIDs
3. **Context Maintenance**: Session/case context tracks entities across operations
4. **Accumulation**: Each operation appends, never replaces
5. **Versioning**: Each append creates new database version with full DSL

## ✅ Recent Implementation Updates

### DSL Verb Validation (Completed)
**Problem**: AI agents could generate unapproved DSL verbs, leading to hallucinated operations.

**Solution**: Implemented verb validation system:
- **Main Onboarding POC** (`internal/agent/dsl_agent.go`): Added `validateDSLVerbs()` function with 70+ approved verbs
- Validation occurs after AI generation, before DSL is stored
- System prompt explicitly lists approved verbs as constraints
- Comprehensive test coverage (`internal/agent/dsl_agent_test.go`)

**Impact**: Prevents AI from inventing operations, ensures DSL correctness, maintains domain vocabulary integrity.

### Testing & Verification
- ✅ Verb validation: 20+ test cases covering all 70+ approved verbs
- ✅ DSL accumulation: Complete workflow tested (create → add-products → discover-kyc → discover-services → discover-resources)
- ✅ State machine progression validated across all commands
- ✅ Event sourcing and versioning operational

---

## 🔑 Second Core Pattern: **AttributeID-as-Type**

**Variables in the DSL are typed by their attributeID (UUID), not by primitive types.**

### The Pattern

S-expression structure:
```lisp
(verb attributeID attributeID attributeID ...)
```

Where each **attributeID** is a **UUID** that references the **dictionary table** (universal schema).

### Example: Hedge Fund Investor

```lisp
(investor.start-opportunity
  @attr{uuid-0001}  ; → investor.legal_name (string, PII)
  @attr{uuid-0002}  ; → investor.type (enum: INDIVIDUAL|CORPORATE|TRUST)
  @attr{uuid-0003}  ; → investor.domicile (string, ISO country code)
)

(kyc.begin
  @attr{uuid-0001}  ; → Same investor.legal_name
  @attr{uuid-0004}  ; → kyc.risk_rating (enum: LOW|MEDIUM|HIGH)
)
```

### Example: Main Onboarding POC

```lisp
(resources.plan
  (resource.create "CustodyAccount"
    (owner "CustodyTech")
    (var (attr-id "8a5d1a77-..."))  ; → custody.account_number
  )
)

(values.bind
  (bind (attr-id "8a5d1a77-...") (value "CUSTODY-ACC-001"))
)
```

### Why AttributeID-as-Type is Powerful

1. **Metadata-Driven Type System**: All type information lives in the dictionary table
   - Data type (string, number, date, enum)
   - Validation rules
   - Privacy classification (PII, PCI, PHI)
   - Allowed values for enums
   - Source metadata (where to get the value)
   - Sink metadata (where to store the value)

2. **Late Binding**: Values can be resolved at different times from different sources
   ```
   Time 1: DSL declares (var (attr-id "uuid")) → placeholder
   Time 2: Value bound from user input → "John Smith"
   Time 3: Value enriched from CRM system → additional metadata
   ```

3. **Universal Data Contract**: AttributeID is the agreement between all systems
   - Frontend knows to collect attr-id "xyz" with specific validation
   - Backend knows to store attr-id "xyz" in specific database column
   - Compliance knows attr-id "xyz" is PII and must be encrypted
   - Analytics knows attr-id "xyz" must be masked in reports

4. **Data Governance Built-In**: Privacy and compliance at the type level
   ```sql
   SELECT attribute_id, name, mask, domain, source, sink
   FROM dictionary
   WHERE attribute_id = '8a5d1a77-...'
   -- Returns: custody.account_number, 'string', 'Settlement', {...}
   ```

5. **Cross-System Semantic Interoperability**: Different systems agree on meaning
   - Not just "this is a string" but "this is an investor legal name"
   - Not just "this is a number" but "this is a subscription amount in base currency"
   - Semantic type system vs syntactic type system

6. **Versioning and Evolution**: Attribute definitions can evolve without breaking DSL
   - Add new validation rules → existing DSL still valid
   - Change source system → DSL references unchanged
   - Migrate data stores → attributeID remains constant

### Dictionary Table Structure

```sql
CREATE TABLE "dsl-ob-poc".dictionary (
    attribute_id UUID PRIMARY KEY,           -- The "type" identifier
    name VARCHAR(255) NOT NULL UNIQUE,       -- Human-readable name
    long_description TEXT,                   -- For AI discovery
    group_id VARCHAR(100),                   -- Logical grouping (KYC, Settlement, etc.)
    mask VARCHAR(50) DEFAULT 'string',       -- Data type/format
    domain VARCHAR(100),                     -- Business domain
    vector TEXT,                             -- For AI semantic search
    source JSONB,                            -- SourceMetadata: where/how to get value
    sink JSONB,                              -- SinkMetadata: where/how to store value
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

### How It Works Together

1. **DSL declares variables** by attributeID:
   ```lisp
   (resources.plan
     (resource.create "CustodyAccount"
       (var (attr-id "8a5d1a77-..."))))
   ```

2. **Dictionary defines the attribute**:
   ```json
   {
     "attribute_id": "8a5d1a77-...",
     "name": "custody.account_number",
     "mask": "string",
     "domain": "Settlement",
     "source": {"type": "generated", "format": "CUST-{sequence}"},
     "sink": {"table": "accounts", "column": "account_number"}
   }
   ```

3. **Runtime resolution** binds the value:
   ```lisp
   (values.bind
     (bind (attr-id "8a5d1a77-...") (value "CUST-000123")))
   ```

4. **Systems consume** using attributeID as the contract:
   - DSL executor knows what to validate (from dictionary.mask)
   - Data collector knows where to get value (from dictionary.source)
   - Data persister knows where to store (from dictionary.sink)
   - Compliance knows how to protect (from dictionary metadata)

### Comparison to Traditional Approaches

| Traditional | AttributeID-as-Type |
|-------------|---------------------|
| `string accountNumber` | `@attr{uuid} → dictionary → "custody.account_number"` |
| Type = syntax (string) | Type = semantics (what it means) |
| Validation in code | Validation in dictionary metadata |
| Privacy in separate system | Privacy in attribute definition |
| Hard-coded sources | Source metadata in dictionary |
| Schema changes break code | Dictionary evolution, DSL stable |

### Real-World Benefits

✅ **Single Source of Truth**: Dictionary is the universal schema  
✅ **AI-Friendly**: LLMs can discover attributes by description  
✅ **Compliance-Ready**: Privacy flags embedded in type system  
✅ **Multi-Source**: Same attribute can come from different sources  
✅ **Auditable**: Complete provenance tracking via source metadata  
✅ **Evolvable**: Change implementation without changing DSL  
✅ **Cross-System**: All systems speak the same "type language"  

### The Two Patterns Together

**DSL-as-State** + **AttributeID-as-Type** = Complete Onboarding System

```
State = Accumulated DSL Document
DSL = S-expressions of (verb attributeID attributeID ...)
AttributeID = UUID → Dictionary (universal schema)
Dictionary = Metadata-driven type system with governance

Result: Self-describing, evolvable, auditable, compliant state machine
```

### Concrete Example: Client Onboarding Workflow

**Command 1: Create case**
```bash
./dsl-poc create --cbu="CBU-1234" --nature-purpose="UCITS equity fund domiciled in LU"
```

**System generates DSL (State Version 1)**:
```lisp
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)
```

**Database State**: Version 1 stored with `cbu_id="CBU-1234"`, `version=1`

---

**Command 2: Add products**
```bash
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"
```

**System accumulates DSL (State Version 2)**:
```lisp
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)

(products.add "CUSTODY" "FUND_ACCOUNTING")
```

**Database State**: Version 2 stored - complete DSL includes both operations

---

**Command 3: Discover KYC requirements**
```bash
./dsl-poc discover-kyc --cbu="CBU-1234"
```

**System accumulates DSL (State Version 3)** - AI agent analyzes context and appends:
```lisp
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)

(products.add "CUSTODY" "FUND_ACCOUNTING")

(kyc.start
  (documents
    (document "CertificateOfIncorporation")
    (document "ArticlesOfAssociation")
  )
  (jurisdictions
    (jurisdiction "LU")
  )
)
```

**Key observations**:
1. **State accumulates** - each operation appends to DSL
2. **Complete audit trail** - entire onboarding history visible in one document
3. **AttributeIDs provide semantics** - `(var (attr-id "..."))` references dictionary for type information
4. **Immutable versioning** - all three versions preserved in database
5. **State reconstruction** - can recreate state at any point by parsing DSL
6. **AI integration** - Gemini generates valid DSL using approved verbs only

This is how **DSL-as-State + AttributeID-as-Type** enables auditable, compliant onboarding workflows!

---

## 🎯 Why This Architecture Works

The combination of **DSL-as-State** + **AttributeID-as-Type** solves fundamental problems in financial onboarding:

### Traditional Problems Solved

| Traditional Approach | DSL-as-State Solution |
|---------------------|----------------------|
| **State scattered across tables** | State = one DSL document |
| **Audit trail requires event logging** | DSL IS the audit trail |
| **Hard to reconstruct past state** | Parse any DSL version |
| **Type info in code** | Type info in dictionary (metadata) |
| **Validation in multiple places** | Validation via approved verbs |
| **Privacy handled separately** | Privacy in attribute definition |
| **Data contracts break easily** | AttributeID is stable contract |
| **Workflow coordination complex** | DSL is universal language |

### Key Benefits Realized

1. **Regulatory Compliance**: Complete, immutable audit trail required by financial regulators
2. **AI Integration**: Structured DSL enables AI agents to participate in workflows safely (with verb validation)
3. **Cross-System Coordination**: Multiple systems consume same DSL document
4. **Human Readability**: Business users can review and understand DSL
5. **Machine Executability**: Systems can parse and execute DSL operations
6. **Evolvability**: Dictionary can evolve without breaking existing DSL
7. **Data Governance**: Privacy, classification, validation all metadata-driven
8. **Time Travel**: Access any historical state by version number

### This Is Not Just Another DSL

This is a **state representation language** where:
- The language IS the state
- Types ARE semantic identifiers
- Execution IS state transitions
- History IS version accumulation
- Compliance IS inherent in design

**This is the architectural foundation that makes sophisticated financial onboarding workflows tractable, auditable, and AI-enabled.**

## Core Architecture Patterns

**DSL-as-State (Primary Pattern)**: The accumulated DSL document IS the state. Each operation appends to the DSL, creating a new version. State is reconstructed by parsing the DSL. See "Core Architectural Pattern" section above.

**Event Sourcing Pattern**: Uses immutable versioning where each state change creates a new database record rather than updating existing ones. This provides complete audit trails and ability to reconstruct any historical state. The DSL itself acts as the event log.

**State Machine Progression**:
1. **CREATE** (`create` command) - Initial case creation with CBU ID
2. **ADD_PRODUCTS** (`add-products` command) - Append products to existing case
3. **DISCOVER_KYC** (`discover-kyc` command) - AI-assisted KYC discovery using Gemini
4. **DISCOVER_SERVICES** (`discover-services` command) - Service discovery and planning
5. **DISCOVER_RESOURCES** (`discover-resources` command) - Resource discovery and planning

## Development Commands

**Build** (uses experimental `greenteagc` GC for 60% better pause times):
```bash
make build-greenteagc    # Preferred build method
./build.sh              # Alternative script-based build
make test               # Run all tests
make test-coverage      # Generate coverage report
make lint               # Run golangci-lint with 20+ linters
```

**Database Setup**:
```bash
export DB_CONN_STRING="postgres://localhost:5432/postgres?sslmode=disable"
make init-db            # Initialize schema and tables
./dsl-poc seed-catalog  # Populate with mock product/service data
```

**Development Workflow**:
```bash
./dsl-poc create --cbu="CBU-1234"
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"
./dsl-poc discover-kyc --cbu="CBU-1234"  # Requires GEMINI_API_KEY
./dsl-poc discover-services --cbu="CBU-1234"  # Service discovery and planning
./dsl-poc discover-resources --cbu="CBU-1234"  # Resource discovery and planning
./dsl-poc history --cbu="CBU-1234"       # View complete DSL evolution
```

## Key Architecture Components

**Database Schema** (`sql/init.sql`):
- `dsl_ob` - Immutable versioned DSL records (event sourcing core)
- `products`, `services`, `prod_resources` - Catalog tables
- `attributes`, `dictionaries` - Data classification with privacy flags
- Uses `"dsl-ob-poc"` schema with UUID primary keys

**Package Structure**:
- `internal/cli/` - Command implementations for state machine operations
- `internal/store/` - PostgreSQL operations with comprehensive error handling
- `internal/dsl/` - S-expression builders and parsers
- `internal/agent/` - Gemini AI integration for KYC discovery
- `internal/mocks/` - Test data generators

**AI Integration** (`internal/agent/agent.go`):
- Uses Google Gemini 2.5 Flash for KYC requirement discovery
- Structured JSON responses parsed into DSL
- Graceful fallback when API key not provided
- Safety settings configured to avoid blocking

## DSL Format

S-expressions with nested structure representing onboarding progression:

```lisp
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)

(products.add "CUSTODY" "FUND_ACCOUNTING")

(kyc.start
  (documents
    (document "CertificateOfIncorporation")
  )
  (jurisdictions
    (jurisdiction "LU")
  )
)
```

## Testing Strategy

- Comprehensive unit tests across all packages
- SQL operations mocked using `go-sqlmock`
- DSL generation and parsing tested with realistic scenarios
- CLI command logic tested with various input combinations
- Run single test: `go test -v ./internal/cli -run TestHistoryCommand`
- Run all tests: `make test`
- Generate coverage report: `make test-coverage`

## Performance Notes

**greenteagc Benefits**: 60% reduction in GC pause times, ~4% better throughput, more predictable latency for concurrent workloads (requires Go 1.21+).

**Database Optimizations**: Composite indexes on `(cbu_id, created_at DESC)` for fast latest lookups, soft deletes preserve data integrity, foreign key constraints with appropriate cascades.

## Code Quality

**Linting and Formatting**:
```bash
make lint               # Run golangci-lint with 20+ linters
make fmt                # Format code with gofmt
make vet                # Run go vet
make check              # Run fmt, vet, and lint (pre-commit check)
```

## CI/CD

GitHub Actions pipeline runs on Ubuntu with Go version from `go.mod`, caches modules and build artifacts, executes lint/build/test phases with 5-minute timeout.

## Known Limitations & Future Work

**DSL CRUD Operations Enhancement**: The onboarding DSL is the key artifact of this POC. Current implementation has temporary workarounds that need to be completed:

1. **Update DSL functions to use DataStore interface**: Functions like `PopulateAttributeValues` currently expect concrete store types
2. **Complete attribute resolution workflow**: The `populate-attributes` and `get-attribute-values` commands need full DataStore integration
3. **Implement missing DataStore methods**: Some operations like `GetAttributesForDictionaryGroup` are commented out
4. **Enhance mock data error handling**: Improve graceful handling for missing mock data files
5. **Complete integration test refactoring**: Update skipped tests to work with DataStore interface injection

These tasks are critical for the full onboarding workflow but were deferred to focus on completing the DataStore interface abstraction successfully.