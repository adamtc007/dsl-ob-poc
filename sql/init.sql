-- Creates the schema and tables for the DSL POC
CREATE SCHEMA IF NOT EXISTS "dsl-ob-poc";

-- Table to store immutable, versioned DSL files
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".dsl_ob (
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
ON "dsl-ob-poc".dsl_ob (cbu_id, created_at DESC);

-- ============================================================================
-- NEW TABLES FOR PRODUCT CATALOG, SERVICES, RESOURCES, AND METADATA
-- ============================================================================

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

-- Join table: Products to Services
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".product_services (
    product_id UUID NOT NULL,
    service_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (product_id, service_id),
    CONSTRAINT fk_ps_product FOREIGN KEY (product_id) REFERENCES "dsl-ob-poc".products (product_id) ON DELETE CASCADE,
    CONSTRAINT fk_ps_service FOREIGN KEY (service_id) REFERENCES "dsl-ob-poc".services (service_id) ON DELETE CASCADE
);

-- Product Resources table: Resources required by products/services
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".prod_resources (
    resource_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner VARCHAR(255) NOT NULL,
    dictionary_id UUID,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (name)
    -- Note: dictionary_id will reference dictionaries table via FK after that table is created
);

CREATE INDEX IF NOT EXISTS idx_prod_resources_name ON "dsl-ob-poc".prod_resources (name);
CREATE INDEX IF NOT EXISTS idx_prod_resources_owner ON "dsl-ob-poc".prod_resources (owner);

-- Dictionaries table: Master data dictionaries that contain attributes
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".dictionaries (
    dictionary_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    attribute_id UUID,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
    -- Note: attribute_id will reference attributes table via FK after that table is created
);

CREATE INDEX IF NOT EXISTS idx_dictionaries_name ON "dsl-ob-poc".dictionaries (name);

-- Join table: Dictionaries to Attributes
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".dictionary_attributes (
    dictionary_id UUID NOT NULL,
    attribute_id UUID NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (dictionary_id, attribute_id),
    CONSTRAINT fk_da_dictionary FOREIGN KEY (dictionary_id) REFERENCES "dsl-ob-poc".dictionaries (dictionary_id) ON DELETE CASCADE,
    CONSTRAINT fk_da_attribute FOREIGN KEY (attribute_id) REFERENCES "dsl-ob-poc".attributes (attribute_id) ON DELETE CASCADE
);

-- Attributes table: Detailed attribute definitions with metadata
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".attributes (
    attribute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
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

CREATE INDEX IF NOT EXISTS idx_attributes_name ON "dsl-ob-poc".attributes (name);
CREATE INDEX IF NOT EXISTS idx_attributes_is_private ON "dsl-ob-poc".attributes (is_private);
CREATE INDEX IF NOT EXISTS idx_attributes_private_type ON "dsl-ob-poc".attributes (private_type);

-- Add foreign key constraints now that all tables exist
ALTER TABLE "dsl-ob-poc".prod_resources
ADD CONSTRAINT fk_prod_resources_dictionary
FOREIGN KEY (dictionary_id) REFERENCES "dsl-ob-poc".dictionaries (dictionary_id) ON DELETE SET NULL;

ALTER TABLE "dsl-ob-poc".dictionaries
ADD CONSTRAINT fk_dictionaries_attribute
FOREIGN KEY (attribute_id) REFERENCES "dsl-ob-poc".attributes (attribute_id) ON DELETE SET NULL;

-- Join table: Services to Resources
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".service_resources (
    service_id UUID NOT NULL,
    resource_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (service_id, resource_id),
    CONSTRAINT fk_sr_service FOREIGN KEY (service_id) REFERENCES "dsl-ob-poc".services (service_id) ON DELETE CASCADE,
    CONSTRAINT fk_sr_resource FOREIGN KEY (resource_id) REFERENCES "dsl-ob-poc".prod_resources (resource_id) ON DELETE CASCADE
);
