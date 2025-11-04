# Fund Entity Resolution - Same Treatment as Investors

## Overview

**Fund entities** (and Share Classes) need the **exact same entity resolution treatment** as investors:
- Fuzzy matching for typos
- User confirmation for multiple matches
- Create new confirmation when not found
- UUID embedding in DSL

## Entity Types Requiring Resolution

| Entity Type | Table | ID Field | Name Field | Why Needed |
|-------------|-------|----------|------------|------------|
| **Investor** | `hf_investors` | `investor_id` | `legal_name` | User types "Alpine Capital" |
| **Fund** | `hf_funds` | `fund_id` | `fund_name` | User types "Global Equity Fund" |
| **Share Class** | `hf_share_classes` | `class_id` | `class_name` | User types "Class A" |
| **Beneficial Owner** | `hf_beneficial_owners` | `bo_id` | `full_name` | User types "John Smith" |

## Two-Stage KYC Model

### Stage 1: General Alternatives Clearance (No Fund)
```
User: "Create opportunity for Alpine Capital for alternatives"

Action: 
  - Resolve "Alpine Capital" → investor_id
  - No fund specified
  - General KYC clearance check
  
DSL Generated:
(kyc.begin
  (investor @attr{investor-id} = "f47ac10b-58cc-4372-a567-0e02b2c3d479")
  (tier "PRELIMINARY")
  (scope "ALTERNATIVES_GENERAL")
  (jurisdictions
    (jurisdiction "CH")))

Result: Investor cleared for alternative investments generally
```

**Purpose**: Initial red flag screening without specific fund commitment

### Stage 2: Fund-Specific Full KYC
```
User: "Start full KYC for Alpine Capital to invest in Global Equity Fund Class A"

Action:
  - Resolve "Alpine Capital" → investor_id
  - Resolve "Global Equity Fund" → fund_id  
  - Resolve "Class A" → class_id (for that fund)
  - Full KYC with all three entities

DSL Generated:
(kyc.begin
  (investor @attr{investor-id} = "f47ac10b-58cc-4372-a567-0e02b2c3d479")
  (fund @attr{fund-id} = "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
  (class @attr{class-id} = "b2c3d4e5-f6a7-8901-bcde-f12345678901")
  (tier "FULL")
  (scope "FUND_SPECIFIC")
  (jurisdictions
    (jurisdiction "CH")
    (jurisdiction "US")))

Result: Investor fully approved for this specific fund/class
```

**Purpose**: Complete due diligence for specific investment

## Multi-Entity Resolution Conversation Flow

### Example: Subscription with Multiple Entities

```
User: "Create subscription for Alpine Capital to Global Equity Fund Class I for 5 million USD"

System needs to resolve 3 entities:
  1. Investor: "Alpine Capital"
  2. Fund: "Global Equity Fund"  
  3. Share Class: "Class I"

Conversation:

🤖 Bot: 🔍 Resolving investor "Alpine Capital"...
       Found 2 matching investors:
       1. Alpine Capital AG (Switzerland)
       2. Alpine Capital LLC (US)
       Which one?

👤 User: "1"

🤖 Bot: ✅ Selected: Alpine Capital AG
       
       🔍 Resolving fund "Global Equity Fund"...
       Found 1 matching fund:
       1. Global Equity Fund (Domicile: LU, Currency: EUR)
       Confirm? (yes/no)

👤 User: "yes"

🤖 Bot: ✅ Selected: Global Equity Fund
       
       🔍 Resolving share class "Class I"...
       Found 2 classes for Global Equity Fund:
       1. Class I (EUR, Min: €1,000,000, Institutional)
       2. Class I-USD (USD, Min: $1,000,000, Institutional)
       Which one?

👤 User: "1"

🤖 Bot: ✅ All entities resolved!
       
       Creating subscription:
       - Investor: Alpine Capital AG (f47ac10b...)
       - Fund: Global Equity Fund (a1b2c3d4...)
       - Class: Class I (b2c3d4e5...)
       - Amount: 5,000,000.00 USD
       
       Confirm? (yes/no)

👤 User: "yes"

🤖 Bot: ✅ Subscription created successfully
       
       (subscribe.request
         (investor @attr{investor-id} = "f47ac10b-58cc-4372-a567-0e02b2c3d479")
         (fund @attr{fund-id} = "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
         (class @attr{class-id} = "b2c3d4e5-f6a7-8901-bcde-f12345678901")
         (amount 5000000.00)
         (currency "USD"))
       
       ✅ DSL Persisted to Database
```

