-- Migration: Add Entity Types Registry Population
-- Purpose: Populate the entity_types table with all supported entity structures
-- and create initial roles for entity relationships within CBUs

SET search_path TO "dsl-ob-poc";

-- ============================================================================
-- ENTITY TYPES REGISTRY POPULATION
-- ============================================================================

-- Clear existing entity types (in case of re-run)
DELETE FROM entity_types WHERE name IN ('LIMITED_COMPANY', 'PARTNERSHIP', 'INDIVIDUAL', 'TRUST');

-- Insert supported entity types with their corresponding table mappings
INSERT INTO entity_types (entity_type_id, name, description, table_name, created_at, updated_at) VALUES
(
    gen_random_uuid(),
    'LIMITED_COMPANY',
    'Limited liability company or corporation with registered share capital',
    'entity_limited_companies',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'PARTNERSHIP',
    'Partnership structure including General Partnership, Limited Partnership, and LLP',
    'entity_partnerships',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'INDIVIDUAL',
    'Natural person with personal identification and KYC requirements',
    'entity_individuals',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'TRUST',
    'Trust structure with settlors, trustees, beneficiaries, and protectors',
    'entity_trusts',
    now() at time zone 'utc',
    now() at time zone 'utc'
);

-- ============================================================================
-- STANDARD ROLES REGISTRY POPULATION
-- ============================================================================

-- Clear existing roles (in case of re-run)
DELETE FROM roles WHERE name IN (
    'CLIENT', 'UBO', 'AUTHORIZED_SIGNATORY', 'BENEFICIAL_OWNER', 'CONTROLLER',
    'SETTLOR', 'TRUSTEE', 'BENEFICIARY', 'PROTECTOR',
    'GENERAL_PARTNER', 'LIMITED_PARTNER', 'MANAGING_PARTNER',
    'DIRECTOR', 'SHAREHOLDER', 'SECRETARY'
);

-- Insert standard roles that entities can play within CBUs
INSERT INTO roles (role_id, name, description, created_at, updated_at) VALUES
-- General Client Roles
(
    gen_random_uuid(),
    'CLIENT',
    'Primary client entity for the onboarding case',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'UBO',
    'Ultimate Beneficial Owner of the client entity',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'AUTHORIZED_SIGNATORY',
    'Person authorized to act on behalf of the entity',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'BENEFICIAL_OWNER',
    'Person with beneficial ownership interest (may not reach UBO threshold)',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'CONTROLLER',
    'Person exercising control over the entity',
    now() at time zone 'utc',
    now() at time zone 'utc'
),

-- Trust-Specific Roles
(
    gen_random_uuid(),
    'SETTLOR',
    'Person who created the trust and transferred assets to it',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'TRUSTEE',
    'Person or entity responsible for managing the trust according to its terms',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'BENEFICIARY',
    'Person entitled to benefit from the trust',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'PROTECTOR',
    'Person with power to oversee trustees or veto certain decisions',
    now() at time zone 'utc',
    now() at time zone 'utc'
),

-- Partnership-Specific Roles
(
    gen_random_uuid(),
    'GENERAL_PARTNER',
    'Partner with unlimited liability and management control',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'LIMITED_PARTNER',
    'Partner with limited liability and no management control',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'MANAGING_PARTNER',
    'Partner designated to manage the partnership operations',
    now() at time zone 'utc',
    now() at time zone 'utc'
),

-- Corporate-Specific Roles
(
    gen_random_uuid(),
    'DIRECTOR',
    'Member of the board of directors with fiduciary duties',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'SHAREHOLDER',
    'Owner of shares in the company',
    now() at time zone 'utc',
    now() at time zone 'utc'
),
(
    gen_random_uuid(),
    'SECRETARY',
    'Company secretary responsible for compliance and governance',
    now() at time zone 'utc',
    now() at time zone 'utc'
);

-- ============================================================================
-- VALIDATION QUERIES
-- ============================================================================

-- Verify entity types were inserted correctly
DO $$
DECLARE
    entity_count INTEGER;
    role_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO entity_count FROM entity_types;
    SELECT COUNT(*) INTO role_count FROM roles;

    RAISE NOTICE 'Migration completed successfully:';
    RAISE NOTICE '  - Entity Types: % records', entity_count;
    RAISE NOTICE '  - Roles: % records', role_count;

    IF entity_count < 4 THEN
        RAISE EXCEPTION 'Entity types migration failed - insufficient records';
    END IF;

    IF role_count < 14 THEN
        RAISE EXCEPTION 'Roles migration failed - insufficient records';
    END IF;
END $$;

-- Display the entity types and their corresponding tables
SELECT
    name as entity_type,
    table_name,
    description
FROM entity_types
ORDER BY name;

-- Display the roles by category
SELECT
    CASE
        WHEN name IN ('CLIENT', 'UBO', 'AUTHORIZED_SIGNATORY', 'BENEFICIAL_OWNER', 'CONTROLLER')
            THEN 'General'
        WHEN name IN ('SETTLOR', 'TRUSTEE', 'BENEFICIARY', 'PROTECTOR')
            THEN 'Trust-Specific'
        WHEN name IN ('GENERAL_PARTNER', 'LIMITED_PARTNER', 'MANAGING_PARTNER')
            THEN 'Partnership-Specific'
        WHEN name IN ('DIRECTOR', 'SHAREHOLDER', 'SECRETARY')
            THEN 'Corporate-Specific'
        ELSE 'Other'
    END as role_category,
    name as role_name,
    description
FROM roles
ORDER BY role_category, name;
