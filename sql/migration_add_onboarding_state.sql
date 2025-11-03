/*
Migration: Add Onboarding State Management
- Adds onboarding state tracking
- Adds incremental version numbering
- Enhances DSL persistence for proper workflow tracking
*/

-- Add onboarding state and version columns to dsl_ob table
ALTER TABLE "dsl-ob-poc".dsl_ob
ADD COLUMN IF NOT EXISTS onboarding_state VARCHAR(50) DEFAULT 'CREATED',
ADD COLUMN IF NOT EXISTS version_number INTEGER DEFAULT 1;

-- Create index for efficient state queries
CREATE INDEX IF NOT EXISTS idx_dsl_ob_state_version
ON "dsl-ob-poc".dsl_ob (cbu_id, onboarding_state, version_number DESC);

-- Create onboarding management table for tracking active onboarding sessions
CREATE TABLE IF NOT EXISTS "dsl-ob-poc".onboarding_sessions (
    onboarding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cbu_id VARCHAR(255) NOT NULL UNIQUE, -- One active onboarding per CBU
    current_state VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    current_version INTEGER NOT NULL DEFAULT 1,
    latest_dsl_version_id UUID REFERENCES "dsl-ob-poc".dsl_ob(version_id),
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_cbu_id ON "dsl-ob-poc".onboarding_sessions (cbu_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_sessions_state ON "dsl-ob-poc".onboarding_sessions (current_state);

-- Create function to automatically update version numbers
CREATE OR REPLACE FUNCTION "dsl-ob-poc".update_dsl_version_number()
RETURNS TRIGGER AS $$
BEGIN
    -- Get the next version number for this CBU
    SELECT COALESCE(MAX(version_number), 0) + 1
    INTO NEW.version_number
    FROM "dsl-ob-poc".dsl_ob
    WHERE cbu_id = NEW.cbu_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-increment version numbers
DROP TRIGGER IF EXISTS trigger_update_dsl_version ON "dsl-ob-poc".dsl_ob;
CREATE TRIGGER trigger_update_dsl_version
    BEFORE INSERT ON "dsl-ob-poc".dsl_ob
    FOR EACH ROW
    EXECUTE FUNCTION "dsl-ob-poc".update_dsl_version_number();

-- Update existing records with version numbers (backfill)
WITH numbered_versions AS (
    SELECT
        version_id,
        ROW_NUMBER() OVER (PARTITION BY cbu_id ORDER BY created_at) as version_num
    FROM "dsl-ob-poc".dsl_ob
    WHERE version_number IS NULL OR version_number = 1
)
UPDATE "dsl-ob-poc".dsl_ob
SET version_number = numbered_versions.version_num
FROM numbered_versions
WHERE "dsl-ob-poc".dsl_ob.version_id = numbered_versions.version_id;

-- Comment documenting onboarding states
COMMENT ON COLUMN "dsl-ob-poc".dsl_ob.onboarding_state IS
'Onboarding progression states: CREATED, PRODUCTS_ADDED, KYC_DISCOVERED, SERVICES_DISCOVERED, RESOURCES_DISCOVERED, ATTRIBUTES_POPULATED, COMPLETED';

COMMENT ON TABLE "dsl-ob-poc".onboarding_sessions IS
'Tracks active onboarding sessions and their current state progression';