# Database Schema Documentation

## Overview

This document describes the complete database schema for the DSL Onboarding POC, including five new tables for managing products, services, resources, attributes, and dictionaries.

## Schema Structure

```
dsl-ob-poc (schema)
├── Core Tables
│   └── dsl_ob (immutable versioned DSL records)
├── Product & Service Configuration
│   ├── products
│   └── services
├── Resource Management
│   └── prod_resources
├── Attributes & Dictionaries
│   ├── attributes
│   └── dictionaries
├── Relationship Tables (Many-to-Many)
│   ├── service_products
│   ├── product_attributes
│   └── dictionary_attributes
```

## Core Tables

### dsl_ob (Domain-Specific Language Onboarding)

Stores immutable, versioned DSL records for tracking state changes over time.

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| version_id | UUID | PRIMARY KEY | Unique identifier for each DSL version |
| cbu_id | VARCHAR(255) | NOT NULL | Customer Business Unit identifier |
| dsl_text | TEXT | NOT NULL | Full S-expression DSL content |
| created_at | TIMESTAMPTZ | DEFAULT now() | Timestamp when version was created |

**Indexes:**
- `idx_dsl_ob_cbu_id_created_at` - Composite index on (cbu_id, created_at DESC) for fast lookups

**Purpose:** Event sourcing pattern - every state change creates a new immutable version

---

## Product & Service Configuration Tables

### products

Master catalog of products offered by the system.

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| product_id | UUID | PRIMARY KEY | Unique product identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Product name (must be unique) |
| description | TEXT | NOT NULL | Product description |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT now() | Last update timestamp |
| is_active | BOOLEAN | DEFAULT true | Soft delete flag |

**Constraints:**
- `products_name_not_empty` - Name cannot be empty or whitespace-only

**Indexes:**
- `idx_products_name` - On product name for lookups
- `idx_products_active` - On is_active for filtering

