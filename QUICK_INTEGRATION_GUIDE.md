# Quick Integration Guide - Hedge Fund Investor IR

## How to Wire Quickly

### Compile-time Check
Add `internal/ir` to your build; from your executor, do:

```go
package main

import (
    "log"
    "dsl-ob-poc/internal/ir"
)

func executeInvestorPlan(data []byte) error {
    // Parse the IR plan
    p, err := ir.ParsePlan(data)
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

    // Validate the plan
    if err := p.Validate(); err != nil {
        return fmt.Errorf("validation error: %w", err)
    }

    // Execute each step
    for i, s := range p.Steps {
        log.Printf("Executing step %d: %s", i, s.Op)

        if err := executeStep(s); err != nil {
            return fmt.Errorf("step %d (%s) failed: %w", i, s.Op, err)
        }
    }

    return nil
}

func executeStep(s ir.Step) error {
    switch s.Op {
    case ir.OpInvestorStartOpportunity:
        var a ir.InvestorStartOpportunityArgs
        if err := s.DecodeArgs(&a); err != nil {
            return err
        }
        return handleInvestorStartOpportunity(a)

    case ir.OpSubscribeRequest:
        var a ir.SubscribeRequestArgs
        if err := s.DecodeArgs(&a); err != nil {
            return err
        }
        return handleSubscribeRequest(a)

    case ir.OpKYCApprove:
        var a ir.KYCApproveArgs
        if err := s.DecodeArgs(&a); err != nil {
            return err
        }
        return handleKYCApprove(a)

    case ir.OpSubscribeIssue:
        var a ir.SubscribeIssueArgs
        if err := s.DecodeArgs(&a); err != nil {
            return err
        }
        return handleSubscribeIssue(a)

    // Add other cases as needed...
    default:
        return fmt.Errorf("unsupported operation: %s", s.Op)
    }
}
```

### Handler Implementation Examples

```go
func handleInvestorStartOpportunity(args ir.InvestorStartOpportunityArgs) error {
    // Resolve AttrRef values
    legalName, err := resolveStringOrAttrRef(args.LegalName)
    if err != nil {
        return fmt.Errorf("legal_name resolution: %w", err)
    }

    domicile, err := resolveStringOrAttrRef(args.Domicile)
    if err != nil {
        return fmt.Errorf("domicile resolution: %w", err)
    }

    // Create investor record
    investor := &Investor{
        Type:      args.Type,
        LegalName: legalName,
        Domicile:  domicile,
        Status:    "OPPORTUNITY",
    }

    // Insert into database
    return db.CreateInvestor(investor)
}

func handleSubscribeRequest(args ir.SubscribeRequestArgs) error {
    // Create trade record
    trade := &Trade{
        InvestorID: args.InvestorID,
        ClassID:    args.ClassID,
        Type:       "SUB",
        Status:     "PENDING",
        Amount:     args.Amount,
        TradeDate:  parseDate(args.TradeDate),
        Currency:   args.Currency,
    }

    // Insert trade with idempotency
    return db.CreateTrade(trade)
}

func handleKYCApprove(args ir.KYCApproveArgs) error {
    // Update KYC profile
    profile := &KYCProfile{
        InvestorID: args.InvestorID,
        Status:     "APPROVED",
        RiskRating: args.Risk,
        RefreshDue: parseDate(args.RefreshDue),
    }

    // Update database and trigger state transition
    if err := db.UpdateKYCProfile(profile); err != nil {
        return err
    }

    // Transition investor state
    return transitionInvestorState(args.InvestorID, "KYC_APPROVED")
}

func handleSubscribeIssue(args ir.SubscribeIssueArgs) error {
    // Create register event (event sourcing)
    event := &RegisterEvent{
        LotID:       findOrCreateLot(args.InvestorID, args.ClassID, args.SeriesID),
        EventType:   "ISSUE",
        DeltaUnits:  args.Units,
        NAVPerShare: args.NAVPerShare,
        NAVDate:     parseDate(*args.NAVDate),
    }

    // Insert event (triggers register_lot update)
    if err := db.CreateRegisterEvent(event); err != nil {
        return err
    }

    // Update trade status
    return db.UpdateTradeStatus(args.InvestorID, "SETTLED")
}
```

### AttrRef Resolution

```go
func resolveStringOrAttrRef(raw json.RawMessage) (string, error) {
    // Try direct string first
    var directString string
    if err := json.Unmarshal(raw, &directString); err == nil {
        return directString, nil
    }

    // Try AttrRef
    var attrRef ir.AttrRef
    if err := json.Unmarshal(raw, &attrRef); err == nil {
        return resolveAttribute(attrRef.ID)
    }

    return "", fmt.Errorf("value must be string or AttrRef")
}

func resolveAttribute(attrID string) (string, error) {
    // Implement your attribute resolution logic
    // e.g., lookup from context, database, external system
    switch attrID {
    case "INV.LEGAL_NAME":
        return getFromContext("investor_legal_name")
    case "INV.DOMICILE":
        return getFromContext("investor_domicile")
    default:
        return "", fmt.Errorf("unknown attribute: %s", attrID)
    }
}
```