## Implementation Architecture

### Generalized Entity Resolver

```go
type EntityResolver struct {
    dataStore EntityDataStore
}

type EntityType string
const (
    EntityTypeInvestor    EntityType = "investor"
    EntityTypeFund        EntityType = "fund"
    EntityTypeShareClass  EntityType = "share_class"
    EntityTypeBeneficiary EntityType = "beneficial_owner"
)

// Same interface for all entity types
func (r *EntityResolver) ResolveInvestor(ctx, searchTerm string) (*ResolutionResult, error)
func (r *EntityResolver) ResolveFund(ctx, searchTerm string) (*ResolutionResult, error)
func (r *EntityResolver) ResolveShareClass(ctx, fundID, className string) (*ResolutionResult, error)
```

### ResolutionResult

```go
type ResolutionResult struct {
    Resolved     bool                   // True if entity resolved to single UUID
    RequiresUser bool                   // True if user confirmation needed
    EntityType   EntityType             // Type of entity
    EntityID     string                 // Resolved UUID (if Resolved=true)
    Entity       map[string]interface{} // Full entity details
    
    // For user prompts
    Candidates        []EntityMatch // Matching entities with scores
    PromptMessage     string        // Message to show user
    PendingActionType string        // "select_fund", "confirm_create_fund", etc.
    
    // Metadata
    SearchTerm string  // Original search term
    Confidence float64 // Similarity score
    MatchType  string  // "exact", "fuzzy", "none"
}
```

### DataStore Interface Extensions

```go
type EntityDataStore interface {
    // Investors
    SearchInvestorsByName(ctx, name string) ([]map[string]interface{}, error)
    GetInvestorByID(ctx, investorID string) (map[string]interface{}, error)
    
    // Funds
    SearchFundsByName(ctx, name string) ([]map[string]interface{}, error)
    GetFundByID(ctx, fundID string) (map[string]interface{}, error)
    
    // Share Classes
    SearchShareClassesByName(ctx, fundID, className string) ([]map[string]interface{}, error)
    GetShareClassByID(ctx, classID string) (map[string]interface{}, error)
    
    // Beneficial Owners
    SearchBeneficialOwnersByName(ctx, investorID, name string) ([]map[string]interface{}, error)
}
```

## Fund Search SQL Examples

### Search Funds by Name
```sql
SELECT
    fund_id,
    fund_name,
    legal_name,
    fund_type,
    domicile,
    currency,
    status,
    inception_date,
    created_at
FROM "hf-investor".hf_funds
WHERE
    LOWER(fund_name) LIKE LOWER($1)
    OR LOWER(legal_name) LIKE LOWER($1)
ORDER BY
    -- Exact match first
    CASE WHEN LOWER(fund_name) = LOWER($2) THEN 0 ELSE 1 END,
    created_at DESC
LIMIT 10;
```

### Search Share Classes
```sql
SELECT
    c.class_id,
    c.fund_id,
    c.class_name,
    c.class_type,
    c.currency,
    c.min_initial_investment,
    c.management_fee_rate,
    f.fund_name,
    f.domicile as fund_domicile
FROM "hf-investor".hf_share_classes c
JOIN "hf-investor".hf_funds f ON f.fund_id = c.fund_id
WHERE
    c.fund_id = $1  -- Optional: filter by specific fund
    AND (
        LOWER(c.class_name) LIKE LOWER($2)
        OR LOWER(c.class_type) LIKE LOWER($2)
    )
ORDER BY
    CASE WHEN LOWER(c.class_name) = LOWER($3) THEN 0 ELSE 1 END,
    c.min_initial_investment ASC
LIMIT 10;
```

## Pending Action State Management

### Session State for Multi-Entity Resolution

```go
type ChatSession struct {
    SessionID     string
    Context       DSLGenerationRequest
    History       []ChatMessage
    PendingAction *PendingAction  // Current pending resolution
    
    // Resolution state for multi-entity operations
    ResolutionQueue []EntityResolution // Entities still to resolve
    ResolvedEntities map[string]string  // entityType -> entityID
}

type EntityResolution struct {
    EntityType EntityType
    SearchTerm string
    Required   bool
}
```

### Example Resolution Queue