**Example:**
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "CUSTODY",
  "description": "Custody and safekeeping of client assets",
  "is_active": true
}
```

### services

Services that bundle multiple products together.

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| service_id | UUID | PRIMARY KEY | Unique service identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Service name (must be unique) |
| description | TEXT | NOT NULL | Service description |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT now() | Last update timestamp |
| is_active | BOOLEAN | DEFAULT true | Soft delete flag |

**Constraints:**
- `services_name_not_empty` - Name cannot be empty or whitespace-only

**Indexes:**
- `idx_services_name` - On service name for lookups
- `idx_services_active` - On is_active for filtering

**Example:**
```json
{
  "service_id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "FULL_INVESTMENT_MANAGEMENT",
  "description": "Complete investment portfolio management including custody, accounting, and reporting",
  "is_active": true
}
```

---

## Resource Management Tables

### prod_resources

Resources (infrastructure, tools, systems) required to deliver products and services.

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| resource_id | UUID | PRIMARY KEY | Unique resource identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Resource name |
| description | TEXT | NOT NULL | Resource description |
| owner | VARCHAR(255) | NOT NULL | Owner/responsible team |
| dictionary_id | UUID | FOREIGN KEY | Reference to dictionaries table (optional) |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT now() | Last update timestamp |
| is_active | BOOLEAN | DEFAULT true | Soft delete flag |

**Constraints:**
- `prod_resources_name_not_empty` - Name cannot be empty
- `prod_resources_owner_not_empty` - Owner cannot be empty
- `fk_prod_resources_dictionary` - Foreign key to dictionaries(dictionary_id) ON DELETE SET NULL

**Indexes:**
- `idx_prod_resources_name` - On resource name
- `idx_prod_resources_owner` - On owner for team lookups
- `idx_prod_resources_active` - On is_active

**Example:**
```json
{
  "resource_id": "770e8400-e29b-41d4-a716-446655440002",
  "name": "CUSTODY_DATABASE",
  "description": "Primary database for custody operations",
  "owner": "Infrastructure Team",
  "dictionary_id": "880e8400-e29b-41d4-a716-446655440003",
  "is_active": true
}
```

---

## Attributes & Dictionaries Tables

### attributes

Individual data attributes with comprehensive metadata for data governance and lineage tracking.

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| attribute_id | UUID | PRIMARY KEY | Unique attribute identifier |
| name | VARCHAR(255) | NOT NULL | Attribute name |
| detailed_description | TEXT | NOT NULL | Comprehensive description |
| is_private | BOOLEAN | NOT NULL DEFAULT false | Whether this is private/sensitive data |
| private_type | VARCHAR(50) | CHECK: 'derived'\|'given'\|NULL | Classification if private (derived=computed, given=external) |
| primary_source_url | TEXT | - | Primary system where attribute originates |
| secondary_source_url | TEXT | - | Secondary/fallback source |
| tertiary_source_url | TEXT | - | Tertiary/fallback source |
| data_mask_type | VARCHAR(50) | NOT NULL DEFAULT 'string' | Data type/structure |
| primary_sink_url | TEXT | NOT NULL | Target system where data is persisted |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT now() | Last update timestamp |
| is_active | BOOLEAN | DEFAULT true | Soft delete flag |

**Constraints:**
- `attributes_name_not_empty` - Name cannot be empty
- `attributes_private_type_valid` - private_type must be 'derived', 'given', or NULL
- `attributes_private_type_logic` - If private_type is set, is_private must be true
- `attributes_sink_url_not_empty` - primary_sink_url cannot be empty

**Indexes:**
- `idx_attributes_name` - On attribute name
- `idx_attributes_is_private` - On is_private flag for privacy filtering
- `idx_attributes_private_type` - On private_type for classification
- `idx_attributes_active` - On is_active

**Data Types (data_mask_type):**
- `string` (default) - Text data
- `integer` - Numeric data
- `decimal` - Decimal/float data
- `boolean` - True/false data
- `json` - JSON structured data
- `date` - Date only
- `timestamp` - Date and time
- `mask` - Masked/hashed data
- `struct` - Complex structured data

**Example 1: Public Attribute**
```json
{
  "attribute_id": "990e8400-e29b-41d4-a716-446655440004",
  "name": "client_account_number",
  "detailed_description": "Unique identifier for client trading account",
  "is_private": false,
  "private_type": null,
  "primary_source_url": "https://api.custody-system.com/accounts",
  "secondary_source_url": null,
  "tertiary_source_url": null,
  "data_mask_type": "string",
  "primary_sink_url": "postgresql://vault.db/accounts/account_numbers",
  "is_active": true
}
```

**Example 2: Private Derived Attribute**
```json
{
  "attribute_id": "aa0e8400-e29b-41d4-a716-446655440005",
  "name": "account_risk_score",
  "detailed_description": "Computed risk score based on trading patterns and positions",
  "is_private": true,
  "private_type": "derived",
  "primary_source_url": "https://api.risk-engine.internal/calculate",
  "secondary_source_url": "https://cache.risk-engine.internal/scores",
  "tertiary_source_url": null,
  "data_mask_type": "decimal",
  "primary_sink_url": "postgresql://vault.db/risk/scores",
  "is_active": true
}
```

**Example 3: Private Given Attribute**
```json
{
  "attribute_id": "bb0e8400-e29b-41d4-a716-446655440006",
  "name": "client_tax_id",
  "detailed_description": "Client's tax identification number (PII)",
  "is_private": true,
  "private_type": "given",
  "primary_source_url": "https://api.kyc-system.com/client-data",
  "secondary_source_url": "https://backup-kyc.external-provider.com",
  "tertiary_source_url": null,
  "data_mask_type": "string",
  "primary_sink_url": "postgresql://vault.db/kyc/tax_identifiers",
  "is_active": true
}
```

### dictionaries

Collections of related attributes for classification and metadata management.

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| dictionary_id | UUID | PRIMARY KEY | Unique dictionary identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Dictionary name |
| description | TEXT | NOT NULL | Dictionary purpose and scope |
| attribute_id | UUID | FOREIGN KEY | Primary/reference attribute (optional) |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT now() | Last update timestamp |
| is_active | BOOLEAN | DEFAULT true | Soft delete flag |

**Constraints:**
- `dictionaries_name_not_empty` - Name cannot be empty
- `fk_dictionaries_attribute` - Foreign key to attributes(attribute_id) ON DELETE SET NULL

**Indexes:**
- `idx_dictionaries_name` - On dictionary name
- `idx_dictionaries_active` - On is_active

**Example:**
```json
{
  "dictionary_id": "cc0e8400-e29b-41d4-a716-446655440007",
  "name": "CLIENT_KYC_DATA",
  "description": "Know Your Client (KYC) data including identification, tax, and regulatory information",
  "attribute_id": "bb0e8400-e29b-41d4-a716-446655440006",
  "is_active": true
}
```

---

## Relationship Tables (Many-to-Many)

### service_products

Maps services to their component products (many-to-many relationship).

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| service_product_id | UUID | PRIMARY KEY | Unique mapping identifier |
| service_id | UUID | NOT NULL, FK | Reference to services table |
| product_id | UUID | NOT NULL, FK | Reference to products table |
| sequence_order | INT | NOT NULL DEFAULT 0 | Order within service composition |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |

**Constraints:**
- `fk_service_products_service` - Foreign key to services(service_id) ON DELETE CASCADE
- `fk_service_products_product` - Foreign key to products(product_id) ON DELETE CASCADE
- `unique_service_product` - Unique constraint on (service_id, product_id)

**Indexes:**
- `idx_service_products_service` - On service_id
- `idx_service_products_product` - On product_id

**Example:**
```json
{
  "service_product_id": "dd0e8400-e29b-41d4-a716-446655440008",
  "service_id": "660e8400-e29b-41d4-a716-446655440001",
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "sequence_order": 0
}
```

### dictionary_attributes

Maps dictionaries to their attributes (many-to-many relationship).

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| dictionary_attribute_id | UUID | PRIMARY KEY | Unique mapping identifier |
| dictionary_id | UUID | NOT NULL, FK | Reference to dictionaries table |
| attribute_id | UUID | NOT NULL, FK | Reference to attributes table |
| sequence_order | INT | NOT NULL DEFAULT 0 | Order within dictionary |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |

**Constraints:**
- `fk_dictionary_attributes_dictionary` - Foreign key to dictionaries(dictionary_id) ON DELETE CASCADE
- `fk_dictionary_attributes_attribute` - Foreign key to attributes(attribute_id) ON DELETE CASCADE
- `unique_dictionary_attribute` - Unique constraint on (dictionary_id, attribute_id)

**Indexes:**
- `idx_dictionary_attributes_dictionary` - On dictionary_id
- `idx_dictionary_attributes_attribute` - On attribute_id

**Example:**
```json
{
  "dictionary_attribute_id": "ee0e8400-e29b-41d4-a716-446655440009",
  "dictionary_id": "cc0e8400-e29b-41d4-a716-446655440007",
  "attribute_id": "bb0e8400-e29b-41d4-a716-446655440006",
  "sequence_order": 0
}
```

### product_attributes

Maps products to their required attributes (many-to-many relationship).

**Columns:**
| Column | Type | Constraints | Description |
|--------|------|-----------|-------------|
| product_attribute_id | UUID | PRIMARY KEY | Unique mapping identifier |
| product_id | UUID | NOT NULL, FK | Reference to products table |
| attribute_id | UUID | NOT NULL, FK | Reference to attributes table |
| is_required | BOOLEAN | NOT NULL DEFAULT true | Whether this attribute is required for the product |
| sequence_order | INT | NOT NULL DEFAULT 0 | Order within product specification |
| created_at | TIMESTAMPTZ | DEFAULT now() | Creation timestamp |

**Constraints:**
- `fk_product_attributes_product` - Foreign key to products(product_id) ON DELETE CASCADE
- `fk_product_attributes_attribute` - Foreign key to attributes(attribute_id) ON DELETE CASCADE
- `unique_product_attribute` - Unique constraint on (product_id, attribute_id)

**Indexes:**
- `idx_product_attributes_product` - On product_id
- `idx_product_attributes_attribute` - On attribute_id

**Example:**
```json
{
  "product_attribute_id": "ff0e8400-e29b-41d4-a716-446655440010",
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "attribute_id": "990e8400-e29b-41d4-a716-446655440004",
  "is_required": true,
  "sequence_order": 0
}
```

---

## Data Flow Architecture

### Attribute Data Lineage

Each attribute defines a complete data flow path:

```
Primary Source URL
    ↓
