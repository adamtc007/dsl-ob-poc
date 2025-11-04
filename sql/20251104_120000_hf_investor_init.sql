-- +goose Up
-- Schema and core tables for "hf-investor"
CREATE SCHEMA IF NOT EXISTS "hf-investor";

-- enable gen_random_uuid() if available
CREATE EXTENSION IF NOT EXISTS pgcrypto;

SET search_path TO "hf-investor", public;

-- Funds
CREATE TABLE IF NOT EXISTS hf_funds (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  legal_name       text NOT NULL,
  short_code       text NOT NULL UNIQUE,
  base_currency    char(3) NOT NULL,
  inception_date   date,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-- Share Classes
CREATE TABLE IF NOT EXISTS hf_share_classes (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  fund_id          uuid NOT NULL REFERENCES hf_funds(id) ON DELETE CASCADE,
  class_code       text NOT NULL,
  currency         char(3) NOT NULL,
  dealing_frequency text DEFAULT 'DAILY',
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (fund_id, class_code)
);

-- Series
CREATE TABLE IF NOT EXISTS hf_series (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id         uuid NOT NULL REFERENCES hf_share_classes(id) ON DELETE CASCADE,
  series_code      text NOT NULL,
  inception_date   date NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (class_id, series_code)
);

-- Investors
CREATE TABLE IF NOT EXISTS hf_investors (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  legal_name       text NOT NULL,
  investor_type    text,
  domicile         char(2),
  status           text NOT NULL DEFAULT 'OPPORTUNITY',
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-- Beneficial Owners
CREATE TABLE IF NOT EXISTS hf_beneficial_owners (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  full_name        text NOT NULL,
  ownership_pct    numeric(6,4),
  country          char(2),
  dob              date,
  created_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (investor_id, full_name)
);

-- KYC Profiles
CREATE TABLE IF NOT EXISTS hf_kyc_profiles (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL UNIQUE REFERENCES hf_investors(id) ON DELETE CASCADE,
  jurisdiction     char(2) NOT NULL,
  program_code     text,
  risk_rating      integer,
  status           text,
  next_refresh_date date,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-- Tax Profiles
CREATE TABLE IF NOT EXISTS hf_tax_profiles (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL UNIQUE REFERENCES hf_investors(id) ON DELETE CASCADE,
  form_type        text NOT NULL,
  tin              text,
  country          char(2),
  status           text,
  effective_date   date,
  expiry_date      date,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

-- Documents
CREATE TABLE IF NOT EXISTS hf_documents (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  doc_type         text NOT NULL,
  file_ref         text,
  status           text,
  uploaded_at      timestamptz NOT NULL DEFAULT now(),
  checked_at       timestamptz,
  checksum         text
);

-- Document Requirements (links to documents when fulfilled)
CREATE TABLE IF NOT EXISTS hf_document_requirements (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  doc_type         text NOT NULL,
  status           text NOT NULL,
  requested_at     timestamptz,
  due_at           timestamptz,
  fulfilled_at     timestamptz,
  source           text,
  document_id      uuid REFERENCES hf_documents(id)
);

-- Bank Instructions
CREATE TABLE IF NOT EXISTS hf_bank_instructions (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  currency         char(3) NOT NULL,
  bank_name        text,
  beneficiary_name text,
  account_number   text,
  iban             text,
  swift            text,
  effective_date   date,
  is_active        boolean NOT NULL DEFAULT true,
  created_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (investor_id, currency, effective_date)
);

-- Trades
CREATE TABLE IF NOT EXISTS hf_trades (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  fund_id          uuid NOT NULL REFERENCES hf_funds(id) ON DELETE CASCADE,
  class_id         uuid NOT NULL REFERENCES hf_share_classes(id) ON DELETE CASCADE,
  series_id        uuid REFERENCES hf_series(id) ON DELETE SET NULL,
  trade_type       text NOT NULL, -- SUBSCRIPTION | REDEMPTION | ADJUST
  amount           numeric(20,4),
  currency         char(3),
  units            numeric(24,8),
  nav_per_unit     numeric(18,8),
  dealing_date     date,
  value_date       date,
  status           text NOT NULL DEFAULT 'PENDING',
  idempotency_key  text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (idempotency_key)
);

-- Register Lots
CREATE TABLE IF NOT EXISTS hf_register_lots (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  fund_id          uuid NOT NULL REFERENCES hf_funds(id) ON DELETE CASCADE,
  class_id         uuid NOT NULL REFERENCES hf_share_classes(id) ON DELETE CASCADE,
  series_id        uuid REFERENCES hf_series(id) ON DELETE SET NULL,
  units            numeric(24,8) NOT NULL DEFAULT 0,
  first_trade_date date,
  last_activity_at timestamptz,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (investor_id, fund_id, class_id, series_id)
);

-- Register Events
CREATE TABLE IF NOT EXISTS hf_register_events (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  lot_id           uuid NOT NULL REFERENCES hf_register_lots(id) ON DELETE CASCADE,
  trade_id         uuid REFERENCES hf_trades(id) ON DELETE SET NULL,
  event_type       text NOT NULL, -- ISSUE | REDEEM | ADJUST | FEE | DIST
  delta_units      numeric(24,8) NOT NULL,
  value_date       date NOT NULL,
  event_timestamp  timestamptz NOT NULL DEFAULT now(),
  event_key        text NOT NULL,
  event_version    integer NOT NULL DEFAULT 1,
  correlation_id   text,
  causation_id     text,
  nav_per_unit     numeric(18,8),
  amount           numeric(20,4),
  CHECK (delta_units <> 0),
  UNIQUE (event_key)
);

-- Lifecycle journal
CREATE TABLE IF NOT EXISTS hf_lifecycle_states (
  id               bigserial PRIMARY KEY,
  investor_id      uuid NOT NULL REFERENCES hf_investors(id) ON DELETE CASCADE,
  from_status      text,
  to_status        text NOT NULL,
  changed_at       timestamptz NOT NULL DEFAULT now(),
  by_user          text,
  reason           text,
  guard_results    jsonb,
  context          jsonb
);

-- DSL executions
CREATE TABLE IF NOT EXISTS hf_dsl_executions (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  runbook_id       uuid,
  step_id          uuid,
  investor_id      uuid,
  status           text NOT NULL, -- PENDING | RUNNING | SUCCESS | FAILED
  started_at       timestamptz NOT NULL DEFAULT now(),
  ended_at         timestamptz,
  error            text
);

-- Audit events
CREATE TABLE IF NOT EXISTS hf_audit_events (
  id               bigserial PRIMARY KEY,
  entity_type      text NOT NULL,
  entity_id        uuid,
  action           text NOT NULL,
  at               timestamptz NOT NULL DEFAULT now(),
  by_user          text,
  data             jsonb
);

-- Indexes (hot paths)
CREATE INDEX IF NOT EXISTS idx_investors_status ON hf_investors(status);
CREATE INDEX IF NOT EXISTS idx_trades_pending ON hf_trades(status) WHERE status = 'PENDING';
CREATE INDEX IF NOT EXISTS idx_events_lot_date_ts ON hf_register_events(lot_id, value_date, event_timestamp);
CREATE INDEX IF NOT EXISTS idx_docs_status ON hf_documents(status);
CREATE INDEX IF NOT EXISTS idx_docreqs_open ON hf_document_requirements(status) WHERE status IN ('REQUESTED','OVERDUE');

-- Projection trigger: keep lots.units in sync with events
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION apply_register_event() RETURNS trigger AS $
BEGIN
  UPDATE hf_register_lots
     SET units = units + NEW.delta_units,
         last_activity_at = COALESCE(NEW.event_timestamp, now()),
         updated_at = now()
   WHERE id = NEW.lot_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Lot % not found for register event %', NEW.lot_id, NEW.id;
  END IF;

  RETURN NEW;
END
$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_apply_register_event ON hf_register_events;
CREATE TRIGGER trg_apply_register_event
AFTER INSERT ON hf_register_events
FOR EACH ROW EXECUTE FUNCTION apply_register_event();