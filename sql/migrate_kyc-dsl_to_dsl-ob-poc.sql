-- Migration: Rename schema from "kyc-dsl" to "dsl-ob-poc"
-- This preserves all existing tables, data, indexes, and constraints.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.schemata WHERE schema_name = 'kyc-dsl'
    ) THEN
        RAISE NOTICE 'Schema "kyc-dsl" does not exist; skipping rename.';
    ELSE
        IF EXISTS (
            SELECT 1 FROM information_schema.schemata WHERE schema_name = 'dsl-ob-poc'
        ) THEN
            RAISE EXCEPTION 'Target schema "dsl-ob-poc" already exists. Aborting migration to avoid conflicts.';
        ELSE
            EXECUTE 'ALTER SCHEMA "kyc-dsl" RENAME TO "dsl-ob-poc"';
            EXECUTE 'ALTER SCHEMA "dsl-ob-poc" OWNER TO "adamtc007"';
            RAISE NOTICE 'Schema renamed to "dsl-ob-poc" and ownership set to "adamtc007"';
        END IF;
    END IF;
END
$$;