Secondary Source URL (fallback)
    ↓
Tertiary Source URL (fallback)
    ↓
[Data Processing/Transformation]
    ↓
Primary Sink URL (Persistence)
```

**Example Flow:**
```
KYC System (primary source)
    ↓
External Provider Backup (secondary)
    ↓
Cache Layer (tertiary)
    ↓
PostgreSQL Vault (primary sink)
```

### Composition Hierarchy

```
Service
├── Product 1
│   ├── Attribute A (required)
│   ├── Attribute B (optional)
│   └── Attribute C (required)
├── Product 2
│   ├── Attribute D (required)
│   └── Attribute E (optional)
└── Product 3
    └── Attribute F (required)
```

---

## Query Examples

### Find all products required by a service

```sql
SELECT DISTINCT p.*
FROM products p
JOIN service_products sp ON p.product_id = sp.product_id
WHERE sp.service_id = $1
ORDER BY sp.sequence_order;
```

### Find all attributes required by a product

```sql
SELECT a.*, pa.is_required, pa.sequence_order
FROM attributes a
JOIN product_attributes pa ON a.attribute_id = pa.attribute_id
WHERE pa.product_id = $1
ORDER BY pa.sequence_order;
```

### Find all private attributes

```sql
SELECT * FROM attributes
WHERE is_private = true AND is_active = true
ORDER BY private_type, name;
```

### Get complete service composition (service → products → attributes)

```sql
SELECT
    s.service_id, s.name as service_name,
    p.product_id, p.name as product_name,
    a.attribute_id, a.name as attribute_name,
    a.is_private, a.private_type,
    a.primary_source_url, a.primary_sink_url
