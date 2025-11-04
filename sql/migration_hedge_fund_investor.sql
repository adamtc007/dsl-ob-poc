/*
Migration: Add Hedge Fund Investor Register Schema
- Implements complete hedge fund investor register with event sourcing
- Supports full investor lifecycle from opportunity to offboarding
- Includes KYC, tax, banking, and trade management
- All tables clearly identified as hedge fund investor implementation

KEY QUERIES:

1) Position as-of query (fast projection):
-- Units per investor/fund/class/series as of a date
SELECT
  l.investor_id, l.fund_id, l.class_id, l.series_id,
  SUM(e.delta_units) AS units
FROM "hf-investor".hf_register_events e
JOIN "hf-investor".hf_register_lots l ON l.lot_id = e.lot_id
WHERE e.value_date <= $1::date
GROUP BY 1,2,3,4;

2) Pipeline funnel (ops dashboard):
-- Count investors by status for operational dashboard
SELECT status, COUNT(*) AS investors
FROM "hf-investor".hf_investors
GROUP BY status
ORDER BY status;

3) Outstanding KYC requirements:
-- Track pending and overdue document requirements
SELECT investor_id, doc_type, status, requested_at, due_at
FROM "hf-investor".hf_document_requirements
WHERE status IN ('REQUESTED','OVERDUE')
ORDER BY due_at NULLS LAST;

*/

-- Create hedge fund investor schema namespace
CREATE SCHEMA IF NOT EXISTS "hf-investor";

-- ============================================================================
-- FUND STRUCTURE TABLES (HEDGE FUND INVESTOR)
-- ============================================================================

-- Hedge Fund definition table
CREATE TABLE IF NOT EXISTS "hf-investor".hf_funds (
    fund_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fund_name VARCHAR(255) NOT NULL UNIQUE,
    legal_name VARCHAR(500) NOT NULL,
    lei VARCHAR(20) UNIQUE,
    domicile VARCHAR(5) NOT NULL,
    fund_type VARCHAR(50) NOT NULL CHECK (fund_type IN ('HEDGE_FUND', 'PRIVATE_EQUITY', 'CREDIT_FUND', 'INFRASTRUCTURE')),
    currency VARCHAR(3) NOT NULL,
    inception_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('SETUP', 'ACTIVE', 'SOFT_CLOSED', 'HARD_CLOSED', 'LIQUIDATING', 'LIQUIDATED')),
    administrator VARCHAR(255),
    custodian VARCHAR(255),
    auditor VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_funds_name ON "hf-investor".hf_funds (fund_name);
CREATE INDEX IF NOT EXISTS idx_hf_funds_status ON "hf-investor".hf_funds (status);
CREATE INDEX IF NOT EXISTS idx_hf_funds_domicile ON "hf-investor".hf_funds (domicile);