```go
// User: "Subscribe Alpine Capital to Global Equity Fund Class A"

session.ResolutionQueue = []EntityResolution{
    {EntityType: "investor", SearchTerm: "Alpine Capital", Required: true},
    {EntityType: "fund", SearchTerm: "Global Equity Fund", Required: true},
    {EntityType: "share_class", SearchTerm: "Class A", Required: true},
}

// Process queue one by one:
// 1. Resolve investor → PendingAction = "select_investor"
// 2. User selects → ResolvedEntities["investor"] = "uuid-123"
// 3. Resolve fund → PendingAction = "select_fund"
// 4. User confirms → ResolvedEntities["fund"] = "uuid-456"
// 5. Resolve class → PendingAction = "select_share_class"
// 6. User selects → ResolvedEntities["share_class"] = "uuid-789"
// 7. All resolved → Execute DSL with all UUIDs
```

## Fund-Specific Fuzzy Matching

### Common Fund Name Variations

| User Input | Database | Match |
|------------|----------|-------|
| Global Equity | Global Equity Fund | ✅ 92% |
| Globa Equity | Global Equity Fund | ✅ 88% (typo) |
| GEF | Global Equity Fund | ❌ 30% (acronym - too low) |
| Alpha Fund | Alpha Hedge Fund LP | ✅ 85% |
| Alpha | Alpha Hedge Fund LP | ✅ 70% (ask user) |

### Class Name Matching

| User Input | Database | Match |
|------------|----------|-------|
| Class A | Class A | ✅ 100% (exact) |
| A | Class A | ✅ 70% (ask user) |
| Institutional | Class I (type: INSTITUTIONAL) | ✅ 90% |
| Retail Class | Class R (type: RETAIL) | ✅ 85% |

## Benefits of Fund Entity Resolution

### 1. Prevents Wrong Fund Selection
```
❌ Without: User types "Global Fund" → Picks first match → Wrong fund!
✅ With:    User types "Global Fund" → Shows 3 "Global" funds → User selects correct one
```

### 2. Handles Fund Name Variations
```
User: "I want to invest in the Euro equity fund"
System: Found:
  1. European Equity Fund (EUR)
  2. Euro Equity Strategy Fund (EUR)
  3. Global Equity Fund - EUR Share Class (EUR)
User selects #2
```

### 3. Class-Specific Validation
```
User: "Subscribe to Alpha Fund Class Z"
System: Found Alpha Fund, but no Class Z exists.
        Available classes: A, B, I
        Did you mean Class A? (yes/no)
```

### 4. Complete Audit Trail
```sql
-- See which fund user was presented and selected
SELECT 
    e.dsl_text,
    e.execution_status,
    e.created_at
FROM hf_dsl_executions e
WHERE e.dsl_text LIKE '%fund-id-here%'
ORDER BY e.created_at DESC;

-- Shows: User resolved "Global Equity Fund" to fund_id XYZ
```

## Edge Cases Handled

### 1. Fund + Class Ambiguity
```
User: "Subscribe to Fund A"
System: Found Fund A with 3 share classes:
        - Class A (Retail, EUR, Min: €1,000)
        - Class B (Retail, USD, Min: $1,000)
        - Class I (Institutional, EUR, Min: €1,000,000)
        Which class?
User: "Class A"
System: ✅ Resolved to Fund A / Class A
```

### 2. Multiple Funds, Same Name
```
User: "Global Fund"
System: Found 2 funds named "Global Fund":
        1. Global Fund (Domicile: LU, Inception: 2020)
        2. Global Fund (Domicile: IE, Inception: 2023)
        Which one?
```

### 3. Fund Not Found
```
User: "Subscribe to New Strategy Fund"
System: No fund found matching "New Strategy Fund"
        Create new fund? (yes/no)
User: "no"
System: ❌ Cancelled
```

## Summary

**Fund Entity Resolution** provides:

✅ **Same UX as investor resolution** - consistent user experience  
✅ **Fuzzy matching** - handles typos in fund names  
✅ **Multi-entity workflows** - resolves investor + fund + class sequentially  
✅ **Two-stage KYC** - general clearance without fund, then fund-specific  
✅ **Complete audit trail** - all entity selections logged  
✅ **Prevents errors** - confirms before using wrong fund/class  
✅ **UUID embedding** - DSL always has correct database IDs  

**Result**: DSL operations always reference the correct funds and classes with proper UUIDs! 🎯