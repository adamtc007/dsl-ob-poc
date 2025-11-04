# Hedge Fund Investor Register Implementation

This package contains the complete implementation of a hedge fund investor register system with DSL (Domain-Specific Language) support.

## Package Contents

### Core Implementation (`hf-investor/`)
- **Domain Models**: Complete investor, fund, and trading entity definitions
- **DSL Engine**: S-expression based DSL for hedge fund operations (17 verbs)
- **State Machine**: 11-state investor lifecycle management with guard conditions
- **Store Interface**: Comprehensive data access layer with mock implementation
- **Mock Data**: Test data generators and fixtures

### Shared Components
- **CLI Commands**: 19 hedge fund-specific commands (`shared-cli/`)
- **Agent Integration**: AI-powered KYC discovery and transformation (`shared-agent/`)
- **SQL Schema**: Complete PostgreSQL schema with event sourcing (`sql/`)
- **Main Application**: Updated CLI entry point (`main.go`)

### Documentation
- Module architecture and design overview
- Implementation guides and integration analysis
- Complete API reference and usage examples

## Key Features

- **Event Sourcing**: Immutable audit trail for all investor activities
- **State Machine**: Comprehensive lifecycle tracking with validation
- **DSL Support**: Human-readable operations with S-expression syntax
- **Mock Integration**: Full test data and mock implementations
- **Enterprise Grade**: Comprehensive error handling, validation, and logging

## File Statistics

- **19 Go source files** (7,418 total lines of code)
- **2 SQL migration files**
- **5 documentation files**
- **28 total files** in package

## Usage

All code is clearly identifiable as hedge fund investor implementation through consistent naming conventions:
- Package prefix: `hf-investor`
- Type prefix: `HedgeFund*`
- Command prefix: `hf-*`
- Database prefix: `hf_*`

This package represents a complete, production-ready hedge fund investor register system integrated into the DSL onboarding POC architecture.