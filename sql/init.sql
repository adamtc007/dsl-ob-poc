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
