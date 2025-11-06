# Dictionary Attributes - AI RAG Optimized

## Overview

This directory contains comprehensive attribute definitions for the `"dsl-ob-poc".dictionary` table, optimized for **Retrieval-Augmented Generation (RAG)** by AI agents.

## Architecture Principle

**AttributeID-as-Type Pattern**: Each attribute's UUID serves as its semantic type identifier in the DSL. The dictionary provides rich metadata that AI agents use to:

1. **Discover relevant attributes** based on natural language descriptions
2. **Generate valid DSL** using correct attribute IDs
3. **Understand context** for intelligent workflow orchestration
4. **Validate data** using mask, constraints, and domain information

## Seed File: `seed_dictionary_attributes.sql`

### Contents

| Domain | Attribute Count | Coverage |
|--------|----------------|----------|
| **Onboarding Lifecycle** | 11 | Opportunity → Go-Live → Offboarding |
| **KYC - Institutional** | 20 | Corporate entities, trusts, partnerships |
| **KYC - Proper Person/Retail** | 28 | Individual investors |
| **Common KYC Workflow** | 10 | Status, approval, review tracking |
| **TOTAL** | **69 attributes** | Complete onboarding + KYC coverage |

### Domains Covered

#### 1. Onboarding Lifecycle Attributes

**Purpose**: Track complete client onboarding journey from opportunity to active servicing

**Key Attributes**:
- `opportunity.cbu_id` - Master client identifier
- `opportunity.nature_purpose` - AI-parseable business description
- `products.selected` - Product codes driving workflow
- `services.discovered` - Catalog-retrieved services
- `resources.planned` - Infrastructure to provision
- `onboarding.current_state` - State machine tracking
- `onboarding.version` - DSL version for event sourcing

**AI Usage**: Agent parses `nature_purpose` to infer products, jurisdictions, entity types

#### 2. KYC - Institutional Attributes

**Purpose**: Identity verification and compliance for corporate entities

**Entity Types Covered**:
- Corporations (AG, SA, Ltd)
- Limited Liability Companies (LLC)
- Partnerships (General, Limited)
- Trusts (Discretionary, Fixed Interest, Unit)
- Foundations

**Key Attributes**:
- Entity identity: `legal_name`, `entity_type`, `jurisdiction`, `registration_number`
- Documents: `certificate_of_incorporation`, `articles_memorandum`, `shareholder_register`
- Risk: `risk_rating`, `pep_exposure`, `sanctions_screening_result`
- Trust-specific: `trust_type`, `settlor_identity`, `trustee_identity`, `beneficiary_class`

**AI Usage**: Agent determines UBO workflow based on entity type, extracts data from documents

#### 3. KYC - Proper Person / Retail Attributes

**Purpose**: Identity verification and suitability for individual investors

**Key Attributes**:
- Identity: `full_legal_name`, `date_of_birth`, `nationality`, `country_of_residence`
- Documents: `passport_number`, `national_id_number`, `tax_identification_number`
- Financial: `employment_status`, `annual_income`, `net_worth`, `source_of_wealth`
- Risk: `pep_status`, `sanctions_screening_result`, `risk_rating`
- Suitability: `investment_experience`, `risk_tolerance`, `investment_objectives`

**AI Usage**: Agent assesses product suitability, determines required documents, calculates risk rating

#### 4. Common KYC Workflow Attributes

**Purpose**: Track KYC status and lifecycle across both institutional and retail

**Key Attributes**:
- `kyc.status` - Workflow state (NOT_STARTED → APPROVED)
- `kyc.tier` - SIMPLIFIED, STANDARD, or ENHANCED due diligence
- `kyc.approval_date` - Timestamp for expiry calculation
- `kyc.next_review_date` - Periodic review scheduling
- `kyc.edd_required` - Enhanced Due Diligence flag

**AI Usage**: Agent routes workflows, determines document requirements, triggers reviews

## Rich Descriptions for AI

Each attribute includes verbose `long_description` fields (100-400 words) containing:

### What AI Agents Learn:

1. **Semantic Meaning**: What the attribute represents in business terms
2. **Workflow Context**: When in the process it's collected
3. **Downstream Usage**: How it's used in subsequent operations
4. **Regulatory Context**: Compliance implications (FATF, FATCA, CRS, MiFID II)
5. **Validation Rules**: Format requirements, enumerations, constraints
6. **Risk Implications**: How it affects risk ratings and due diligence
7. **Related Attributes**: Dependencies and relationships
8. **Examples**: Concrete values for pattern recognition

### Example - AI-Optimized Description

```sql
'kyc.institutional.legal_name'
→ "Official registered legal name of the institutional entity exactly as it
   appears on incorporation or formation documents. Used for identity
   verification, sanctions screening, and legal contract generation. Must
   match name on Certificate of Incorporation, Trust Deed, or Partnership
   Agreement. Critical for regulatory reporting and audit trail."
```

**AI learns**:
- **Source documents** to extract from (Certificate of Incorporation, Trust Deed)
- **Validation requirement** (must match official documents exactly)
- **Downstream uses** (sanctions screening, contracts, reporting)
- **Criticality** (required for audit trail)

