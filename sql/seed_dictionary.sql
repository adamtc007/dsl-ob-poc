-- ============================================================================
-- DICTIONARY SEED SCRIPT
-- ============================================================================
--
-- This script populates the dictionary table with initial attributes for
-- the DSL onboarding system. These attributes represent the metadata-driven
-- type system where each attribute has source and sink metadata.
--
-- Usage:
--   psql "$DB_CONN_STRING" -f sql/seed_dictionary.sql
--
-- OR (if DB_CONN_STRING is set):
--   psql -d your_database -f sql/seed_dictionary.sql
--
-- This script is IDEMPOTENT - it can be run multiple times safely.
-- ============================================================================

\echo '═══════════════════════════════════════════════════════════════════════'
\echo 'SEEDING DICTIONARY TABLE WITH ATTRIBUTES'
\echo '═══════════════════════════════════════════════════════════════════════'
\echo ''

-- ----------------------------------------------------------------------------
-- 1. VERIFY DICTIONARY TABLE EXISTS
-- ----------------------------------------------------------------------------
\echo '1. Verifying dictionary table exists...'

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'dsl-ob-poc' AND table_name = 'dictionary'
    ) THEN
        RAISE EXCEPTION 'Dictionary table does not exist. Please run: psql -f sql/init.sql first';
    END IF;
END
$$;

\echo '   ✓ Dictionary table found'
\echo ''

-- ----------------------------------------------------------------------------
-- 2. INSERT DICTIONARY ATTRIBUTES
-- ----------------------------------------------------------------------------
\echo '2. Inserting/updating dictionary attributes...'
\echo ''

