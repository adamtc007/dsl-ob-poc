-- Migration: Register of Investors Schema
-- Created: 2025-11-03
-- Purpose: Hedge Fund Investor Management with Event Sourcing

-- ============================================================================
-- Extensions and Setup
-- ============================================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable trigram similarity for name searches
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Enable JSON operations
CREATE EXTENSION IF NOT EXISTS "plpgsql";

-- Create schema for investor management
CREATE SCHEMA IF NOT EXISTS "register";

-- Set search path for this migration
SET search_path TO "register", "public";

-- ============================================================================
-- Core Domain Tables
-- ============================================================================

-- Fund Structure
CREATE TABLE fund (
    fund_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    legal_name TEXT NOT NULL,
    domicile TEXT NOT NULL,
    admin_name TEXT,
    inception_date DATE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'CLOSED', 'LIQUIDATING')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE share_class (
    class_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fund_id UUID NOT NULL REFERENCES fund(fund_id),
    code TEXT NOT NULL, -- e.g., "A USD", "B EUR"
    currency TEXT NOT NULL, -- ISO 4217
    dealing_freq TEXT NOT NULL, -- "Daily", "Weekly", "Monthly"
    notice_days INTEGER DEFAULT 0,
    lockup_days INTEGER DEFAULT 0,
    management_fee_bps INTEGER DEFAULT 0,
    performance_fee_bps INTEGER DEFAULT 0,
    minimum_investment NUMERIC(28,2),
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(fund_id, code)
);

