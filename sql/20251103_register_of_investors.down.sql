-- Down Migration: Drop Register of Investors Schema
-- Created: 2025-11-03
-- Purpose: Rollback hedge fund investor management system

-- Set search path for rollback
SET search_path TO "register", "public";

-- ============================================================================
-- Drop Views (must be dropped before tables they depend on)
-- ============================================================================

DROP VIEW IF EXISTS trading_activity_v;
DROP VIEW IF EXISTS kyc_status_v;
DROP VIEW IF EXISTS register_of_investors_v;

-- ============================================================================
-- Drop Functions
-- ============================================================================

DROP FUNCTION IF EXISTS create_sample_fund(TEXT, TEXT);
DROP FUNCTION IF EXISTS log_lifecycle_transition();
DROP FUNCTION IF EXISTS update_register_lot_from_event();
DROP FUNCTION IF EXISTS update_investor_timestamp();

-- ============================================================================
-- Drop Audit and Compliance Tables
-- ============================================================================

-- Drop audit_event indexes and table
DROP INDEX IF EXISTS idx_audit_event_actor_time;
DROP INDEX IF EXISTS idx_audit_event_entity_time;
DROP TABLE IF EXISTS audit_event;

-- Drop document indexes and table
DROP INDEX IF EXISTS idx_document_expires;
DROP INDEX IF EXISTS idx_document_type;
DROP INDEX IF EXISTS idx_document_subject;
DROP TABLE IF EXISTS document;

-- ============================================================================
-- Drop Lifecycle State Management
-- ============================================================================

-- Drop lifecycle state journal indexes and table
DROP INDEX IF EXISTS idx_lifecycle_journal_entity_time;
DROP TRIGGER IF EXISTS lifecycle_state_journal_trigger ON lifecycle_state;
DROP TABLE IF EXISTS lifecycle_state_journal;

-- Drop lifecycle state indexes and table
DROP INDEX IF EXISTS idx_lifecycle_state_entity;
DROP TABLE IF EXISTS lifecycle_state;

-- ============================================================================
-- Drop Event Sourcing Tables
-- ============================================================================

-- Drop register_event indexes, triggers, and table
DROP INDEX IF EXISTS idx_register_event_trade;
DROP INDEX IF EXISTS idx_register_event_type;
DROP INDEX IF EXISTS idx_register_event_lot_time;
DROP TRIGGER IF EXISTS register_event_update_lot ON register_event;
DROP TABLE IF EXISTS register_event;

-- Drop register_lot indexes and table
DROP INDEX IF EXISTS idx_register_lot_class;
DROP INDEX IF EXISTS idx_register_lot_investor;
DROP TABLE IF EXISTS register_lot;

-- ============================================================================
-- Drop Trading and Settlement Tables
-- ============================================================================

-- Drop trade indexes and table
DROP INDEX IF EXISTS idx_trade_type;
DROP INDEX IF EXISTS idx_trade_status;
DROP INDEX IF EXISTS idx_trade_class_date;
DROP INDEX IF EXISTS idx_trade_investor_date;
DROP TRIGGER IF EXISTS trade_updated_at ON trade;
DROP TABLE IF EXISTS trade;

-- ============================================================================
-- Drop Banking Tables
-- ============================================================================

-- Drop bank_instruction indexes and table
DROP INDEX IF EXISTS idx_bank_instruction_active;
DROP INDEX IF EXISTS idx_bank_instruction_investor_currency;
DROP TABLE IF EXISTS bank_instruction;

-- ============================================================================
-- Drop Compliance Tables
-- ============================================================================

-- Drop tax_profile table
DROP TABLE IF EXISTS tax_profile;

-- Drop kyc_profile indexes and table
DROP INDEX IF EXISTS idx_kyc_refresh_due;
DROP INDEX IF EXISTS idx_kyc_status;
DROP TABLE IF EXISTS kyc_profile;

-- ============================================================================
-- Drop Identity Tables
-- ============================================================================

-- Drop beneficial_owner indexes and table
DROP INDEX IF EXISTS idx_beneficial_owner_investor;
DROP TABLE IF EXISTS beneficial_owner;

-- Drop investor indexes, triggers, and table
DROP INDEX IF EXISTS idx_investor_type;
DROP INDEX IF EXISTS idx_investor_status;
DROP INDEX IF EXISTS idx_investor_lei;
DROP INDEX IF EXISTS idx_investor_name_trgm;
DROP TRIGGER IF EXISTS investor_updated_at ON investor;
DROP TABLE IF EXISTS investor;

-- ============================================================================
-- Drop Fund Structure Tables
-- ============================================================================

-- Drop series table
DROP TABLE IF EXISTS series;

-- Drop share_class table
DROP TABLE IF EXISTS share_class;

-- Drop fund table
DROP TABLE IF EXISTS fund;

-- ============================================================================
-- Drop Schema
-- ============================================================================

-- Reset search path before dropping schema
RESET search_path;

-- Drop the register schema (will fail if other objects exist)
DROP SCHEMA IF EXISTS "register" CASCADE;

-- ============================================================================
-- Drop Extensions (Optional - only if not used elsewhere)
-- ============================================================================

-- Note: Commented out to avoid breaking other applications
-- Only uncomment if you're sure these extensions aren't used elsewhere

-- DROP EXTENSION IF EXISTS pg_trgm;
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Rollback successful
-- All Register of Investors components have been removed