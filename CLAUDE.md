# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**DSL Onboarding POC** is a Go-based proof-of-concept for a client onboarding Domain-Specific Language (DSL) system. It implements an immutable, versioned state machine that tracks client onboarding progression through stages while generating S-expression DSL output.

## Core Architecture

**Event Sourcing Pattern**: Uses immutable versioning where each state change creates a new database record rather than updating existing ones. This provides complete audit trails and ability to reconstruct any historical state.

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

## Pending Tasks (Deferred to Next Session)

**DSL CRUD Operations Enhancement**: The onboarding DSL is the key artifact of this POC. Current implementation has temporary workarounds that need to be completed:

1. **Update DSL functions to use DataStore interface**: Functions like `PopulateAttributeValues` currently expect concrete store types
2. **Complete attribute resolution workflow**: The `populate-attributes` and `get-attribute-values` commands need full DataStore integration
3. **Implement missing DataStore methods**: Some operations like `GetAttributesForDictionaryGroup` are commented out
4. **Enhance mock data error handling**: Improve graceful handling for missing mock data files
5. **Complete integration test refactoring**: Update skipped tests to work with DataStore interface injection

These tasks are critical for the full onboarding workflow but were deferred to focus on completing the DataStore interface abstraction successfully.