-- Hedge Fund Share Classes
CREATE TABLE IF NOT EXISTS "hf-investor".hf_share_classes (
    class_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fund_id UUID NOT NULL REFERENCES "hf-investor".hf_funds (fund_id) ON DELETE CASCADE,
    class_name VARCHAR(10) NOT NULL, -- 'A', 'B', 'C', 'I', etc.
    class_type VARCHAR(50) NOT NULL CHECK (class_type IN ('RETAIL', 'INSTITUTIONAL', 'FOUNDER', 'EMPLOYEE', 'SEEDED')),
    currency VARCHAR(3) NOT NULL,
    min_initial_investment NUMERIC(28,2) NOT NULL,
    min_subsequent_investment NUMERIC(28,2) NOT NULL,
    management_fee_rate NUMERIC(8,6) NOT NULL DEFAULT 0,
    performance_fee_rate NUMERIC(8,6) NOT NULL DEFAULT 0,
    high_water_mark BOOLEAN NOT NULL DEFAULT true,
    dealing_frequency VARCHAR(20) NOT NULL CHECK (dealing_frequency IN ('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'SEMI_ANNUAL', 'ANNUAL')),
    notice_period_days INTEGER NOT NULL DEFAULT 90,
    gate_percentage NUMERIC(5,2), -- Optional gate (e.g., 25.00 for 25%)
    lockup_months INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (fund_id, class_name)
);

CREATE INDEX IF NOT EXISTS idx_hf_share_classes_fund ON "hf-investor".hf_share_classes (fund_id);
CREATE INDEX IF NOT EXISTS idx_hf_share_classes_type ON "hf-investor".hf_share_classes (class_type);

-- Hedge Fund Series (for equalisation)
CREATE TABLE IF NOT EXISTS "hf-investor".hf_series (
    series_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id UUID NOT NULL REFERENCES "hf-investor".hf_share_classes (class_id) ON DELETE CASCADE,
    series_name VARCHAR(50) NOT NULL,
    inception_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CLOSED', 'MERGED')),
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (class_id, series_name)
);

CREATE INDEX IF NOT EXISTS idx_hf_series_class ON "hf-investor".hf_series (class_id);
CREATE INDEX IF NOT EXISTS idx_hf_series_status ON "hf-investor".hf_series (status);

-- ============================================================================
-- INVESTOR IDENTITY TABLES (HEDGE FUND INVESTOR)
-- ============================================================================

-- Hedge Fund Investors
CREATE TABLE IF NOT EXISTS "hf-investor".hf_investors (
    investor_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_code VARCHAR(50) NOT NULL UNIQUE, -- Human-readable ID like 'INV-001'
    type VARCHAR(20) NOT NULL CHECK (type IN ('INDIVIDUAL', 'CORPORATE', 'TRUST', 'FOHF', 'NOMINEE', 'PENSION_FUND', 'INSURANCE_CO')),
    legal_name VARCHAR(500) NOT NULL,
    short_name VARCHAR(100),
    lei VARCHAR(20),
    registration_number VARCHAR(100),
    domicile VARCHAR(5) NOT NULL,

    -- Address information
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    address_line3 VARCHAR(255),
    address_line4 VARCHAR(255),
    city VARCHAR(100),
    state_province VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(5),

    -- Contact information
    primary_contact_name VARCHAR(255),
    primary_contact_email VARCHAR(255),
    primary_contact_phone VARCHAR(50),

    -- Status and lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'OPPORTUNITY' CHECK (status IN ('OPPORTUNITY', 'PRECHECKS', 'KYC_PENDING', 'KYC_APPROVED', 'SUB_PENDING_CASH', 'FUNDED_PENDING_NAV', 'ISSUED', 'ACTIVE', 'REDEEM_PENDING', 'REDEEMED', 'OFFBOARDED')),
    source VARCHAR(100), -- How investor was sourced

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_investors_code ON "hf-investor".hf_investors (investor_code);
CREATE INDEX IF NOT EXISTS idx_hf_investors_status ON "hf-investor".hf_investors (status);
CREATE INDEX IF NOT EXISTS idx_hf_investors_type ON "hf-investor".hf_investors (type);
CREATE INDEX IF NOT EXISTS idx_hf_investors_domicile ON "hf-investor".hf_investors (domicile);
CREATE INDEX IF NOT EXISTS idx_hf_investors_legal_name ON "hf-investor".hf_investors (legal_name);

-- Hedge Fund Beneficial Owners
CREATE TABLE IF NOT EXISTS "hf-investor".hf_beneficial_owners (
    bo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    date_of_birth DATE,
    nationality VARCHAR(5),
    ownership_percentage NUMERIC(5,2) NOT NULL CHECK (ownership_percentage >= 0 AND ownership_percentage <= 100),
    control_type VARCHAR(50) CHECK (control_type IN ('OWNERSHIP', 'VOTING', 'CONTROL', 'SENIOR_MANAGING_OFFICIAL')),

    -- Risk flags
    is_pep BOOLEAN NOT NULL DEFAULT false,
    pep_details TEXT,
    sanctions_flag BOOLEAN NOT NULL DEFAULT false,
    sanctions_details TEXT,

    -- Address
    address_line1 VARCHAR(255),
    city VARCHAR(100),
    country VARCHAR(5),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_beneficial_owners_investor ON "hf-investor".hf_beneficial_owners (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_beneficial_owners_pep ON "hf-investor".hf_beneficial_owners (is_pep);
CREATE INDEX IF NOT EXISTS idx_hf_beneficial_owners_sanctions ON "hf-investor".hf_beneficial_owners (sanctions_flag);

-- ============================================================================
-- KYC AND COMPLIANCE TABLES (HEDGE FUND INVESTOR)
-- ============================================================================

-- Hedge Fund KYC Profiles
CREATE TABLE IF NOT EXISTS "hf-investor".hf_kyc_profiles (
    kyc_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    risk_rating VARCHAR(10) NOT NULL CHECK (risk_rating IN ('LOW', 'MEDIUM', 'HIGH', 'PROHIBITED')),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'EXPIRED')),

    -- KYC details
    kyc_tier VARCHAR(20) CHECK (kyc_tier IN ('SIMPLIFIED', 'STANDARD', 'ENHANCED')),
    screening_provider VARCHAR(100),
    screening_reference VARCHAR(255),
    screening_date DATE,
    screening_result VARCHAR(20) CHECK (screening_result IN ('CLEAR', 'POTENTIAL_MATCH', 'TRUE_POSITIVE')),

    -- Approval details
    approved_by VARCHAR(255),
    approved_at TIMESTAMPTZ,
    approval_comments TEXT,

    -- Refresh schedule
    refresh_frequency VARCHAR(20) DEFAULT 'ANNUAL' CHECK (refresh_frequency IN ('MONTHLY', 'QUARTERLY', 'SEMI_ANNUAL', 'ANNUAL', 'BIENNIAL')),
    refresh_due_at TIMESTAMPTZ,
    last_refreshed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (investor_id)
);

CREATE INDEX IF NOT EXISTS idx_hf_kyc_profiles_investor ON "hf-investor".hf_kyc_profiles (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_kyc_profiles_status ON "hf-investor".hf_kyc_profiles (status);
CREATE INDEX IF NOT EXISTS idx_hf_kyc_profiles_risk ON "hf-investor".hf_kyc_profiles (risk_rating);
CREATE INDEX IF NOT EXISTS idx_hf_kyc_profiles_refresh_due ON "hf-investor".hf_kyc_profiles (refresh_due_at);

-- Hedge Fund Tax Profiles
CREATE TABLE IF NOT EXISTS "hf-investor".hf_tax_profiles (
    tax_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,

    -- FATCA classification
    fatca_status VARCHAR(50) CHECK (fatca_status IN ('US_PERSON', 'NON_US_PERSON', 'SPECIFIED_US_PERSON', 'EXEMPT_BENEFICIAL_OWNER')),
    fatca_giin VARCHAR(19), -- Global Intermediary Identification Number

    -- CRS classification
    crs_classification VARCHAR(50) CHECK (crs_classification IN ('INDIVIDUAL', 'ENTITY', 'FINANCIAL_INSTITUTION', 'INVESTMENT_ENTITY')),
    crs_jurisdiction VARCHAR(5),

    -- Tax forms and documentation
    form_type VARCHAR(50) CHECK (form_type IN ('W9', 'W8_BEN', 'W8_BEN_E', 'W8_ECI', 'W8_EXP', 'W8_IMY', 'ENTITY_SELF_CERT')),
    form_date DATE,
    form_valid_until DATE,

    -- Tax rates
    withholding_rate NUMERIC(5,4) DEFAULT 0, -- e.g., 0.30 for 30%
    backup_withholding BOOLEAN DEFAULT false,

    -- TIN information
    tin_type VARCHAR(20) CHECK (tin_type IN ('SSN', 'ITIN', 'EIN', 'FOREIGN_TIN', 'GIIN')),
    tin_value VARCHAR(50),
    tin_jurisdiction VARCHAR(5),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (investor_id)
);

CREATE INDEX IF NOT EXISTS idx_hf_tax_profiles_investor ON "hf-investor".hf_tax_profiles (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_tax_profiles_fatca ON "hf-investor".hf_tax_profiles (fatca_status);
CREATE INDEX IF NOT EXISTS idx_hf_tax_profiles_crs ON "hf-investor".hf_tax_profiles (crs_classification);
CREATE INDEX IF NOT EXISTS idx_hf_tax_profiles_form_expiry ON "hf-investor".hf_tax_profiles (form_valid_until);

-- ============================================================================
-- BANKING AND SETTLEMENT TABLES (HEDGE FUND INVESTOR)
-- ============================================================================

-- Hedge Fund Bank Instructions
CREATE TABLE IF NOT EXISTS "hf-investor".hf_bank_instructions (
    bank_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    currency VARCHAR(3) NOT NULL,
    instruction_type VARCHAR(20) NOT NULL DEFAULT 'SETTLEMENT' CHECK (instruction_type IN ('SETTLEMENT', 'FEE_PAYMENT', 'DISTRIBUTION')),

    -- Bank details
    bank_name VARCHAR(255) NOT NULL,
    swift_bic VARCHAR(11),
    iban VARCHAR(34),
    account_number VARCHAR(50),
    account_name VARCHAR(255) NOT NULL,

    -- Address
    bank_address_line1 VARCHAR(255),
    bank_address_line2 VARCHAR(255),
    bank_city VARCHAR(100),
    bank_country VARCHAR(5),

    -- Intermediary bank (for USD/correspondent banking)
    intermediary_swift VARCHAR(11),
    intermediary_name VARCHAR(255),
    intermediary_account VARCHAR(50),

    -- Status and versioning
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUPERSEDED')),
    version_number INTEGER NOT NULL DEFAULT 1,
    superseded_by UUID REFERENCES "hf-investor".hf_bank_instructions (bank_id),

    -- Verification
    verified_by VARCHAR(255),
    verified_at TIMESTAMPTZ,
    verification_method VARCHAR(50),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_bank_instructions_investor ON "hf-investor".hf_bank_instructions (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_bank_instructions_currency ON "hf-investor".hf_bank_instructions (currency);
CREATE INDEX IF NOT EXISTS idx_hf_bank_instructions_status ON "hf-investor".hf_bank_instructions (status);
CREATE INDEX IF NOT EXISTS idx_hf_bank_instructions_type ON "hf-investor".hf_bank_instructions (instruction_type);

-- ============================================================================
-- TRADING AND REGISTER TABLES (HEDGE FUND INVESTOR)
-- ============================================================================

-- Hedge Fund Trades (Orders and Allocations)
CREATE TABLE IF NOT EXISTS "hf-investor".hf_trades (
    trade_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_reference VARCHAR(100) NOT NULL UNIQUE, -- Human-readable reference
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    fund_id UUID NOT NULL REFERENCES "hf-investor".hf_funds (fund_id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES "hf-investor".hf_share_classes (class_id) ON DELETE CASCADE,
    series_id UUID REFERENCES "hf-investor".hf_series (series_id) ON DELETE CASCADE,

    -- Trade type and status
    trade_type VARCHAR(20) NOT NULL CHECK (trade_type IN ('SUB', 'RED', 'TRANSFER_IN', 'TRANSFER_OUT', 'SWITCH_IN', 'SWITCH_OUT', 'CORP_ACTION')),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ALLOCATED', 'SETTLED', 'CANCELLED', 'REJECTED')),

    -- Trade details
    trade_date DATE NOT NULL,
    value_date DATE NOT NULL,
    settlement_date DATE,
    nav_date DATE,

    -- Amounts and pricing
    requested_amount NUMERIC(28,2), -- Amount investor requested
    gross_amount NUMERIC(28,2), -- Actual gross amount
    fees_amount NUMERIC(28,2) DEFAULT 0,
    net_amount NUMERIC(28,2), -- Net amount after fees
    nav_per_share NUMERIC(18,8),
    units NUMERIC(28,10),

    -- Currency and FX
    currency VARCHAR(3) NOT NULL,
    fx_rate NUMERIC(18,8) DEFAULT 1,
    base_currency_amount NUMERIC(28,2),

    -- Settlement
    bank_id UUID REFERENCES "hf-investor".hf_bank_instructions (bank_id),
    settlement_reference VARCHAR(255),

    -- Workflow
    idempotency_key VARCHAR(255) UNIQUE, -- Prevents duplicate trades
    notice_date DATE, -- For redemptions
    comments TEXT,

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_trades_investor ON "hf-investor".hf_trades (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_trades_fund_class ON "hf-investor".hf_trades (fund_id, class_id);
CREATE INDEX IF NOT EXISTS idx_hf_trades_type_status ON "hf-investor".hf_trades (trade_type, status);
CREATE INDEX IF NOT EXISTS idx_hf_trades_dates ON "hf-investor".hf_trades (trade_date, value_date);
CREATE INDEX IF NOT EXISTS idx_hf_trades_reference ON "hf-investor".hf_trades (trade_reference);
CREATE INDEX IF NOT EXISTS idx_hf_trades_idempotency ON "hf-investor".hf_trades (idempotency_key);

-- Hedge Fund Register Lots (Current Holdings)
CREATE TABLE IF NOT EXISTS "hf-investor".hf_register_lots (
    lot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    fund_id UUID NOT NULL REFERENCES "hf-investor".hf_funds (fund_id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES "hf-investor".hf_share_classes (class_id) ON DELETE CASCADE,
    series_id UUID REFERENCES "hf-investor".hf_series (series_id) ON DELETE CASCADE,

    -- Current position
    units NUMERIC(28,10) NOT NULL DEFAULT 0,
    average_cost NUMERIC(18,8),
    total_cost NUMERIC(28,2) DEFAULT 0,

    -- Dates
    first_trade_date DATE,
    last_activity_at TIMESTAMPTZ,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CLOSED')),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    UNIQUE (investor_id, fund_id, class_id, series_id)
);

CREATE INDEX IF NOT EXISTS idx_hf_register_lots_investor ON "hf-investor".hf_register_lots (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_register_lots_fund_class ON "hf-investor".hf_register_lots (fund_id, class_id);
CREATE INDEX IF NOT EXISTS idx_hf_register_lots_status ON "hf-investor".hf_register_lots (status);
CREATE INDEX IF NOT EXISTS idx_hf_register_lots_last_activity ON "hf-investor".hf_register_lots (last_activity_at);

-- Hedge Fund Register Events (Immutable Event Sourcing)
CREATE TABLE IF NOT EXISTS "hf-investor".hf_register_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id UUID NOT NULL REFERENCES "hf-investor".hf_register_lots (lot_id) ON DELETE CASCADE,
    trade_id UUID REFERENCES "hf-investor".hf_trades (trade_id) ON DELETE CASCADE,

    -- Event details
    event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('ISSUE', 'REDEEM', 'TRANSFER_IN', 'TRANSFER_OUT', 'CORP_ACTION', 'FEE_CHARGE', 'DIVIDEND')),
    event_timestamp TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc'),
    value_date DATE NOT NULL,

    -- Unit movement
    delta_units NUMERIC(28,10) NOT NULL, -- Positive for additions, negative for reductions
    running_balance NUMERIC(28,10) NOT NULL, -- Units after this event

    -- Pricing
    nav_per_share NUMERIC(18,8),
    price_per_share NUMERIC(18,8), -- Actual trade price (may differ from NAV due to fees)

    -- Amounts
    gross_amount NUMERIC(28,2),
    fees_amount NUMERIC(28,2) DEFAULT 0,
    net_amount NUMERIC(28,2),

    -- Metadata
    description TEXT,
    external_reference VARCHAR(255),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_register_events_lot ON "hf-investor".hf_register_events (lot_id);
CREATE INDEX IF NOT EXISTS idx_hf_register_events_trade ON "hf-investor".hf_register_events (trade_id);
CREATE INDEX IF NOT EXISTS idx_hf_register_events_type ON "hf-investor".hf_register_events (event_type);
CREATE INDEX IF NOT EXISTS idx_hf_register_events_timestamp ON "hf-investor".hf_register_events (event_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_hf_register_events_value_date ON "hf-investor".hf_register_events (value_date DESC);

-- ============================================================================
-- LIFECYCLE AND DOCUMENT MANAGEMENT TABLES (HEDGE FUND INVESTOR)
-- ============================================================================

-- Hedge Fund Lifecycle States
CREATE TABLE IF NOT EXISTS "hf-investor".hf_lifecycle_states (
    state_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    from_state VARCHAR(30),
    to_state VARCHAR(30) NOT NULL,
    transition_trigger VARCHAR(100), -- DSL verb that triggered the transition
    guard_conditions JSONB, -- Conditions that were checked
    metadata JSONB, -- Additional context
    transitioned_by VARCHAR(255),
    transitioned_at TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_lifecycle_states_investor ON "hf-investor".hf_lifecycle_states (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_lifecycle_states_to_state ON "hf-investor".hf_lifecycle_states (to_state);
CREATE INDEX IF NOT EXISTS idx_hf_lifecycle_states_timestamp ON "hf-investor".hf_lifecycle_states (transitioned_at DESC);

-- Hedge Fund Documents
CREATE TABLE IF NOT EXISTS "hf-investor".hf_documents (
    document_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    document_type VARCHAR(100) NOT NULL, -- 'passport', 'articles_of_association', 'board_resolution', etc.
    document_subject VARCHAR(100), -- 'primary_signatory', 'authorized_signatory', 'beneficial_owner_1', etc.
    document_title VARCHAR(255) NOT NULL,

    -- Document metadata
    file_name VARCHAR(255),
    file_size BIGINT,
    mime_type VARCHAR(100),
    file_hash VARCHAR(64), -- SHA-256 hash for integrity

    -- Storage location
    storage_provider VARCHAR(50), -- 'S3', 'AZURE', 'LOCAL', etc.
    storage_path TEXT,

    -- Lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'RECEIVED' CHECK (status IN ('RECEIVED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'EXPIRED', 'SUPERSEDED')),
    reviewed_by VARCHAR(255),
    reviewed_at TIMESTAMPTZ,
    review_comments TEXT,

    -- Expiry and refresh
    issued_date DATE,
    expiry_date DATE,
    superseded_by UUID REFERENCES "hf-investor".hf_documents (document_id),

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_documents_investor ON "hf-investor".hf_documents (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_documents_type ON "hf-investor".hf_documents (document_type);
CREATE INDEX IF NOT EXISTS idx_hf_documents_status ON "hf-investor".hf_documents (status);
CREATE INDEX IF NOT EXISTS idx_hf_documents_expiry ON "hf-investor".hf_documents (expiry_date);

-- Hedge Fund Document Requirements (Links requirements to fulfillment)
CREATE TABLE IF NOT EXISTS "hf-investor".hf_document_requirements (
    requirement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    doc_type VARCHAR(100) NOT NULL, -- Required document type
    status VARCHAR(20) NOT NULL DEFAULT 'REQUESTED' CHECK (status IN ('REQUESTED', 'SUBMITTED', 'APPROVED', 'REJECTED', 'OVERDUE', 'WAIVED')),
    priority VARCHAR(20) DEFAULT 'MEDIUM' CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),

    -- Timing
    requested_at TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc'),
    due_at TIMESTAMPTZ, -- When document is required by
    fulfilled_at TIMESTAMPTZ, -- When requirement was satisfied

    -- Linkage to actual documents
    fulfilling_document_id UUID REFERENCES "hf-investor".hf_documents (document_id) ON DELETE SET NULL,

    -- Context
    source VARCHAR(100), -- Who/what triggered this requirement (KYC_ONBOARD, PERIODIC_REFRESH, REGULATORY_CHANGE, etc.)
    requirements_notes TEXT,
    waiver_reason TEXT, -- If waived, why?
    waived_by VARCHAR(255), -- Who waived it?
    waived_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc'),

    -- Ensure one active requirement per investor/doc_type combination
    UNIQUE (investor_id, doc_type, status) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS idx_hf_doc_requirements_investor ON "hf-investor".hf_document_requirements (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_doc_requirements_status ON "hf-investor".hf_document_requirements (status);
CREATE INDEX IF NOT EXISTS idx_hf_doc_requirements_due ON "hf-investor".hf_document_requirements (due_at);
CREATE INDEX IF NOT EXISTS idx_hf_doc_requirements_type ON "hf-investor".hf_document_requirements (doc_type);
CREATE INDEX IF NOT EXISTS idx_hf_doc_requirements_overdue ON "hf-investor".hf_document_requirements (due_at) WHERE status IN ('REQUESTED', 'OVERDUE') AND due_at < now();

-- Hedge Fund Audit Events
CREATE TABLE IF NOT EXISTS "hf-investor".hf_audit_events (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL, -- 'investor', 'trade', 'document', etc.
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'CREATE', 'UPDATE', 'DELETE', 'APPROVE', etc.
    details JSONB, -- Structured details of what changed
    user_id VARCHAR(255) NOT NULL,
    user_ip INET,
    user_agent TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_audit_events_investor ON "hf-investor".hf_audit_events (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_audit_events_entity ON "hf-investor".hf_audit_events (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_hf_audit_events_action ON "hf-investor".hf_audit_events (action);
CREATE INDEX IF NOT EXISTS idx_hf_audit_events_timestamp ON "hf-investor".hf_audit_events (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_hf_audit_events_user ON "hf-investor".hf_audit_events (user_id);

-- ============================================================================
-- HEDGE FUND INVESTOR DSL INTEGRATION TABLES
-- ============================================================================

-- Hedge Fund DSL Execution Log (extends main dsl_ob table concept)
CREATE TABLE IF NOT EXISTS "hf-investor".hf_dsl_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id UUID NOT NULL REFERENCES "hf-investor".hf_investors (investor_id) ON DELETE CASCADE,
    dsl_text TEXT NOT NULL,
    execution_status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (execution_status IN ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    idempotency_key VARCHAR(255) UNIQUE,

    -- Execution context
    triggered_by VARCHAR(255), -- User or system that initiated
    execution_engine VARCHAR(50) DEFAULT 'hedge-fund-dsl-v1',

    -- Results
    affected_entities JSONB, -- List of entities that were modified
    error_details TEXT,
    execution_time_ms INTEGER,

    -- Timing
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT (now() at time zone 'utc')
);

CREATE INDEX IF NOT EXISTS idx_hf_dsl_executions_investor ON "hf-investor".hf_dsl_executions (investor_id);
CREATE INDEX IF NOT EXISTS idx_hf_dsl_executions_status ON "hf-investor".hf_dsl_executions (execution_status);
CREATE INDEX IF NOT EXISTS idx_hf_dsl_executions_idempotency ON "hf-investor".hf_dsl_executions (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_hf_dsl_executions_created ON "hf-investor".hf_dsl_executions (created_at DESC);

-- ============================================================================
-- HEDGE FUND INVESTOR VIEWS AND REPORTING
-- ============================================================================

-- Complete Register of Investors View
CREATE OR REPLACE VIEW "hf-investor".register_of_investors_v AS
SELECT
    i.investor_id,
    i.investor_code,
    i.legal_name as investor_name,
    i.type as investor_type,
    i.domicile,
    i.status as investor_status,

    -- Fund and class details
    f.fund_name,
    f.fund_id,
    sc.class_name,
    sc.currency as class_currency,
    s.series_name,

    -- Current holdings
    rl.units as current_units,
    rl.total_cost as total_investment,
    rl.average_cost,
    rl.first_trade_date,
    rl.last_activity_at,

    -- KYC status
    kyc.risk_rating,
    kyc.status as kyc_status,
    kyc.refresh_due_at,

    -- Tax status
    tax.fatca_status,
    tax.crs_classification,
    tax.withholding_rate,

    -- Contact info
    i.primary_contact_name,
    i.primary_contact_email,

    -- Timestamps
    i.created_at as investor_created_at,
    rl.updated_at as position_updated_at

FROM "hf-investor".hf_investors i
LEFT JOIN "hf-investor".hf_register_lots rl ON i.investor_id = rl.investor_id
LEFT JOIN "hf-investor".hf_funds f ON rl.fund_id = f.fund_id
LEFT JOIN "hf-investor".hf_share_classes sc ON rl.class_id = sc.class_id
LEFT JOIN "hf-investor".hf_series s ON rl.series_id = s.series_id
LEFT JOIN "hf-investor".hf_kyc_profiles kyc ON i.investor_id = kyc.investor_id
LEFT JOIN "hf-investor".hf_tax_profiles tax ON i.investor_id = tax.investor_id
WHERE rl.status = 'ACTIVE' OR rl.status IS NULL
ORDER BY i.investor_code, f.fund_name, sc.class_name;

-- Fund Summary View
CREATE OR REPLACE VIEW "hf-investor".fund_summary_v AS
SELECT
    f.fund_id,
    f.fund_name,
    f.status as fund_status,
    sc.class_id,
    sc.class_name,
    sc.currency,

    -- Aggregate statistics
    COUNT(DISTINCT rl.investor_id) as total_investors,
    COUNT(DISTINCT CASE WHEN i.status = 'ACTIVE' THEN rl.investor_id END) as active_investors,
    SUM(rl.units) as total_units_outstanding,
    SUM(rl.total_cost) as total_aum,

    -- Last activity
    MAX(rl.last_activity_at) as last_activity_date,

    -- Fund details
    f.inception_date,
    f.created_at as fund_created_at

FROM "hf-investor".hf_funds f
LEFT JOIN "hf-investor".hf_share_classes sc ON f.fund_id = sc.fund_id
LEFT JOIN "hf-investor".hf_register_lots rl ON sc.class_id = rl.class_id AND rl.status = 'ACTIVE'
LEFT JOIN "hf-investor".hf_investors i ON rl.investor_id = i.investor_id
GROUP BY f.fund_id, f.fund_name, f.status, sc.class_id, sc.class_name, sc.currency, f.inception_date, f.created_at
ORDER BY f.fund_name, sc.class_name;

-- KYC Dashboard View
CREATE OR REPLACE VIEW "hf-investor".kyc_dashboard_v AS
SELECT
    i.investor_id,
    i.investor_code,
    i.legal_name,
    i.status as investor_status,
    kyc.risk_rating,
    kyc.status as kyc_status,
    kyc.refresh_due_at,
    kyc.last_refreshed_at,

    -- Document counts
    COALESCE(doc_summary.total_docs, 0) as total_documents,
    COALESCE(doc_summary.approved_docs, 0) as approved_documents,
    COALESCE(doc_summary.pending_docs, 0) as pending_documents,
    COALESCE(doc_summary.expired_docs, 0) as expired_documents,

    -- Risk flags
    CASE WHEN EXISTS (SELECT 1 FROM "hf-investor".hf_beneficial_owners bo WHERE bo.investor_id = i.investor_id AND bo.is_pep = true) THEN true ELSE false END as has_pep,
    CASE WHEN EXISTS (SELECT 1 FROM "hf-investor".hf_beneficial_owners bo WHERE bo.investor_id = i.investor_id AND bo.sanctions_flag = true) THEN true ELSE false END as has_sanctions_flag,

    -- Next actions
    CASE
        WHEN kyc.refresh_due_at < CURRENT_DATE THEN 'REFRESH_OVERDUE'
        WHEN kyc.refresh_due_at <= CURRENT_DATE + INTERVAL '30 days' THEN 'REFRESH_DUE'
        WHEN kyc.status = 'PENDING' THEN 'INITIAL_REVIEW'
        WHEN doc_summary.pending_docs > 0 THEN 'PENDING_DOCUMENTS'
        ELSE 'NO_ACTION'
    END as next_action

FROM "hf-investor".hf_investors i
LEFT JOIN "hf-investor".hf_kyc_profiles kyc ON i.investor_id = kyc.investor_id
LEFT JOIN (
    SELECT
        investor_id,
        COUNT(*) as total_docs,
        COUNT(CASE WHEN status = 'APPROVED' THEN 1 END) as approved_docs,
        COUNT(CASE WHEN status IN ('RECEIVED', 'UNDER_REVIEW') THEN 1 END) as pending_docs,
        COUNT(CASE WHEN status = 'EXPIRED' OR (expiry_date IS NOT NULL AND expiry_date < CURRENT_DATE) THEN 1 END) as expired_docs
    FROM "hf-investor".hf_documents
    GROUP BY investor_id
) doc_summary ON i.investor_id = doc_summary.investor_id
ORDER BY
    CASE next_action
        WHEN 'REFRESH_OVERDUE' THEN 1
        WHEN 'REFRESH_DUE' THEN 2
        WHEN 'INITIAL_REVIEW' THEN 3
        WHEN 'PENDING_DOCUMENTS' THEN 4
        ELSE 5
    END,
    kyc.refresh_due_at ASC NULLS LAST,
    i.investor_code;

-- ============================================================================
-- CONSTRAINTS AND DATA INTEGRITY
-- ============================================================================

-- Ensure register lots maintain data integrity
ALTER TABLE "hf-investor".hf_register_lots
    ADD CONSTRAINT chk_hf_register_lots_units_non_negative
    CHECK (units >= 0);

-- Ensure beneficial ownership doesn't exceed 100%
CREATE OR REPLACE FUNCTION "hf-investor".check_beneficial_ownership()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COALESCE(SUM(ownership_percentage), 0)
        FROM "hf-investor".hf_beneficial_owners
        WHERE investor_id = NEW.investor_id) > 100 THEN
        RAISE EXCEPTION 'Total beneficial ownership cannot exceed 100%%';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_hf_beneficial_ownership_check
    AFTER INSERT OR UPDATE ON "hf-investor".hf_beneficial_owners
    FOR EACH ROW
    EXECUTE FUNCTION "hf-investor".check_beneficial_ownership();

-- Ensure only one active bank instruction per currency per investor
CREATE UNIQUE INDEX idx_hf_bank_instructions_unique_active
    ON "hf-investor".hf_bank_instructions (investor_id, currency, instruction_type)
    WHERE status = 'ACTIVE';

-- Comments for documentation
COMMENT ON SCHEMA "hf-investor" IS 'Hedge Fund Investor Register - Complete investor lifecycle management with event sourcing';
COMMENT ON TABLE "hf-investor".hf_investors IS 'Hedge Fund Investor entities with full lifecycle status tracking';
COMMENT ON TABLE "hf-investor".hf_register_lots IS 'Current unit holdings by investor/fund/class/series (derived from events)';
COMMENT ON TABLE "hf-investor".hf_register_events IS 'Immutable event sourcing log of all unit movements and register changes';
COMMENT ON TABLE "hf-investor".hf_trades IS 'All subscription, redemption, and transfer orders with full settlement tracking';
COMMENT ON TABLE "hf-investor".hf_kyc_profiles IS 'KYC/KYB profiles with risk ratings and refresh schedules';
COMMENT ON TABLE "hf-investor".hf_tax_profiles IS 'FATCA/CRS tax classifications and withholding information';
COMMENT ON TABLE "hf-investor".hf_bank_instructions IS 'Multi-currency settlement instructions with versioning';
COMMENT ON TABLE "hf-investor".hf_dsl_executions IS 'Hedge Fund DSL execution log with idempotency and audit trail';