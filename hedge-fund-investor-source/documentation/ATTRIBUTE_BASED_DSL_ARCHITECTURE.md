# Attribute-Based DSL Architecture with RAG Integration

## Overview

This document describes the **Attribute-Based DSL (Domain-Specific Language)** architecture for the Hedge Fund Investor Register system, which enables:

1. **Self-Describing DSL**: Variables reference dictionary attributes by UUID
2. **Complete Auditability**: Every data element has provenance and lineage
3. **RAG-Enabled AI Generation**: Rich metadata enables intelligent DSL generation
4. **Type-Safe Validation**: Parser validates attribute UUIDs and types
5. **Semantic Search**: Vector embeddings enable natural language queries

## Core Concept

### Traditional DSL (Without Attributes)
```lisp
(investor.start-opportunity
  :legal-name "Acme Capital Partners LP"
  :type "CORPORATE"
  :domicile "CH")
```

**Problems:**
- Hard-coded field names with no metadata
- No validation of field existence or type
- No audit trail for data lineage
- AI agent must "guess" valid field names

### Attribute-Based DSL (With Dictionary)
```lisp
(investor.start-opportunity
  @attr{a1b2c3d4-0001-0000-0000-000000000001} = "Acme Capital Partners LP"
  @attr{a1b2c3d4-0002-0000-0000-000000000002} = "CORPORATE"
  @attr{a1b2c3d4-0003-0000-0000-000000000003} = "CH")
```

**Benefits:**
- Each `@attr{uuid}` references a dictionary entry with full metadata
- Parser validates UUID exists and value matches expected type
- Complete audit trail: who created, when, where stored
- AI agent retrieves valid attributes via RAG

## Data Dictionary Schema

### Core Table Structure

