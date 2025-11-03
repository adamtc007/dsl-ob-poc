# Database Schema Documentation

## Overview

Database schema for the DSL Onboarding POC implementing an immutable, versioned state machine with comprehensive entity relationship management and attribute-driven configuration.

## Schema: `"dsl-ob-poc"`

### Core Architecture

```
Event Sourcing Core
├── dsl_ob (immutable versioned DSL records)
├── attribute_values (runtime values with versioning)

Catalog Tables
├── products (core product definitions)
├── services (services offered with products)
├── prod_resources (production resources)
├── product_services (many-to-many relationship)
├── service_resources (many-to-many relationship)

Entity Relationship Model
├── cbus (Client Business Units)
├── roles (entity roles within CBUs)
├── entity_types (entity type definitions)
├── entities (central entity registry)
├── cbu_entity_roles (CBU-entity-role relationships)
├── entity_limited_companies (limited company details)
├── entity_partnerships (partnership details)
├── entity_individuals (individual person details)

Data Dictionary
├── dictionary (attribute definitions with JSONB metadata)
```

## Core Tables

### `dsl_ob` - Event Sourcing Core
Immutable versioned DSL records implementing event sourcing pattern.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| version_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique version identifier |
| cbu_id | VARCHAR(255) | NOT NULL | Client Business Unit identifier |
| dsl_text | TEXT | NOT NULL | S-expression DSL content |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |

**Indexes:**
- `idx_dsl_ob_cbu_id_created_at` - Composite index (cbu_id, created_at DESC) for fast latest lookups

### `attribute_values` - Runtime Values
Stores resolved attribute values with versioning for deterministic DSL generation.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| av_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique identifier |
| cbu_id | UUID | NOT NULL, FK to cbus(cbu_id) | CBU reference |
| dsl_ob_id | UUID | NULLABLE | Optional precise DSL row reference |
| dsl_version | INTEGER | NOT NULL | DSL runbook version |
| attribute_id | UUID | NOT NULL, FK to dictionary(attribute_id) | Attribute reference |
| value | JSONB | NOT NULL | Resolved attribute value |
| state | TEXT | NOT NULL, DEFAULT 'resolved' | 'pending', 'resolved', 'invalid' |
| source | JSONB | NULLABLE | Provenance metadata |
| observed_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Observation timestamp |

**Constraints:**
- `UNIQUE (cbu_id, dsl_version, attribute_id)` - One value per attribute per version

**Indexes:**
- `idx_attr_vals_lookup` - Composite index (cbu_id, attribute_id, dsl_version)

## Entity Relationship Model

### `cbus` - Client Business Units
Central registry for Client Business Units (funds, companies, etc.).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| cbu_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique CBU identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | CBU name |
| description | TEXT | NULLABLE | CBU description |
| nature_purpose | TEXT | NULLABLE | Business nature and purpose |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `roles` - Entity Roles
Defines roles entities can play within CBUs.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| role_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique role identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Role name |
| description | TEXT | NULLABLE | Role description |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

**Common Roles:**
- Investment Manager
- Asset Owner
- SiCAV
- Management Company
- Strategy Owner
- Main Client (Commercial)

### `entity_types` - Entity Type Definitions
Categorizes different types of entities with table references.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| entity_type_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique type identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Entity type name |
| description | TEXT | NULLABLE | Type description |
| table_name | VARCHAR(255) | NOT NULL | Reference to specific entity type table |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `entities` - Central Entity Registry
Central registry linking to specific entity type tables.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| entity_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique entity identifier |
| entity_type_id | UUID | NOT NULL, FK to entity_types(entity_type_id) | Entity type reference |
| external_id | VARCHAR(255) | NULLABLE | Reference to specific entity type table |
| name | VARCHAR(255) | NOT NULL | Entity name |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `cbu_entity_roles` - CBU-Entity-Role Relationships
Links CBUs to entities through specific roles (many-to-many-to-many).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| cbu_entity_role_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique relationship identifier |
| cbu_id | UUID | NOT NULL, FK to cbus(cbu_id) | CBU reference |
| entity_id | UUID | NOT NULL, FK to entities(entity_id) | Entity reference |
| role_id | UUID | NOT NULL, FK to roles(role_id) | Role reference |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |

**Constraints:**
- `UNIQUE (cbu_id, entity_id, role_id)` - Prevents duplicate role assignments

## Entity Type Tables