### Database Integration

```go
// Example with PostgreSQL
func (db *DB) CreateInvestor(inv *Investor) error {
    query := `
        INSERT INTO register.investor (
            type, legal_name, domicile, status, created_at
        ) VALUES ($1, $2, $3, $4, NOW())
        RETURNING investor_id`

    return db.QueryRow(query,
        inv.Type, inv.LegalName, inv.Domicile, inv.Status,
    ).Scan(&inv.ID)
}

func (db *DB) CreateRegisterEvent(event *RegisterEvent) error {
    query := `
        INSERT INTO register.register_event (
            lot_id, event_type, delta_units, nav_per_share, nav_date
        ) VALUES ($1, $2, $3, $4, $5)`

    _, err := db.Exec(query,
        event.LotID, event.EventType, event.DeltaUnits,
        event.NAVPerShare, event.NAVDate,
    )
    return err
}
```

### Error Handling Patterns

```go
// Step-level error wrapping
func executeStepWithRecovery(s ir.Step) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("step %s panicked: %v", s.Op, r)
        }
    }()

    // Log step execution
    log.Printf("Executing %s with idempotency key: %s",
        s.Op, safeString(s.IdempotencyKey))

    return executeStep(s)
}

// Idempotency handling
func handleWithIdempotency(key *string, fn func() error) error {
    if key != nil {
        if exists, err := checkIdempotencyKey(*key); err != nil {
            return err
        } else if exists {
            log.Printf("Skipping duplicate operation: %s", *key)
            return nil // Already processed
        }
    }

    err := fn()

    if err == nil && key != nil {
        recordIdempotencyKey(*key)
    }

    return err
}
```

### CLI Integration

```go
// Add to your main CLI
func RunExecuteIR(ctx context.Context, ds datastore.DataStore, args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("usage: execute-ir <plan.json>")
    }

    // Read IR file
    data, err := os.ReadFile(args[0])
    if err != nil {
        return fmt.Errorf("failed to read IR file: %w", err)
    }

    // Execute the plan
    executor := NewInvestorExecutor(ds)
    return executor.ExecutePlan(data)
}

type InvestorExecutor struct {
    db datastore.DataStore
}

func (e *InvestorExecutor) ExecutePlan(data []byte) error {
    plan, err := ir.ParsePlan(data)
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

    if err := plan.Validate(); err != nil {
        return fmt.Errorf("validation error: %w", err)
    }

    // Execute in transaction
    return e.db.WithTransaction(func(tx datastore.Transaction) error {
        for i, step := range plan.Steps {
            if err := e.executeStepInTx(tx, step); err != nil {
                return fmt.Errorf("step %d failed: %w", i, err)
            }
        }
        return nil
    })
}
```

### Testing Integration

```go
func TestInvestorWorkflow(t *testing.T) {
    // Load test IR
    data, err := os.ReadFile("testdata/corporate_subscription.json")
    require.NoError(t, err)

    // Parse and validate
    plan, err := ir.ParsePlan(data)
    require.NoError(t, err)
    require.NoError(t, plan.Validate())

    // Mock database
    db := &MockDataStore{}
    executor := NewInvestorExecutor(db)

    // Execute plan
    err = executor.ExecutePlan(data)
    require.NoError(t, err)

    // Verify results
    investor := db.GetInvestor("2fd4e5a7-3e84-4d2b-9f1a-7f6d0a9b1234")
    assert.Equal(t, "CORPORATE", investor.Type)
    assert.Equal(t, "ACTIVE", investor.Status)
}
```

## Quick Start Checklist

1. **✅ Add IR package to imports**
   ```go
   import "dsl-ob-poc/internal/ir"
   ```

2. **✅ Implement basic executor pattern**
   ```go
   plan, err := ir.ParsePlan(data)
   // Handle parse error
   err = plan.Validate()
   // Handle validation error
   for _, step := range plan.Steps { /* execute */ }
   ```

3. **✅ Add operation handlers**
   ```go
   switch step.Op {
   case ir.OpSubscribeRequest:
       var args ir.SubscribeRequestArgs
       step.DecodeArgs(&args)
       // Handle subscription
   }
   ```

4. **✅ Implement AttrRef resolution**
   ```go
   func resolveStringOrAttrRef(raw json.RawMessage) (string, error)
   ```

5. **✅ Add database integration**
   ```go
   // Map IR operations to database operations
   ```

6. **✅ Handle idempotency**
   ```go
   if step.IdempotencyKey != nil {
       // Check for duplicate execution
   }
   ```

7. **✅ Add error handling and logging**
   ```go
   log.Printf("Executing %s", step.Op)
   ```

This pattern gives you immediate compile-time safety, strong typing, and a clear execution model for the hedge fund investor lifecycle DSL.