```sql
CREATE TABLE "dsl-ob-poc".dictionary (
    attribute_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,          -- Human-readable name
    long_description TEXT,                       -- Rich description for RAG
    group_id VARCHAR(100) NOT NULL,             -- Logical grouping
    mask VARCHAR(50) DEFAULT 'string',          -- Data type
    domain VARCHAR(100),                        -- Domain ownership
    vector TEXT,                                -- Semantic keywords for RAG
    source JSONB,                               -- Where data comes from
    sink JSONB,                                 -- Where data is stored
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

### Example Dictionary Entry

```sql
INSERT INTO "dsl-ob-poc".dictionary VALUES (
  'a1b2c3d4-0001-0000-0000-000000000001',
  'hf.investor.legal-name',
  'Official legal name of the investor entity or individual as it appears on 
   incorporation documents, passports, or other legal identification. This is 
   the name used for all legal agreements, subscription documents, tax forms, 
   and regulatory reporting. For individuals: full legal name. For entities: 
   exact name as registered. Critical for KYC/AML compliance and legal documentation.',
  'hf-investor-identity',
  'string',
  'hedge-fund-investor',
  'legal name official name entity name individual name full name investor name 
   legal entity registered name incorporation name kyc compliance',
  '{"system": "hedge-fund-investor", "collector": "kyc-process", "required": true}',
  '{"table": "hf_investors", "column": "legal_name", "max_length": 500}'
);
```

## Attribute Groups for Hedge Fund Investor Domain

### Identity Attributes (hf-investor-identity)
- `hf.investor.investor-id` - UUID primary key
- `hf.investor.investor-code` - Human-readable code (INV-2024-001)
- `hf.investor.legal-name` - Official legal name
- `hf.investor.short-name` - Display name
- `hf.investor.type` - INDIVIDUAL, CORPORATE, TRUST, etc.
- `hf.investor.domicile` - Country code (ISO-3166)
- `hf.investor.lei` - Legal Entity Identifier
- `hf.investor.registration-number` - Company registration
- `hf.investor.source` - Lead source

### Address Attributes (hf-investor-address)
- `hf.investor.address-line1` - Primary address
- `hf.investor.address-line2` - Secondary address
- `hf.investor.city` - City/municipality
- `hf.investor.state-province` - State/province/region
- `hf.investor.postal-code` - ZIP/postal code
- `hf.investor.country` - Country code

### Contact Attributes (hf-investor-contact)
- `hf.investor.primary-contact-name` - Contact person name
- `hf.investor.primary-contact-email` - Email address
- `hf.investor.primary-contact-phone` - Phone number

### Lifecycle Attributes (hf-investor-lifecycle)
- `hf.investor.status` - Current state (OPPORTUNITY → OFFBOARDED)

### KYC Attributes (hf-kyc-profile)
- `hf.kyc.risk-rating` - LOW, MEDIUM, HIGH, PROHIBITED
- `hf.kyc.tier` - SIMPLIFIED, STANDARD, ENHANCED
- `hf.kyc.screening-provider` - worldcheck, refinitiv, etc.
- `hf.kyc.screening-result` - CLEAR, POTENTIAL_MATCH, TRUE_POSITIVE
- `hf.kyc.approved-by` - Approver name and title
- `hf.kyc.refresh-frequency` - MONTHLY, QUARTERLY, ANNUAL, etc.
- `hf.kyc.refresh-due-at` - Next refresh date

### Tax Attributes (hf-tax-profile)
- `hf.tax.fatca-status` - US_PERSON, NON_US_PERSON, etc.
- `hf.tax.crs-classification` - INDIVIDUAL, ENTITY, etc.
- `hf.tax.form-type` - W9, W8_BEN, W8_BEN_E, etc.
- `hf.tax.withholding-rate` - Tax withholding percentage
- `hf.tax.tin-type` - SSN, EIN, ITIN, FOREIGN_TIN, GIIN
- `hf.tax.tin-value` - Actual tax ID (encrypted)

### Fund Structure Attributes (hf-fund-structure)
- `hf.fund.fund-id` - Fund UUID
- `hf.fund.fund-name` - Fund name
- `hf.fund.class-id` - Share class UUID
- `hf.fund.class-name` - Share class name (A, I, S, etc.)
- `hf.fund.series-id` - Series UUID (for equalization)

### Trading Attributes (hf-trading)
- `hf.trade.trade-id` - Trade UUID
- `hf.trade.trade-type` - SUBSCRIPTION, REDEMPTION
- `hf.trade.amount` - Trade amount
- `hf.trade.currency` - Settlement currency
- `hf.trade.trade-date` - Trade date
- `hf.trade.value-date` - Settlement date
- `hf.trade.nav-per-share` - NAV at execution
- `hf.trade.units` - Units allocated/redeemed

### Banking Attributes (hf-banking)
- `hf.bank.instruction-id` - Bank instruction UUID
- `hf.bank.currency` - Account currency
- `hf.bank.bank-name` - Bank name
- `hf.bank.account-name` - Account holder name
- `hf.bank.iban` - IBAN
- `hf.bank.swift` - SWIFT/BIC code
- `hf.bank.account-number` - Account number

## RAG (Retrieval Augmented Generation) Integration

### How RAG Works with Dictionary Attributes

1. **Semantic Search**: AI agent converts user instruction to embedding
2. **Vector Retrieval**: Search `vector` field for relevant attributes
3. **Context Enrichment**: Load `long_description` for matched attributes
4. **DSL Generation**: Generate DSL using retrieved attribute UUIDs
5. **Validation**: Parser validates UUIDs and types

### RAG System Prompt Enhancement

```
You are a hedge fund investor DSL generation agent with access to a comprehensive
attribute dictionary. When generating DSL:

1. SEARCH the attribute dictionary using semantic similarity
2. USE attribute UUIDs in DSL (@attr{uuid} syntax)
3. VALIDATE attribute types match values
4. PROVIDE rich context from long_description fields

Example attribute retrieval:
User: "Create opportunity for Swiss investor"
RAG Search: "investor swiss domicile country switzerland"
Retrieved Attributes:
  - a1b2c3d4-0003-0000-0000-000000000003 (hf.investor.domicile)
    Description: "Country of domicile or tax residence..."
    Type: country-code
    Valid values: ISO-3166-1-alpha-2

