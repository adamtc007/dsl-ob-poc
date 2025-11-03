/*
v3: Refactors the attributes table to be the central dictionary.
- Uses JSONB to store rich, complex metadata for sources and sinks.
- Renames to 'dictionary' as it's the master table.
- Removes the old 'dictionaries' and 'dictionary_attributes' tables,
  as an attribute's 'dictionary_id' (now 'group_id') is just a string for grouping.
- **Sets main schema to "dsl-ob-poc"**
*/
CREATE SCHEMA IF NOT EXISTS "dsl-ob-poc";

-- Table to store immutable, versioned DSL files
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".dsl_ob (
    version_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cbu_id VARCHAR(255) NOT NULL,
    dsl_text TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_dsl_ob_cbu_id_created_at
ON "dsl-ob-poc".dsl_ob (cbu_id, created_at DESC);

-- CBU table: Client Business Unit definitions
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".cbus (
    cbu_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    nature_purpose TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_cbus_name ON "dsl-ob-poc".cbus (name);

-- Products table: Core product definitions
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".products (
    product_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_products_name ON "dsl-ob-poc".products (name);

-- Services table: Services that can be offered with or without products
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_services_name ON "dsl-ob-poc".services (name);

-- Product <-> Service Join Table
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".product_services (
    product_id UUID NOT NULL REFERENCES "dsl-ob-poc".products (product_id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES "dsl-ob-poc".services (service_id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, service_id)
);

-- ============================================================================
-- DICTIONARY AND RESOURCE TABLES (REFACTORED)
-- ============================================================================

-- Master Data Dictionary (Attributes table)
-- This is the central pillar.
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".dictionary (
    attribute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- The unique "variable name" for the DSL, e.g., "entity.legal_name"
    name VARCHAR(255) NOT NULL UNIQUE,

    -- Description for AI agent discovery and human readability
    long_description TEXT,

    -- The "dictionary" this attribute belongs to (e.g., "KYC", "Settlement")
    -- This replaces the old 'dictionaries' table.
    group_id VARCHAR(100) NOT NULL DEFAULT 'default',

    -- Metadata
    mask VARCHAR(50) DEFAULT 'string', -- 'string', 'ssn', 'date'
    domain VARCHAR(100), -- 'KYC', 'AML', 'Trading', 'Settlement'
    vector TEXT,         -- For AI semantic search

    -- Rich metadata stored as JSON
    source JSONB,        -- See SourceMetadata struct in Go
    sink JSONB,          -- See SinkMetadata struct in Go

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_dictionary_name ON "dsl-ob-poc".dictionary (name);
CREATE INDEX IF NOT EXISTS idx_dictionary_group_id ON "dsl-ob-poc".dictionary (group_id);
CREATE INDEX IF NOT EXISTS idx_dictionary_domain ON "dsl-ob-poc".dictionary (domain);

-- Attribute Values table: Runtime values for onboarding instances
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".attribute_values (
    av_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cbu_id        UUID NOT NULL REFERENCES "dsl-ob-poc".cbus(cbu_id),
    dsl_ob_id     UUID,                  -- optional: reference precise DSL row, if you store dsl_ob.id
    dsl_version   INTEGER NOT NULL,      -- tie values to the exact runbook snapshot
    attribute_id  UUID NOT NULL REFERENCES "dsl-ob-poc".dictionary (attribute_id) ON DELETE CASCADE,
    value         JSONB NOT NULL,
    state         TEXT NOT NULL DEFAULT 'resolved', -- 'pending' | 'resolved' | 'invalid'
    source        JSONB,                 -- provenance (table/column/system/collector)
    observed_at   TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (cbu_id, dsl_version, attribute_id)
);

CREATE INDEX IF NOT EXISTS idx_attr_vals_lookup ON "dsl-ob-poc".attribute_values (cbu_id, attribute_id, dsl_version);

-- Production Resources table
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".prod_resources (
    resource_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    owner VARCHAR(255) NOT NULL,

    -- A resource is now defined by its "dictionary_group"
    -- This replaces the foreign key to the old 'dictionaries' table.
    -- e.g., "CustodyAccount" resource uses the "CustodyAccount" group_id.
    dictionary_group VARCHAR(100),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_prod_resources_name ON "dsl-ob-poc".prod_resources (name);