FROM services s
LEFT JOIN service_products sp ON s.service_id = sp.service_id
LEFT JOIN products p ON sp.product_id = p.product_id
LEFT JOIN product_attributes pa ON p.product_id = pa.product_id
LEFT JOIN attributes a ON pa.attribute_id = a.attribute_id
WHERE s.service_id = $1
ORDER BY sp.sequence_order, pa.sequence_order;
```

### Find attributes that need to be synced to a specific sink

```sql
SELECT * FROM attributes
WHERE primary_sink_url = $1 AND is_active = true
ORDER BY name;
```

### Find derived private attributes

```sql
SELECT * FROM attributes
WHERE is_private = true AND private_type = 'derived' AND is_active = true
ORDER BY name;
```

---

## Constraints & Business Rules

1. **Unique Names**: Products, services, and dictionaries must have unique names (case-sensitive)
2. **Non-empty Strings**: All text fields cannot be empty or contain only whitespace
3. **Private Type Logic**: 
   - If `private_type` is set, `is_private` must be true
   - `private_type` can only be 'derived' or 'given'
4. **Required Sink URL**: All attributes must have a `primary_sink_url` specified
5. **Soft Deletes**: All tables support logical deletion via `is_active` flag
6. **Cascade Deletes**: Deleting a service/dictionary cascades to relationship tables
7. **Optional Foreign Keys**: Resources and dictionaries may not reference related tables

---

## Go Domain Models

The following structs are defined in `internal/store/store.go`:

```go
type Product struct {
    ProductID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    IsActive    bool
}