Generated DSL:
(investor.start-opportunity
  @attr{a1b2c3d4-0001-0000-0000-000000000001} = "Acme Capital LP"
  @attr{a1b2c3d4-0003-0000-0000-000000000003} = "CH")
```

### Vector Field Content Strategy

The `vector` field contains **semantic keywords** optimized for RAG retrieval:

```sql
vector = 'legal name official name entity name investor name 
          kyc compliance aml documentation regulatory reporting 
          incorporation certificate passport identification'
```

**Keywords Include:**
- **Synonyms**: legal name, official name, entity name
- **Use Cases**: kyc compliance, regulatory reporting
- **Related Concepts**: incorporation, passport, identification
- **Domain Terms**: investor, entity, compliance, documentation

### RAG Query Examples

**User Query**: "I need the investor's tax ID number"
**Vector Search**: "tax id number tin ssn ein identification"
**Retrieved Attribute**: `hf.tax.tin-value`
**Context**: "Actual Tax Identification Number. Format varies by type..."

**User Query**: "What's their risk rating?"
**Vector Search**: "risk rating kyc aml compliance assessment"
**Retrieved Attribute**: `hf.kyc.risk-rating`
**Context**: "AML/CFT risk rating assigned after KYC assessment..."

**User Query**: "Bank details for USD"
**Vector Search**: "bank banking usd currency account swift"
**Retrieved Attributes**: 
- `hf.bank.currency`
- `hf.bank.bank-name`
- `hf.bank.swift`

## DSL Parser with Attribute Validation

### Parser Architecture

```go
type AttributeBasedDSLParser struct {
    dictionary map[uuid.UUID]*DictionaryAttribute
}

func (p *AttributeBasedDSLParser) Parse(dsl string) (*ParsedDSL, error) {
    // 1. Extract @attr{uuid} references
    attrRefs := extractAttributeReferences(dsl)
    
    // 2. Validate each UUID exists in dictionary
    for _, ref := range attrRefs {
        attr, exists := p.dictionary[ref.UUID]
        if !exists {
            return nil, fmt.Errorf("unknown attribute: %s", ref.UUID)
        }
        
        // 3. Validate value matches attribute type
        if err := validateValue(attr, ref.Value); err != nil {
            return nil, fmt.Errorf("invalid value for %s: %w", attr.Name, err)
        }
    }
    
    // 4. Build parsed structure
    return &ParsedDSL{
        Verb: extractVerb(dsl),
        Attributes: attrRefs,
        Validated: true,
    }, nil
}
```

### Type Validation

```go
func validateValue(attr *DictionaryAttribute, value interface{}) error {
    switch attr.Mask {
    case "uuid":
        _, err := uuid.Parse(value.(string))
        return err
    case "country-code":
        return validateISO3166(value.(string))
    case "email":
        return validateEmail(value.(string))
    case "enum":
        return validateEnum(attr.Source, value.(string))
    case "decimal":
        return validateDecimal(value)
    case "date":
        return validateDateFormat(value.(string))
    case "ssn":
        return validateSSN(value.(string)) // Encrypted
    default:
        return validateString(value.(string), attr.Sink)
    }
}
```

## Complete DSL Lifecycle Example

### Step 1: User Instruction (Natural Language)
```
User: "Create an opportunity for Acme Capital Partners LP, 
       a corporate investor from Switzerland"
```

### Step 2: RAG Attribute Retrieval

**AI Agent Searches Dictionary:**
```sql
SELECT attribute_id, name, long_description, mask, source, sink
FROM "dsl-ob-poc".dictionary
WHERE domain = 'hedge-fund-investor'
  AND group_id = 'hf-investor-identity'
  AND vector @@ to_tsquery('investor legal name corporate switzerland domicile');
