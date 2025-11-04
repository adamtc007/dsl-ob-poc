-- +goose Down
-- Drop hedge fund investor schema and all related objects

SET search_path TO "hf-investor", public;

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_apply_register_event ON hf_register_events;
DROP FUNCTION IF EXISTS apply_register_event();

-- Drop indexes explicitly (though they'll cascade anyway)
DROP INDEX IF EXISTS idx_investors_status;
DROP INDEX IF EXISTS idx_trades_pending;
DROP INDEX IF EXISTS idx_events_lot_date_ts;
DROP INDEX IF EXISTS idx_docs_status;
DROP INDEX IF EXISTS idx_docreqs_open;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS hf_audit_events;
DROP TABLE IF EXISTS hf_dsl_executions;
DROP TABLE IF EXISTS hf_lifecycle_states;
DROP TABLE IF EXISTS hf_register_events;
DROP TABLE IF EXISTS hf_register_lots;
DROP TABLE IF EXISTS hf_trades;
DROP TABLE IF EXISTS hf_bank_instructions;
DROP TABLE IF EXISTS hf_document_requirements;
DROP TABLE IF EXISTS hf_documents;
DROP TABLE IF EXISTS hf_tax_profiles;
DROP TABLE IF EXISTS hf_kyc_profiles;
DROP TABLE IF EXISTS hf_beneficial_owners;
DROP TABLE IF EXISTS hf_investors;
DROP TABLE IF EXISTS hf_series;
DROP TABLE IF EXISTS hf_share_classes;
DROP TABLE IF EXISTS hf_funds;

-- Drop the schema (only if empty)
DROP SCHEMA IF EXISTS "hf-investor";