CREATE TABLE series (
    series_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    class_id UUID NOT NULL REFERENCES share_class(class_id),
    inception_date DATE NOT NULL,
    code TEXT, -- e.g., "2025-Q1"
    status TEXT DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CLOSED')),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Investor Identity and Legal Structure
-- ============================================================================

CREATE TABLE investor (
    investor_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type TEXT NOT NULL CHECK (type IN ('INDIVIDUAL', 'CORPORATE', 'TRUST', 'FOHF', 'NOMINEE')),
    legal_name TEXT NOT NULL,
    lei TEXT, -- Legal Entity Identifier
    registration_number TEXT, -- Company registration number
    domicile TEXT NOT NULL,

    -- Address fields
    address_line1 TEXT,
    address_line2 TEXT,
    address_line3 TEXT,
    address_line4 TEXT,
    city TEXT,
    postal_code TEXT,
    country TEXT NOT NULL,

    -- Status and lifecycle
    status TEXT NOT NULL DEFAULT 'OPPORTUNITY'
        CHECK (status IN ('OPPORTUNITY', 'KYC_PENDING', 'APPROVED', 'ACTIVE', 'REDEEMED', 'OFFBOARDED')),

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Beneficial ownership for KYB compliance
CREATE TABLE beneficial_owner (
    bo_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investor_id UUID NOT NULL REFERENCES investor(investor_id),
    subject_type TEXT NOT NULL CHECK (subject_type IN ('NATURAL_PERSON', 'LEGAL_ENTITY')),
    full_name TEXT NOT NULL,
    date_of_birth DATE,
    country TEXT NOT NULL,
    ownership_pct NUMERIC(5,2) NOT NULL CHECK (ownership_pct >= 0 AND ownership_pct <= 100),

    -- Risk flags
    pep BOOLEAN DEFAULT FALSE, -- Politically Exposed Person
    sanctions_flag BOOLEAN DEFAULT FALSE,

    -- Verification status
    verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Compliance and Risk Management
-- ============================================================================

CREATE TABLE kyc_profile (
    investor_id UUID PRIMARY KEY REFERENCES investor(investor_id),
    risk_rating TEXT NOT NULL DEFAULT 'MEDIUM'
        CHECK (risk_rating IN ('LOW', 'MEDIUM', 'HIGH')),
    status TEXT NOT NULL DEFAULT 'NOT_STARTED'
        CHECK (status IN ('NOT_STARTED', 'IN_PROGRESS', 'APPROVED', 'REJECTED', 'REFRESH_DUE')),

    -- Screening and refresh
    last_screened_at TIMESTAMPTZ,
    refresh_due_at TIMESTAMPTZ,

    -- Source of wealth/funds
    sow_summary TEXT, -- Source of Wealth
    sof_summary TEXT, -- Source of Funds

    -- Approval details
    approved_by TEXT,
    approved_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE tax_profile (
    investor_id UUID PRIMARY KEY REFERENCES investor(investor_id),

    -- FATCA classification
    fatca_class TEXT, -- e.g., "Active NFFE", "Passive NFFE"

    -- CRS classification
    crs_class TEXT, -- e.g., "FI", "NFE"

    -- Tax identification
    tin TEXT, -- Tax Identification Number
    withholding_rate NUMERIC(5,2) DEFAULT 0,

    -- Form details
    form_type TEXT, -- W-8BEN, W-8BEN-E, W-9, CRS
    form_signed_at DATE,
    form_expires_at DATE,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Banking and Settlement
-- ============================================================================

CREATE TABLE bank_instruction (
    bank_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investor_id UUID NOT NULL REFERENCES investor(investor_id),
    currency TEXT NOT NULL, -- ISO 4217

    -- Account details
    account_name TEXT NOT NULL,
    iban TEXT,
    account_no TEXT,
    swift_bic TEXT NOT NULL,

    -- Intermediary bank for correspondent banking
    intermediary_bank TEXT,
    intermediary_swift TEXT,

    -- Validity period
    active_from DATE NOT NULL DEFAULT CURRENT_DATE,
    active_to DATE,

    created_at TIMESTAMPTZ DEFAULT now(),

    -- Only one active instruction per currency per investor
    CONSTRAINT unique_active_bank_per_currency
        EXCLUDE (investor_id WITH =, currency WITH =)
        WHERE (active_to IS NULL)
);

-- ============================================================================
-- Trading and Allocations
-- ============================================================================

CREATE TABLE trade (
    trade_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investor_id UUID NOT NULL REFERENCES investor(investor_id),
    class_id UUID NOT NULL REFERENCES share_class(class_id),
    series_id UUID REFERENCES series(series_id),

    -- Trade details
    type TEXT NOT NULL CHECK (type IN ('SUB', 'RED', 'TRANSFER_IN', 'TRANSFER_OUT', 'SWITCH_IN', 'SWITCH_OUT')),
    status TEXT NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'ALLOCATED', 'SETTLED', 'CANCELLED')),

    -- Dates
    trade_date DATE NOT NULL,
    nav_date DATE,
    settlement_date DATE,

    -- Financial details
    nav_per_share NUMERIC(18,8),
    units NUMERIC(28,10),
    gross_amount NUMERIC(28,2),
    fees_amount NUMERIC(28,2) DEFAULT 0,
    net_amount NUMERIC(28,2) GENERATED ALWAYS AS (gross_amount - fees_amount) STORED,

    -- Currency and FX
    currency TEXT NOT NULL,
    fx_rate NUMERIC(18,8) DEFAULT 1,

    -- Settlement reference
    bank_id UUID REFERENCES bank_instruction(bank_id),

    -- Idempotency and audit
    idempotency_key TEXT UNIQUE,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Event Sourcing: Register Lots and Events
-- ============================================================================

-- Current holdings (derived from events)
CREATE TABLE register_lot (
    lot_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investor_id UUID NOT NULL REFERENCES investor(investor_id),
    class_id UUID NOT NULL REFERENCES share_class(class_id),
    series_id UUID REFERENCES series(series_id),

    -- Current position
    units NUMERIC(28,10) NOT NULL DEFAULT 0,

    -- Historical tracking
    first_trade_date DATE NOT NULL,
    last_activity_at TIMESTAMPTZ DEFAULT now(),

    created_at TIMESTAMPTZ DEFAULT now(),

    -- One lot per investor/class/series combination
    UNIQUE(investor_id, class_id, COALESCE(series_id, '00000000-0000-0000-0000-000000000000'::UUID))
);

-- Immutable event log for all unit movements
CREATE TABLE register_event (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lot_id UUID NOT NULL REFERENCES register_lot(lot_id),

    -- Event classification
    event_type TEXT NOT NULL CHECK (event_type IN ('ISSUE', 'REDEEM', 'TRANSFER_IN', 'TRANSFER_OUT', 'CORP_ACTION')),
    event_ts TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Unit movement (positive = issue/transfer in, negative = redeem/transfer out)
    delta_units NUMERIC(28,10) NOT NULL,

    -- Pricing information
    nav_per_share NUMERIC(18,8),
    nav_date DATE,

    -- Reference to source transaction
    source_trade_id UUID REFERENCES trade(trade_id),

    -- Additional context
    note TEXT,
    reference_id TEXT, -- External reference

    created_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Lifecycle State Management
-- ============================================================================

CREATE TABLE lifecycle_state (
    entity_type TEXT NOT NULL, -- 'INVESTOR', 'TRADE', etc.
    entity_id UUID NOT NULL,
    state TEXT NOT NULL CHECK (state IN (
        'OPPORTUNITY', 'PRECHECKS', 'KYC_PENDING', 'KYC_APPROVED',
        'SUB_PENDING_CASH', 'FUNDED_PENDING_NAV', 'ISSUED', 'ACTIVE',
        'REDEEM_PENDING', 'REDEEMED', 'OFFBOARDED'
    )),
    effective_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    data JSONB DEFAULT '{}', -- Additional state-specific data

    PRIMARY KEY (entity_type, entity_id)
);

-- State transition history
CREATE TABLE lifecycle_state_journal (
    journal_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    from_state TEXT,
    to_state TEXT NOT NULL,
    transitioned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    triggered_by TEXT, -- DSL verb that caused transition
    data JSONB DEFAULT '{}',

    FOREIGN KEY (entity_type, entity_id)
        REFERENCES lifecycle_state(entity_type, entity_id)
);

-- ============================================================================
-- Document Management
-- ============================================================================

CREATE TABLE document (
    doc_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject_type TEXT NOT NULL, -- 'INVESTOR', 'BENEFICIAL_OWNER'
    subject_id UUID NOT NULL,

    -- Document classification
    doc_type TEXT NOT NULL, -- 'passport', 'cert_of_incorporation', 'W-8BEN-E', etc.
    doc_subtype TEXT, -- Additional classification

    -- Storage and integrity
    uri TEXT, -- Storage location
    sha256 TEXT, -- Content hash for integrity

    -- Validity
    issued_on DATE,
    expires_on DATE,

    -- Verification
    verified BOOLEAN DEFAULT FALSE,
    verified_by TEXT,
    verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Audit and Compliance Logging
-- ============================================================================

CREATE TABLE audit_event (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor TEXT NOT NULL, -- User or system identifier
    action TEXT NOT NULL, -- DSL verb or operation
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Event payload and context
    data JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,

    -- Compliance flags
    regulatory_event BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT now()
);

-- ============================================================================
-- Indexes for Performance
-- ============================================================================

-- Investor searches
CREATE INDEX idx_investor_name_trgm ON investor USING gin (legal_name gin_trgm_ops);
CREATE INDEX idx_investor_lei ON investor(lei) WHERE lei IS NOT NULL;
CREATE INDEX idx_investor_status ON investor(status);
CREATE INDEX idx_investor_type ON investor(type);

-- Trading and positions
CREATE INDEX idx_trade_investor_date ON trade(investor_id, trade_date DESC);
CREATE INDEX idx_trade_class_date ON trade(class_id, trade_date DESC);
CREATE INDEX idx_trade_status ON trade(status);
CREATE INDEX idx_trade_type ON trade(type);

-- Register lots and events
CREATE INDEX idx_register_lot_investor ON register_lot(investor_id);
CREATE INDEX idx_register_lot_class ON register_lot(class_id);
CREATE INDEX idx_register_event_lot_time ON register_event(lot_id, event_ts DESC);
CREATE INDEX idx_register_event_type ON register_event(event_type);
CREATE INDEX idx_register_event_trade ON register_event(source_trade_id) WHERE source_trade_id IS NOT NULL;

-- KYC and compliance
CREATE INDEX idx_kyc_status ON kyc_profile(status);
CREATE INDEX idx_kyc_refresh_due ON kyc_profile(refresh_due_at) WHERE refresh_due_at IS NOT NULL;
CREATE INDEX idx_beneficial_owner_investor ON beneficial_owner(investor_id);

-- Banking
CREATE INDEX idx_bank_instruction_investor_currency ON bank_instruction(investor_id, currency);
CREATE INDEX idx_bank_instruction_active ON bank_instruction(active_to) WHERE active_to IS NULL;

-- Lifecycle and audit
CREATE INDEX idx_lifecycle_state_entity ON lifecycle_state(entity_type, entity_id);
CREATE INDEX idx_lifecycle_journal_entity_time ON lifecycle_state_journal(entity_type, entity_id, transitioned_at DESC);
CREATE INDEX idx_audit_event_entity_time ON audit_event(entity_type, entity_id, at DESC);
CREATE INDEX idx_audit_event_actor_time ON audit_event(actor, at DESC);

-- Documents
CREATE INDEX idx_document_subject ON document(subject_type, subject_id);
CREATE INDEX idx_document_type ON document(doc_type);
CREATE INDEX idx_document_expires ON document(expires_on) WHERE expires_on IS NOT NULL;

-- ============================================================================
-- Views for Register Reporting
-- ============================================================================

-- Canonical Register of Investors view
CREATE VIEW register_of_investors_v AS
SELECT
    i.investor_id,
    i.legal_name,
    i.type,
    i.domicile,
    i.address_line1,
    i.city,
    i.postal_code,
    i.country,
    i.status,

    f.fund_id,
    f.legal_name AS fund_name,

    sc.class_id,
    sc.code AS share_class_code,
    sc.currency,

    s.series_id,
    s.code AS series_code,

    rl.units AS units_held,
    rl.first_trade_date,
    rl.last_activity_at,

    -- Calculate total value at last known NAV (simplified)
    (SELECT nav_per_share
     FROM trade t
     WHERE t.class_id = sc.class_id
       AND t.nav_per_share IS NOT NULL
     ORDER BY t.nav_date DESC
     LIMIT 1) AS last_nav_per_share

FROM register_lot rl
JOIN investor i ON i.investor_id = rl.investor_id
JOIN share_class sc ON sc.class_id = rl.class_id
JOIN fund f ON f.fund_id = sc.fund_id
LEFT JOIN series s ON s.series_id = rl.series_id
WHERE rl.units > 0; -- Only show current holdings

-- KYC status overview
CREATE VIEW kyc_status_v AS
SELECT
    i.investor_id,
    i.legal_name,
    i.type,
    i.status AS investor_status,

    kp.risk_rating,
    kp.status AS kyc_status,
    kp.last_screened_at,
    kp.refresh_due_at,

    -- Days until refresh due
    CASE
        WHEN kp.refresh_due_at IS NOT NULL
        THEN kp.refresh_due_at - CURRENT_DATE
    END AS days_until_refresh,

    -- Document counts
    (SELECT COUNT(*) FROM document d WHERE d.subject_id = i.investor_id AND d.verified = TRUE) AS verified_docs,
    (SELECT COUNT(*) FROM document d WHERE d.subject_id = i.investor_id AND d.verified = FALSE) AS pending_docs,

    -- Beneficial ownership
    (SELECT COUNT(*) FROM beneficial_owner bo WHERE bo.investor_id = i.investor_id AND bo.verified = TRUE) AS verified_bos,
    (SELECT COUNT(*) FROM beneficial_owner bo WHERE bo.investor_id = i.investor_id AND bo.verified = FALSE) AS pending_bos

FROM investor i
LEFT JOIN kyc_profile kp ON kp.investor_id = i.investor_id;

-- Trading activity summary
CREATE VIEW trading_activity_v AS
SELECT
    i.investor_id,
    i.legal_name,

    COUNT(t.trade_id) AS total_trades,
    COUNT(CASE WHEN t.type = 'SUB' THEN 1 END) AS subscriptions,
    COUNT(CASE WHEN t.type = 'RED' THEN 1 END) AS redemptions,

    SUM(CASE WHEN t.type = 'SUB' AND t.status = 'SETTLED' THEN t.gross_amount ELSE 0 END) AS total_subscribed,
    SUM(CASE WHEN t.type = 'RED' AND t.status = 'SETTLED' THEN t.gross_amount ELSE 0 END) AS total_redeemed,

    MIN(t.trade_date) AS first_trade_date,
    MAX(t.trade_date) AS last_trade_date

FROM investor i
LEFT JOIN trade t ON t.investor_id = i.investor_id
GROUP BY i.investor_id, i.legal_name;

-- ============================================================================
-- Triggers for Automation
-- ============================================================================

-- Update investor.updated_at on changes
CREATE OR REPLACE FUNCTION update_investor_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER investor_updated_at
    BEFORE UPDATE ON investor
    FOR EACH ROW
    EXECUTE FUNCTION update_investor_timestamp();

-- Automatically update register_lot when register_event is inserted
CREATE OR REPLACE FUNCTION update_register_lot_from_event()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the lot units
    UPDATE register_lot
    SET
        units = units + NEW.delta_units,
        last_activity_at = NEW.event_ts
    WHERE lot_id = NEW.lot_id;

    -- If units become zero or negative, handle appropriately
    UPDATE register_lot
    SET units = 0
    WHERE lot_id = NEW.lot_id AND units < 0;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER register_event_update_lot
    AFTER INSERT ON register_event
    FOR EACH ROW
    EXECUTE FUNCTION update_register_lot_from_event();

-- Log lifecycle state transitions
CREATE OR REPLACE FUNCTION log_lifecycle_transition()
RETURNS TRIGGER AS $$
BEGIN
    -- Only log if state actually changed
    IF TG_OP = 'UPDATE' AND OLD.state = NEW.state THEN
        RETURN NEW;
    END IF;

    INSERT INTO lifecycle_state_journal (
        entity_type, entity_id, from_state, to_state, data
    ) VALUES (
        NEW.entity_type,
        NEW.entity_id,
        CASE WHEN TG_OP = 'UPDATE' THEN OLD.state ELSE NULL END,
        NEW.state,
        NEW.data
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER lifecycle_state_journal_trigger
    AFTER INSERT OR UPDATE ON lifecycle_state
    FOR EACH ROW
    EXECUTE FUNCTION log_lifecycle_transition();

-- ============================================================================
-- Sample Data Functions
-- ============================================================================

-- Function to create a new fund with default share classes
CREATE OR REPLACE FUNCTION create_sample_fund(
    p_fund_name TEXT,
    p_domicile TEXT DEFAULT 'LU'
) RETURNS UUID AS $$
DECLARE
    v_fund_id UUID;
    v_class_usd_id UUID;
    v_class_eur_id UUID;
BEGIN
    -- Create fund
    INSERT INTO fund (legal_name, domicile, admin_name, inception_date, status)
    VALUES (p_fund_name, p_domicile, 'FundAdmin Corp', CURRENT_DATE, 'ACTIVE')
    RETURNING fund_id INTO v_fund_id;

    -- Create USD share class
    INSERT INTO share_class (fund_id, code, currency, dealing_freq, notice_days, management_fee_bps)
    VALUES (v_fund_id, 'A USD', 'USD', 'Monthly', 30, 200)
    RETURNING class_id INTO v_class_usd_id;

    -- Create EUR share class
    INSERT INTO share_class (fund_id, code, currency, dealing_freq, notice_days, management_fee_bps)
    VALUES (v_fund_id, 'A EUR', 'EUR', 'Monthly', 30, 200)
    RETURNING class_id INTO v_class_eur_id;

    RETURN v_fund_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Comments and Documentation
-- ============================================================================

COMMENT ON SCHEMA register IS 'Hedge Fund Register of Investors - Event Sourced Investor Management';

COMMENT ON TABLE investor IS 'Core investor legal identity and contact information';
COMMENT ON TABLE register_lot IS 'Current unit holdings per investor/class/series - derived from events';
COMMENT ON TABLE register_event IS 'Immutable log of all unit movements - source of truth for positions';
COMMENT ON TABLE trade IS 'Subscription and redemption orders with allocation details';
COMMENT ON TABLE kyc_profile IS 'Know Your Customer profile with risk rating and screening status';
COMMENT ON TABLE lifecycle_state IS 'Current investor lifecycle state for workflow management';

COMMENT ON VIEW register_of_investors_v IS 'Canonical Register of Investors view for regulatory reporting';

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Reset search path
RESET search_path;

-- Grant appropriate permissions (adjust as needed)
-- GRANT USAGE ON SCHEMA register TO investor_app_role;
-- GRANT ALL ON ALL TABLES IN SCHEMA register TO investor_app_role;
-- GRANT ALL ON ALL SEQUENCES IN SCHEMA register TO investor_app_role;