### `entity_limited_companies` - Limited Company Details

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| limited_company_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique company identifier |
| company_name | VARCHAR(255) | NOT NULL | Company name |
| registration_number | VARCHAR(100) | NULLABLE | Company registration number |
| jurisdiction | VARCHAR(100) | NULLABLE | Jurisdiction of incorporation |
| incorporation_date | DATE | NULLABLE | Date of incorporation |
| registered_address | TEXT | NULLABLE | Registered office address |
| business_nature | TEXT | NULLABLE | Nature of business |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `entity_partnerships` - Partnership Details

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| partnership_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique partnership identifier |
| partnership_name | VARCHAR(255) | NOT NULL | Partnership name |
| partnership_type | VARCHAR(100) | NULLABLE | 'General', 'Limited', 'Limited Liability' |
| jurisdiction | VARCHAR(100) | NULLABLE | Jurisdiction of formation |
| formation_date | DATE | NULLABLE | Date of formation |
| principal_place_business | TEXT | NULLABLE | Principal place of business |
| partnership_agreement_date | DATE | NULLABLE | Partnership agreement date |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `entity_individuals` - Individual Person Details

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| individual_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique individual identifier |
| first_name | VARCHAR(255) | NOT NULL | First name |
| last_name | VARCHAR(255) | NOT NULL | Last name |
| middle_names | VARCHAR(255) | NULLABLE | Middle names |
| date_of_birth | DATE | NULLABLE | Date of birth |
| nationality | VARCHAR(100) | NULLABLE | Nationality |
| residence_address | TEXT | NULLABLE | Residence address |
| id_document_type | VARCHAR(100) | NULLABLE | 'Passport', 'National ID', 'Driving License' |
| id_document_number | VARCHAR(100) | NULLABLE | ID document number |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

## Catalog Tables

### `products` - Core Product Definitions

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| product_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique product identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Product name |
| description | TEXT | NULLABLE | Product description |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `services` - Service Definitions

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| service_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique service identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Service name |
| description | TEXT | NULLABLE | Service description |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

### `prod_resources` - Production Resources

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| resource_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique resource identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Resource name |
| description | TEXT | NULLABLE | Resource description |
| owner | VARCHAR(255) | NOT NULL | Resource owner |
| dictionary_group | VARCHAR(100) | NULLABLE | Associated dictionary group |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

## Data Dictionary

### `dictionary` - Attribute Definitions
Central data dictionary with JSONB metadata for rich attribute definitions.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| attribute_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique attribute identifier |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Attribute variable name |
| long_description | TEXT | NULLABLE | Human-readable description |
| group_id | VARCHAR(100) | NOT NULL, DEFAULT 'default' | Dictionary group |
| mask | VARCHAR(50) | DEFAULT 'string' | Data mask type |
| domain | VARCHAR(100) | NULLABLE | Domain classification |
| vector | TEXT | NULLABLE | AI semantic search vector |
| source | JSONB | NULLABLE | Source metadata |
| sink | JSONB | NULLABLE | Sink metadata |
| created_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT (now() at time zone 'utc') | Last update timestamp |

**JSONB Metadata Examples:**
```json
// Source metadata
{
  "type": "manual",
  "required": true,
  "format": "CBU-[0-9]+"
}

// Sink metadata
{
  "type": "database",
  "table": "onboarding_cases"
}
```

## Relationship Tables

### `product_services` - Product-Service Relationships

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| product_id | UUID | NOT NULL, FK to products(product_id) | Product reference |
| service_id | UUID | NOT NULL, FK to services(service_id) | Service reference |

**Constraints:**
- `PRIMARY KEY (product_id, service_id)`

### `service_resources` - Service-Resource Relationships

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| service_id | UUID | NOT NULL, FK to services(service_id) | Service reference |
| resource_id | UUID | NOT NULL, FK to prod_resources(resource_id) | Resource reference |

**Constraints:**
- `PRIMARY KEY (service_id, resource_id)`

## Performance Optimizations

### Indexing Strategy
- **Composite indexes** on frequently queried combinations
- **Time-series optimization** with (created_at DESC) for latest lookups
- **Foreign key indexes** for efficient JOIN operations

### Query Patterns
- **Latest DSL lookup**: `(cbu_id, created_at DESC)` index
- **Attribute resolution**: `(cbu_id, attribute_id, dsl_version)` index
- **Entity relationships**: Multi-table JOINs through relationship tables

### Data Integrity
- **CASCADE DELETE** for dependent relationships
- **UNIQUE constraints** preventing data duplication
- **NOT NULL constraints** ensuring data completeness
- **Foreign key constraints** maintaining referential integrity