type Service struct {
    ServiceID   string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    IsActive    bool
}

type ProdResource struct {
    ResourceID   string
    Name         string
    Description  string
    Owner        string
    DictionaryID *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    IsActive     bool
}

type Dictionary struct {
    DictionaryID string
    Name         string
    Description  string
    AttributeID  *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    IsActive     bool
}

type Attribute struct {
    AttributeID         string
    Name                string
    DetailedDescription string
    IsPrivate           bool
    PrivateType         *string // "derived" or "given"
    PrimarySourceURL    *string
    SecondarySourceURL  *string
    TertiarySourceURL   *string
    DataMaskType        string
    PrimarySinkURL      string
    CreatedAt           time.Time
    UpdatedAt           time.Time
    IsActive            bool
}
```

---

## Store Methods

The `Store` interface in `internal/store/store.go` provides the following methods:

### Product Operations
- `CreateProduct(ctx, name, description) → *Product`
- `GetProductByName(ctx, name) → *Product`
- `GetProductByID(ctx, productID) → *Product`
- `ListProducts(ctx) → []*Product`

### Service Operations
- `CreateService(ctx, name, description) → *Service`
- `GetServiceByName(ctx, name) → *Service`
- `GetServiceByID(ctx, serviceID) → *Service`
- `ListServices(ctx) → []*Service`

### Resource Operations
- `CreateProdResource(ctx, name, description, owner, dictionaryID) → *ProdResource`
- `GetProdResourceByName(ctx, name) → *ProdResource`
- `GetProdResourceByID(ctx, resourceID) → *ProdResource`
- `ListProdResources(ctx) → []*ProdResource`

### Attribute Operations
- `CreateAttribute(ctx, attr) → *Attribute`
- `GetAttributeByName(ctx, name) → *Attribute`
- `GetAttributeByID(ctx, attributeID) → *Attribute`
- `ListAttributes(ctx) → []*Attribute`
- `ListPrivateAttributes(ctx) → []*Attribute`

### Dictionary Operations
- `CreateDictionary(ctx, name, description, attributeID) → *Dictionary`
- `GetDictionaryByName(ctx, name) → *Dictionary`
- `GetDictionaryByID(ctx, dictionaryID) → *Dictionary`
- `ListDictionaries(ctx) → []*Dictionary`

### Relationship Operations
- `AddAttributeToDictionary(ctx, dictionaryID, attributeID, sequenceOrder) → error`
- `AddProductToService(ctx, serviceID, productID, sequenceOrder) → error`
- `AddAttributeToProduct(ctx, productID, attributeID, isRequired, sequenceOrder) → error`
- `GetServiceComposition(ctx, serviceID) → map[string]interface{}`

---

## Migration Notes

To integrate these new tables:

1. Run `./dsl-poc init-db` to initialize the schema
2. The `initSQL` in `store.go` contains all table definitions
3. Tables are created with `CREATE TABLE IF NOT EXISTS` for idempotency
4. All timestamps use UTC timezone
5. All IDs are UUIDs generated server-side

---

## Performance Considerations

- **Indexes on frequently queried columns**: name, is_active, is_private, private_type
- **Composite indexes**: (cbu_id, created_at DESC) for DSL lookups
- **Cascade deletes**: Deleting services/dictionaries automatically cleans relationship tables
- **Unique constraints**: Prevent duplicate names and ensure data integrity
- **Soft deletes**: Logical deletion via is_active allows historical queries

---

## Future Enhancements

1. **Add audit logging**: Track who modified records and when
2. **Add approval workflows**: Gate creation/modification of sensitive attributes
3. **Add data quality metrics**: Track attribute freshness and accuracy
4. **Add versioning**: Create attribute versions with change tracking
5. **Add tagging**: Allow flexible metadata tagging of attributes and resources
6. **Add access control**: Define who can view/modify specific attributes (privacy levels)

---

**Last Updated**: November 2024
**Version**: 1.0
**Status**: Production Ready