CREATE INDEX IF NOT EXISTS idx_prod_resources_owner ON "dsl-ob-poc".prod_resources (owner);
CREATE INDEX IF NOT EXISTS idx_prod_resources_dict_group ON "dsl-ob-poc".prod_resources (dictionary_group);


-- Service <-> Resource Join Table
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".service_resources (
    service_id UUID NOT NULL REFERENCES "dsl-ob-poc".services (service_id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES "dsl-ob-poc".prod_resources (resource_id) ON DELETE CASCADE,
    PRIMARY KEY (service_id, resource_id)
);

-- ============================================================================
-- ENTITY RELATIONSHIP MODEL
-- ============================================================================

-- Roles table: Defines roles entities can play within a CBU
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".roles (
    role_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_roles_name ON "dsl-ob-poc".roles (name);

-- Entity Types table: Defines the different types of entities
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".entity_types (
    entity_type_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    table_name VARCHAR(255) NOT NULL, -- Points to specific entity type table
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_entity_types_name ON "dsl-ob-poc".entity_types (name);
CREATE INDEX IF NOT EXISTS idx_entity_types_table ON "dsl-ob-poc".entity_types (table_name);

-- Entities table: Central entity registry
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".entities (
    entity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type_id UUID NOT NULL REFERENCES "dsl-ob-poc".entity_types (entity_type_id) ON DELETE CASCADE,
    external_id VARCHAR(255), -- Reference to the specific entity type table
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_entities_type ON "dsl-ob-poc".entities (entity_type_id);
CREATE INDEX IF NOT EXISTS idx_entities_external_id ON "dsl-ob-poc".entities (external_id);
CREATE INDEX IF NOT EXISTS idx_entities_name ON "dsl-ob-poc".entities (name);

-- CBU Entity Roles table: Links CBUs to entities through roles
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".cbu_entity_roles (
    cbu_entity_role_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cbu_id UUID NOT NULL REFERENCES "dsl-ob-poc".cbus (cbu_id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES "dsl-ob-poc".entities (entity_id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES "dsl-ob-poc".roles (role_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (cbu_id, entity_id, role_id)
);
CREATE INDEX IF NOT EXISTS idx_cbu_entity_roles_cbu ON "dsl-ob-poc".cbu_entity_roles (cbu_id);
CREATE INDEX IF NOT EXISTS idx_cbu_entity_roles_entity ON "dsl-ob-poc".cbu_entity_roles (entity_id);
CREATE INDEX IF NOT EXISTS idx_cbu_entity_roles_role ON "dsl-ob-poc".cbu_entity_roles (role_id);

-- ============================================================================
-- ENTITY TYPE TABLES
-- ============================================================================

-- Limited Company entity type
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".entity_limited_companies (
    limited_company_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100),
    jurisdiction VARCHAR(100),
    incorporation_date DATE,
    registered_address TEXT,
    business_nature TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_limited_companies_reg_num ON "dsl-ob-poc".entity_limited_companies (registration_number);
CREATE INDEX IF NOT EXISTS idx_limited_companies_jurisdiction ON "dsl-ob-poc".entity_limited_companies (jurisdiction);

-- Partnership entity type
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".entity_partnerships (
    partnership_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partnership_name VARCHAR(255) NOT NULL,
    partnership_type VARCHAR(100), -- 'General', 'Limited', 'Limited Liability'
    jurisdiction VARCHAR(100),
    formation_date DATE,
    principal_place_business TEXT,
    partnership_agreement_date DATE,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_partnerships_type ON "dsl-ob-poc".entity_partnerships (partnership_type);
CREATE INDEX IF NOT EXISTS idx_partnerships_jurisdiction ON "dsl-ob-poc".entity_partnerships (jurisdiction);

-- Proper Person (Individual) entity type
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".entity_individuals (
    individual_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    middle_names VARCHAR(255),
    date_of_birth DATE,
    nationality VARCHAR(100),
    residence_address TEXT,
    id_document_type VARCHAR(100), -- 'Passport', 'National ID', 'Driving License'
    id_document_number VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);
CREATE INDEX IF NOT EXISTS idx_individuals_full_name ON "dsl-ob-poc".entity_individuals (last_name, first_name);
CREATE INDEX IF NOT EXISTS idx_individuals_nationality ON "dsl-ob-poc".entity_individuals (nationality);
CREATE INDEX IF NOT EXISTS idx_individuals_id_document ON "dsl-ob-poc".entity_individuals (id_document_type, id_document_number);