## Vector Database Optimization

The `vector` column (currently text) is designed for future semantic embeddings:

```sql
CREATE TABLE "dsl-ob-poc".dictionary (
    ...
    vector TEXT,  -- Will store embedding vectors for semantic search
    ...
);
```

**Future Enhancement**: Generate embeddings from `long_description` using:
- OpenAI text-embedding-3-large
- Google text-embedding-gecko
- Anthropic voyage embeddings

**Use Case**: AI agent query like "What attributes track someone's wealth?" returns:
- `kyc.individual.net_worth`
- `kyc.individual.source_of_wealth`
- `kyc.individual.source_of_funds`
- `kyc.individual.annual_income`

## Source and Sink Metadata

Each attribute has structured JSONB metadata describing data provenance:

### Source Metadata
```json
{
  "primary": "PASSPORT_DOCUMENT",
  "secondary": "NATIONAL_ID_CARD",
  "tertiary": "BIRTH_CERTIFICATE"
}
```

**AI Usage**: Agent knows where to retrieve or request the value

### Sink Metadata
```json
{
  "primary": "INDIVIDUAL_REGISTRY",
  "secondary": "SANCTIONS_SCREENING"
}
```

**AI Usage**: Agent knows where to store/sync the value downstream

## Execution

### Prerequisites
```bash
export DB_CONN_STRING="postgres://user:password@localhost:5432/postgres?sslmode=disable"
```

### Run Seed
```bash
psql $DB_CONN_STRING -f sql/seed_dictionary_attributes.sql
```

### Verify
```sql
-- Count attributes by domain
SELECT domain, COUNT(*)
FROM "dsl-ob-poc".dictionary
GROUP BY domain
ORDER BY COUNT(*) DESC;

-- Sample onboarding attributes
SELECT name, LEFT(long_description, 100) || '...' as description
FROM "dsl-ob-poc".dictionary
WHERE domain = 'ONBOARDING'
ORDER BY name;

-- Sample KYC attributes
SELECT name, group_id, mask
FROM "dsl-ob-poc".dictionary
WHERE domain = 'KYC' AND group_id LIKE '%individual%'
ORDER BY group_id, name;
```

## Integration with AI Agent

### discover-kyc Command

**Before**: AI generates generic KYC requirements

**After**: AI retrieves specific attributes from dictionary
```
User: "Start KYC for UCITS fund"
Agent queries dictionary:
  - WHERE domain='KYC' AND group_id LIKE '%institutional%'
  - Finds: legal_name, entity_type, jurisdiction, etc.
Agent generates DSL:
  (kyc.start
    (collect (attr-id "uuid-1") ; kyc.institutional.legal_name
             (attr-id "uuid-2") ; kyc.institutional.entity_type
             (attr-id "uuid-3")) ; kyc.institutional.jurisdiction
  )
```

### discover-resources Command

**Before**: Resources have undefined variables

**After**: Resources reference dictionary attributes by group_id
```sql
SELECT * FROM prod_resources WHERE dictionary_group = 'custody_account';
-- Returns: CustodyAccount resource

SELECT * FROM dictionary WHERE group_id = 'custody_account';
-- Returns: account_number, account_type, custody_bank, etc.

Agent generates DSL:
  (resources.plan
    (resource.create "CustodyAccount"
      (var (attr-id "uuid-1"))  ; custody.account_number
      (var (attr-id "uuid-2"))  ; custody.account_type
    )
  )
```

## Maintenance

### Adding New Attributes

1. Identify domain (ONBOARDING, KYC, UBO, SETTLEMENT, etc.)
2. Define group_id for logical clustering
3. Write verbose long_description (AI context)
4. Specify mask (STRING, ENUM, BOOLEAN, DATE, DECIMAL, etc.)
5. Define source and sink metadata
6. Add to seed file with `gen_random_uuid()`

### Updating Descriptions

**DO NOT** modify existing attribute_id UUIDs (breaks DSL references)

**DO** enhance long_description to improve AI understanding

```sql
UPDATE "dsl-ob-poc".dictionary
SET long_description = '<enhanced description>',
    updated_at = NOW()
WHERE name = 'kyc.institutional.legal_name';
```

## Future Enhancements

1. **Vector Embeddings**: Generate semantic embeddings for similarity search
2. **Multi-language Support**: Descriptions in EN, FR, DE for global operations
3. **Validation Rules**: Expand constraints to include regex, ranges, dependencies
4. **Derivation Rules**: Define calculated attributes (e.g., risk_score from multiple inputs)
5. **Deprecation Workflow**: Mark obsolete attributes without breaking existing DSL
6. **Version History**: Track attribute definition changes over time

## Related Documentation

- **CLAUDE.md** - Core architectural patterns (DSL-as-State, AttributeID-as-Type)
- **sql/init.sql** - Database schema definition
- **internal/dictionary/** - Go models and repository
- **internal/cli/discover_resources.go** - Attribute usage in resource planning