```

**Retrieved Attributes:**
1. `a1b2c3d4-0001-0000-0000-000000000001` - hf.investor.legal-name (string)
2. `a1b2c3d4-0002-0000-0000-000000000002` - hf.investor.type (enum: CORPORATE)
3. `a1b2c3d4-0003-0000-0000-000000000003` - hf.investor.domicile (country-code: CH)

### Step 3: AI-Generated DSL with Attributes

```lisp
(investor.start-opportunity
  @attr{a1b2c3d4-0001-0000-0000-000000000001} = "Acme Capital Partners LP"
  @attr{a1b2c3d4-0002-0000-0000-000000000002} = "CORPORATE"
  @attr{a1b2c3d4-0003-0000-0000-000000000003} = "CH")
```

### Step 4: Parser Validation

```go
parser.Validate(dsl) {
    // Validate attr{uuid-0001} exists
    attr1 := dictionary.Get("a1b2c3d4-0001-0000-0000-000000000001")
    // Check: attr1.mask == "string" ✓
    // Check: value "Acme Capital..." is valid string ✓
    
    // Validate attr{uuid-0002} exists
    attr2 := dictionary.Get("a1b2c3d4-0002-0000-0000-000000000002")
    // Check: attr2.mask == "enum" ✓
    // Check: "CORPORATE" in attr2.source.values ✓
    
    // Validate attr{uuid-0003} exists
    attr3 := dictionary.Get("a1b2c3d4-0003-0000-0000-000000000003")
    // Check: attr3.mask == "country-code" ✓
    // Check: "CH" is valid ISO-3166 code ✓
}
```

### Step 5: DSL Execution

```go
executor.Execute(parsedDSL) {
    // Resolve attribute names
    params := map[string]interface{}{
        "legal_name": parsedDSL.Attributes[uuid-0001].Value,
        "type":       parsedDSL.Attributes[uuid-0002].Value,
        "domicile":   parsedDSL.Attributes[uuid-0003].Value,
    }
    
    // Execute database insert
    investorID := createInvestor(params)
    
    // Store attribute values
    for attrUUID, value := range parsedDSL.Attributes {
        storeAttributeValue(investorID, attrUUID, value)
    }
}
```

### Step 6: Attribute Value Storage

```sql
INSERT INTO "dsl-ob-poc".attribute_values (
    cbu_id,
    attribute_id,
    value,
    state,
    source
) VALUES (
    'investor-uuid-here',
    'a1b2c3d4-0001-0000-0000-000000000001',
    '{"value": "Acme Capital Partners LP"}',
    'resolved',
    '{"dsl_operation": "investor.start-opportunity", "timestamp": "2024-01-15T10:00:00Z"}'
);
```

## Audit Trail and Lineage

### Complete Data Provenance

Every attribute value has:

1. **Source**: Where it came from
   - DSL operation that created it
   - User who executed the operation
   - Timestamp of creation

2. **Sink**: Where it's stored
   - Database table and column
   - Constraints and validations
   - Encryption requirements (for sensitive data)

3. **History**: Full audit trail
   - All DSL operations that touched this attribute
   - State transitions
   - Validation results

### Example Audit Query

```sql
-- Get complete lineage for investor legal name
SELECT 
    d.name as attribute_name,
    d.long_description,
    av.value,
    av.state,
    av.source->>'dsl_operation' as operation,
    av.observed_at,
    de.dsl_text,
    de.triggered_by
FROM "dsl-ob-poc".attribute_values av
JOIN "dsl-ob-poc".dictionary d ON d.attribute_id = av.attribute_id
JOIN "hf-investor".hf_dsl_executions de ON de.investor_id::text = av.cbu_id
WHERE d.name = 'hf.investor.legal-name'
  AND av.cbu_id = 'investor-uuid'