-- Onboarding CBU ID
INSERT INTO "dsl-ob-poc".dictionary (
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink
) VALUES (
    'onboard.cbu_id',
    'Client Business Unit identifier for onboarding case tracking and workflow management',
    'Onboarding',
    'string',
    'Onboarding',
    '{"type": "manual", "url": "https://onboarding.example.com/cbu", "required": true, "format": "CBU-[0-9]+"}'::jsonb,
    '{"type": "database", "url": "postgres://onboarding_db/cases", "table": "onboarding_cases", "field": "cbu_id"}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
    long_description = EXCLUDED.long_description,
    group_id = EXCLUDED.group_id,
    mask = EXCLUDED.mask,
    domain = EXCLUDED.domain,
    source = EXCLUDED.source,
    sink = EXCLUDED.sink,
    updated_at = (now() at time zone 'utc');

\echo '   ✓ Inserted/updated: onboard.cbu_id'

-- Entity Legal Name (KYC)
INSERT INTO "dsl-ob-poc".dictionary (
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink
) VALUES (
    'entity.legal_name',
    'Legal name of the entity for KYC purposes',
    'KYC',
    'string',
    'KYC',
    '{"type": "manual", "url": "https://kyc.example.com/entity", "required": true}'::jsonb,
    '{"type": "database", "url": "postgres://kyc_db/entities", "table": "legal_entities", "field": "legal_name"}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
    long_description = EXCLUDED.long_description,
    group_id = EXCLUDED.group_id,
    mask = EXCLUDED.mask,
    domain = EXCLUDED.domain,
    source = EXCLUDED.source,
    sink = EXCLUDED.sink,
    updated_at = (now() at time zone 'utc');

\echo '   ✓ Inserted/updated: entity.legal_name'

-- Custody Account Number
INSERT INTO "dsl-ob-poc".dictionary (
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink
) VALUES (
    'custody.account_number',
    'Custody account identifier for asset safekeeping',
    'CustodyAccount',
    'string',
    'Custody',
    '{"type": "api", "url": "https://custody.example.com/accounts", "method": "GET"}'::jsonb,
    '{"type": "database", "url": "postgres://custody_db/accounts", "table": "accounts", "field": "account_number"}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
    long_description = EXCLUDED.long_description,
    group_id = EXCLUDED.group_id,
    mask = EXCLUDED.mask,
    domain = EXCLUDED.domain,
    source = EXCLUDED.source,
    sink = EXCLUDED.sink,
    updated_at = (now() at time zone 'utc');

\echo '   ✓ Inserted/updated: custody.account_number'

-- Entity Domicile
INSERT INTO "dsl-ob-poc".dictionary (
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink
) VALUES (
    'entity.domicile',
    'Domicile jurisdiction of the fund or entity',
    'KYC',
    'string',
    'KYC',
    '{"type": "registry", "url": "https://registry.example.com/jurisdictions", "validated": true}'::jsonb,
    '{"type": "database", "url": "postgres://kyc_db/entities", "table": "entities", "field": "domicile"}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
    long_description = EXCLUDED.long_description,
    group_id = EXCLUDED.group_id,
    mask = EXCLUDED.mask,
    domain = EXCLUDED.domain,
    source = EXCLUDED.source,
    sink = EXCLUDED.sink,
    updated_at = (now() at time zone 'utc');

\echo '   ✓ Inserted/updated: entity.domicile'

-- Security ISIN
INSERT INTO "dsl-ob-poc".dictionary (
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink
) VALUES (
    'security.isin',
    'International Securities Identification Number',
    'Security',
    'string',
    'Trading',
    '{"type": "api", "url": "https://isin-registry.example.com/lookup", "authoritative": true}'::jsonb,
    '{"type": "database", "url": "postgres://trading_db/securities", "table": "securities", "field": "isin"}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
    long_description = EXCLUDED.long_description,
    group_id = EXCLUDED.group_id,
    mask = EXCLUDED.mask,
    domain = EXCLUDED.domain,
    source = EXCLUDED.source,
    sink = EXCLUDED.sink,
    updated_at = (now() at time zone 'utc');

\echo '   ✓ Inserted/updated: security.isin'

-- Accounting NAV Value
INSERT INTO "dsl-ob-poc".dictionary (
    name,
    long_description,
    group_id,
    mask,
    domain,
    source,
    sink
) VALUES (
    'accounting.nav_value',
    'Net Asset Value calculated daily',
    'FundAccounting',
    'string',
    'Accounting',
    '{"type": "calculated", "formula": "total_assets - total_liabilities", "frequency": "daily"}'::jsonb,
    '{"type": "database", "url": "postgres://accounting_db/nav", "table": "daily_nav", "field": "nav_value"}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
    long_description = EXCLUDED.long_description,
    group_id = EXCLUDED.group_id,
    mask = EXCLUDED.mask,
    domain = EXCLUDED.domain,
    source = EXCLUDED.source,
    sink = EXCLUDED.sink,
    updated_at = (now() at time zone 'utc');

\echo '   ✓ Inserted/updated: accounting.nav_value'

\echo ''

-- ----------------------------------------------------------------------------
-- 3. VERIFY INSERTION
-- ----------------------------------------------------------------------------
\echo '3. Verification - Counting inserted attributes...'
\echo ''

SELECT
    COUNT(*) as total_attributes,
    COUNT(DISTINCT group_id) as unique_groups,
    COUNT(DISTINCT domain) as unique_domains
FROM "dsl-ob-poc".dictionary;

\echo ''

-- ----------------------------------------------------------------------------
-- 4. SHOW INSERTED ATTRIBUTES
-- ----------------------------------------------------------------------------
\echo '4. Inserted Attributes Summary:'
\echo ''

SELECT
    name,
    group_id,
    domain,
    mask,
    LEFT(long_description, 50) || '...' as description
FROM "dsl-ob-poc".dictionary
ORDER BY domain, name;

\echo ''

-- ----------------------------------------------------------------------------
-- 5. COMPLETION SUMMARY
-- ----------------------------------------------------------------------------
\echo '═══════════════════════════════════════════════════════════════════════'
\echo 'DICTIONARY SEEDING COMPLETE'
\echo '═══════════════════════════════════════════════════════════════════════'
\echo ''
\echo 'Successfully seeded 6 attributes across multiple domains:'
\echo '  • Onboarding: onboard.cbu_id'
\echo '  • KYC: entity.legal_name, entity.domicile'
\echo '  • Custody: custody.account_number'
\echo '  • Trading: security.isin'
\echo '  • Accounting: accounting.nav_value'
\echo ''
\echo 'These attributes support the AttributeID-as-Type architectural pattern'
\echo 'where each UUID references metadata-driven type information.'
\echo ''
\echo 'Next Steps:'
\echo '  1. Verify: SELECT * FROM "dsl-ob-poc".dictionary;'
\echo '  2. Test: ./dsl-poc create --cbu="CBU-TEST-001"'
\echo '  3. Explore: ./dsl-poc history --cbu="CBU-TEST-001"'
\echo ''
\echo '═══════════════════════════════════════════════════════════════════════'
