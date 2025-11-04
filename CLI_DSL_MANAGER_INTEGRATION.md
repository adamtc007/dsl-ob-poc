# CLI DSL State Manager Integration Guide

This document outlines the systematic integration of DSL State Manager across all CLI commands to enforce the architectural constraint that ALL DSL state changes must flow through the DSL State Manager.

## Integration Status

### ✅ COMPLETED
- `get_attribute_values.go` - Fixed to use session manager for DSL accumulation
- `create.go` - Fixed to use session manager for initial DSL creation  
- `add_products.go` - Fixed to use session manager for product DSL accumulation

### ❌ PENDING INTEGRATION
- `populate_attributes.go` - Direct DSL manipulation via `AddPopulatedAttributes`
- `agent_transform.go` - Direct DSL insertion via `InsertDSLWithState`
- `discover_resources.go` - Direct DSL insertion via `InsertDSLWithState`
- `discover_services.go` - Direct DSL insertion via `InsertDSLWithState`
- `discover_kyc.go` - Direct DSL insertion via `InsertDSLWithState`

## Standard Integration Pattern

All CLI commands MUST follow this pattern:

```go
import (
    // ... other imports
    "dsl-ob-poc/internal/shared-dsl/session"
)

func RunCommand(ctx context.Context, ds datastore.DataStore, args []string) error {
    // 1. Parse flags and validate input
    
    // 2. Get current DSL state (if exists)
    currentDSL, err := ds.GetLatestDSL(ctx, *cbuID)
    if err != nil {
        return fmt.Errorf("failed to get current DSL: %w", err)
    }
    
    // 3. Create DSL session manager (SINGLE SOURCE OF TRUTH)
    sessionMgr := session.NewManager()
    dslSession := sessionMgr.GetOrCreate(*cbuID, "onboarding")
    
    // 4. Accumulate existing DSL
    err = dslSession.AccumulateDSL(currentDSL)
    if err != nil {
        return fmt.Errorf("failed to accumulate current DSL: %w", err)
    }
    
    // 5. Generate new DSL fragment using builder functions
    newFragment, err := dsl.BuildOperation(...) // Use DSL builders, NOT string manipulation
    if err != nil {
        return fmt.Errorf("failed to generate DSL: %w", err)
    }
    
    // 6. Accumulate new fragment through state manager
    err = dslSession.AccumulateDSL(newFragment)
    if err != nil {
        return fmt.Errorf("failed to accumulate new DSL: %w", err)
    }
    
    // 7. Get final DSL from state manager and save to database
    finalDSL := dslSession.GetDSL()
    versionID, err := ds.InsertDSLWithState(ctx, *cbuID, finalDSL, newState)
    if err != nil {
        return fmt.Errorf("failed to save DSL: %w", err)
    }
    
    // 8. Output result
    fmt.Printf("✅ Command completed successfully\n")
    fmt.Printf("📝 DSL version: %s\n", versionID)
    fmt.Println("---")
    fmt.Println(finalDSL)
    fmt.Println("---")
    
    return nil
}
```

## Specific Command Fixes Needed

### populate_attributes.go
**Current Violation:**
```go
finalDSL, err := dsl.AddPopulatedAttributes(currentDSL, populatedValues)
```

**Required Fix:**
```go
// Get current DSL
currentDSL, err := ds.GetLatestDSL(ctx, *cbuID)

// Create session manager
sessionMgr := session.NewManager()
dslSession := sessionMgr.GetOrCreate(*cbuID, "onboarding")

// Accumulate current DSL
err = dslSession.AccumulateDSL(currentDSL)

// Generate populated attributes DSL
populatedFragment, err := dsl.BuildPopulatedAttributes(populatedValues)

// Accumulate through state manager
err = dslSession.AccumulateDSL(populatedFragment)

// Get final DSL and save
finalDSL := dslSession.GetDSL()
versionID, err := ds.InsertDSL(ctx, *cbuID, finalDSL)
```