ORDER BY av.observed_at DESC;
```

**Result:**
```
attribute_name          | hf.investor.legal-name
long_description        | Official legal name of the investor...
value                   | {"value": "Acme Capital Partners LP"}
state                   | resolved
operation               | investor.start-opportunity
observed_at             | 2024-01-15 10:00:00+00
dsl_text                | (investor.start-opportunity @attr{...}...)
triggered_by            | operations@fundadmin.com
```

## Benefits of Attribute-Based DSL

### 1. Self-Describing System
- Every field has rich metadata
- AI can understand context and constraints
- Humans can query "what does this field mean?"

### 2. Type Safety
- Parser validates types at parse time
- No runtime type errors
- Enum validation against allowed values

### 3. Complete Auditability
- Every data point has provenance
- Track: who, what, when, where, why
- Regulatory compliance built-in

### 4. Intelligent AI Generation
- RAG retrieves only relevant attributes
- Rich descriptions improve AI accuracy
- Constraints prevent invalid generation

### 5. Evolutionary Schema
- Add new attributes without code changes
- AI automatically discovers new attributes
- Backward compatibility maintained

### 6. Data Governance
- Centralized data dictionary
- Ownership and stewardship clear
- Privacy and sensitivity flagged

## RAG Prompt Template for DSL Generation

```
You are an expert hedge fund investor onboarding agent with access to a 
comprehensive attribute dictionary.

TASK: Generate DSL for the following instruction:
"{user_instruction}"

CONTEXT:
- Current investor state: {current_state}
- Available attributes retrieved via semantic search:

{for each retrieved attribute:
  UUID: {attribute_id}
  Name: {name}
  Description: {long_description}
  Type: {mask}
  Constraints: {source}
  Valid Values: {if enum, list values}
}

RULES:
1. Use @attr{uuid} syntax for all data attributes
2. Validate value types match attribute mask
3. Only use attributes from the retrieved list above
4. Generate valid S-expression DSL format
5. Include state transition information

OUTPUT FORMAT (JSON):
{
  "dsl": "(verb @attr{uuid1} = \"value1\" @attr{uuid2} = \"value2\")",
  "verb": "investor.start-opportunity",
  "attributes": [
    {"uuid": "uuid1", "name": "hf.investor.legal-name", "value": "Acme Capital LP"},
    {"uuid": "uuid2", "name": "hf.investor.type", "value": "CORPORATE"}
  ],
  "from_state": "OPPORTUNITY",
  "to_state": "PRECHECKS",
  "explanation": "Creates initial investor opportunity with legal name and type",
  "confidence": 0.95
}
```

## Implementation Roadmap

### Phase 1: Dictionary Population (Week 1)
- [ ] Run `data_dictionary_hedge_fund_investor.sql`
- [ ] Populate 50+ hedge fund investor attributes
- [ ] Validate UUIDs generated correctly
- [ ] Test attribute retrieval queries

### Phase 2: Parser Enhancement (Week 2)
- [ ] Implement `@attr{uuid}` syntax parser
- [ ] Add UUID validation against dictionary
- [ ] Implement type validation for each mask type
- [ ] Create comprehensive test suite

### Phase 3: RAG Integration (Week 3)
- [ ] Implement semantic search on vector field
- [ ] Enhance AI agent prompt with attribute context
- [ ] Test RAG retrieval accuracy
- [ ] Fine-tune vector field content

### Phase 4: DSL Execution (Week 4)
- [ ] Update execution engine to resolve @attr{uuid}
- [ ] Store attribute values in attribute_values table
- [ ] Implement audit trail queries
- [ ] Create lineage visualization

### Phase 5: Testing & Validation (Week 5)
- [ ] End-to-end lifecycle test with attributes
- [ ] Validate RAG accuracy across all 17 verbs
- [ ] Performance testing (parse, validate, execute)
- [ ] Documentation and examples

## Conclusion

The **Attribute-Based DSL Architecture** transforms the hedge fund investor register into a self-describing, auditable, AI-enhanced system where:

- **Every data element** has rich metadata
- **AI agents** intelligently generate DSL using RAG
- **Parsers** validate correctness deterministically
- **Audit trails** provide complete data lineage
- **Evolution** happens without breaking changes

This architecture enables sophisticated AI-powered onboarding while maintaining enterprise-grade governance, compliance, and auditability.

---

**Version**: 1.0.0  
**Last Updated**: December 2024  
**Related Documents**: 
- `data_dictionary_hedge_fund_investor.sql` - Attribute definitions
- `hf_dsl_agent.go` - RAG-powered DSL generator
- `HEDGE_FUND_INVESTOR.md` - System overview