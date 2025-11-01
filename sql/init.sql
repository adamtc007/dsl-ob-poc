-- Creates the schema and tables for the DSL POC
CREATE SCHEMA IF NOT EXISTS "kyc-dsl";

-- Table to store immutable, versioned DSL files
CREATE TABLE IF NOT EXISTS "kyc-dsl".dsl_ob (
    -- Use a UUID for the version ID
    version_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The CBU this DSL version belongs to
    cbu_id VARCHAR(255) NOT NULL,

    -- The full S-expression DSL text
    dsl_text TEXT NOT NULL,

    -- Timestamp for ordering
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

-- Index for fast lookups of the latest DSL for a CBU
CREATE INDEX IF NOT EXISTS idx_dsl_ob_cbu_id_created_at
ON "kyc-dsl".dsl_ob (cbu_id, created_at DESC);

-- ============================================================================
-- NEW TABLES FOR PRODUCT CATALOG, SERVICES, RESOURCES, AND METADATA
-- ============================================================================

-- Products table: Core product definitions
CREATE TABLE IF NOT EXISTS "kyc-dsl".products (
    product_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_products_name ON "kyc-dsl".products (name);

-- Services table: Services that can be offered with or without products
CREATE TABLE IF NOT EXISTS "kyc-dsl".services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_services_name ON "kyc-dsl".services (name);

-- Product Resources table: Resources required by products/services
CREATE TABLE IF NOT EXISTS "kyc-dsl".prod_resources (
    resource_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner VARCHAR(255) NOT NULL,
    dictionary_id UUID,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
    -- Note: dictionary_id will reference dictionaries table via FK after that table is created
);

CREATE INDEX IF NOT EXISTS idx_prod_resources_name ON "kyc-dsl".prod_resources (name);
CREATE INDEX IF NOT EXISTS idx_prod_resources_owner ON "kyc-dsl".prod_resources (owner);

-- Dictionaries table: Master data dictionaries that contain attributes
CREATE TABLE IF NOT EXISTS "kyc-dsl".dictionaries (
    dictionary_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    attribute_id UUID,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
    -- Note: attribute_id will reference attributes table via FK after that table is created
);

CREATE INDEX IF NOT EXISTS idx_dictionaries_name ON "kyc-dsl".dictionaries (name);

-- Attributes table: Detailed attribute definitions with metadata
CREATE TABLE IF NOT EXISTS "kyc-dsl".attributes (
    attribute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    detailed_description TEXT,
    is_private BOOLEAN DEFAULT FALSE,
    private_type VARCHAR(50),  -- 'derived' or 'given', only applicable if is_private = TRUE
    primary_source_url TEXT,
    secondary_source_url TEXT,
    tertiary_source_url TEXT,
    data_type VARCHAR(50) DEFAULT 'string',  -- 'string', 'mask', 'struct', etc.
    primary_sink_url TEXT NOT NULL,  -- Where the attribute is persisted
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    CONSTRAINT chk_private_type CHECK (
        (is_private = FALSE AND private_type IS NULL) OR
        (is_private = TRUE AND private_type IN ('derived', 'given'))
    ),
    CONSTRAINT chk_data_type CHECK (data_type IN ('string', 'mask', 'struct'))
);

CREATE INDEX IF NOT EXISTS idx_attributes_name ON "kyc-dsl".attributes (name);
CREATE INDEX IF NOT EXISTS idx_attributes_is_private ON "kyc-dsl".attributes (is_private);
CREATE INDEX IF NOT EXISTS idx_attributes_private_type ON "kyc-dsl".attributes (private_type);

-- Add foreign key constraints now that all tables exist
ALTER TABLE "kyc-dsl".prod_resources
ADD CONSTRAINT fk_prod_resources_dictionary
FOREIGN KEY (dictionary_id) REFERENCES "kyc-dsl".dictionaries (dictionary_id) ON DELETE SET NULL;

ALTER TABLE "kyc-dsl".dictionaries
ADD CONSTRAINT fk_dictionaries_attribute
FOREIGN KEY (attribute_id) REFERENCES "kyc-dsl".attributes (attribute_id) ON DELETE SET NULL;