### agent_transform.go
**Current Violation:**
```go
versionID, err := ds.InsertDSLWithState(ctx, *cbuID, response.NewDSL, saveState)
```

**Required Fix:**
```go
// Get current DSL
currentDSL, err := ds.GetLatestDSL(ctx, *cbuID)

// Create session manager
sessionMgr := session.NewManager()
dslSession := sessionMgr.GetOrCreate(*cbuID, "onboarding")

// Accumulate current DSL
err = dslSession.AccumulateDSL(currentDSL)

// Accumulate transformed DSL from agent
err = dslSession.AccumulateDSL(response.NewDSL)

// Get final DSL and save
finalDSL := dslSession.GetDSL()
versionID, err := ds.InsertDSLWithState(ctx, *cbuID, finalDSL, saveState)
```

### discover_*.go Commands
**Current Violation:**
```go
versionID, err := ds.InsertDSLWithState(ctx, *cbuID, newDSL, newState)
```

**Required Fix Pattern:**
```go
// Get current DSL
currentDSL, err := ds.GetLatestDSL(ctx, *cbuID)

// Create session manager
sessionMgr := session.NewManager()
dslSession := sessionMgr.GetOrCreate(*cbuID, "onboarding")

// Accumulate current DSL
err = dslSession.AccumulateDSL(currentDSL)

// Generate discovery DSL using agent or builder
discoveryDSL, err := generateDiscoveryDSL(...)

// Accumulate discovery DSL
err = dslSession.AccumulateDSL(discoveryDSL)

// Get final DSL and save
finalDSL := dslSession.GetDSL()
versionID, err := ds.InsertDSLWithState(ctx, *cbuID, finalDSL, newState)
```

## Validation Requirements

After integration, ALL CLI commands MUST:

1. ✅ Import `"dsl-ob-poc/internal/shared-dsl/session"`
2. ✅ Create session manager: `sessionMgr := session.NewManager()`
3. ✅ Get/create session: `dslSession := sessionMgr.GetOrCreate(*cbuID, "onboarding")`
4. ✅ Accumulate DSL: `err = dslSession.AccumulateDSL(dslFragment)`
5. ✅ Get final DSL: `finalDSL := dslSession.GetDSL()`
6. ❌ NEVER directly manipulate DSL strings: `dsl = oldDSL + newDSL` (FORBIDDEN)
7. ❌ NEVER bypass state manager for DSL operations

## Testing Validation

Each command integration must be validated with:

1. **Build Test**: `go build ./internal/cli/`
2. **Unit Test**: Verify session manager is used
3. **Integration Test**: End-to-end command execution
4. **Audit Test**: No direct DSL string manipulation remains

## Enforcement

### Linting Rules (Future)
Add custom linters to detect:
- Direct DSL string concatenation: `dsl.*=.*+`
- Bypassing session manager: `InsertDSL.*without.*session`
- Missing session manager imports

### Code Review Checklist
- [ ] Session manager imported and used
- [ ] All DSL accumulation goes through `dslSession.AccumulateDSL()`
- [ ] Final DSL retrieved with `dslSession.GetDSL()`
- [ ] No direct string manipulation of DSL
- [ ] Proper error handling for state manager operations

## Implementation Priority

1. **HIGH**: `populate_attributes.go` - Core attribute binding functionality
2. **HIGH**: `agent_transform.go` - AI-generated DSL integration  
3. **MEDIUM**: `discover_kyc.go` - KYC workflow integration
4. **MEDIUM**: `discover_services.go` - Service discovery workflow
5. **MEDIUM**: `discover_resources.go` - Resource planning workflow

## Success Criteria

✅ **Integration Complete When:**
- All CLI commands use DSL State Manager
- Zero direct DSL string manipulation
- All tests pass
- Linter shows no DSL-related violations
- End-to-end workflows function correctly

🚨 **Critical**: This integration is MANDATORY for architectural compliance. 
No exceptions or workarounds are permitted. The DSL State Manager is the 
single source of truth for all DSL operations across the entire system.