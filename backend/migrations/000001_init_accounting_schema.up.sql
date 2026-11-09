-- 000001_init_accounting_schema.up.sql
-- Enterprise Ultra-Fast ERP Accounting Schema
-- Strictly Zero-Float, Append-Only, Partitioned Ledger

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. Companies & Currency Master
-- ============================================================================
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    base_currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    country VARCHAR(100) NOT NULL DEFAULT 'Bangladesh',
    fiscal_year_start_month SMALLINT NOT NULL DEFAULT 1 CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. Chart of Accounts (Tree Structure)
-- ============================================================================
CREATE TYPE account_root_type AS ENUM ('Asset', 'Liability', 'Equity', 'Income', 'Expense');

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    root_type account_root_type NOT NULL,
    parent_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_account_code UNIQUE (company_id, code)
);

CREATE INDEX idx_accounts_company_parent ON accounts(company_id, parent_id);
CREATE INDEX idx_accounts_root_type ON accounts(company_id, root_type);

-- ============================================================================
-- 3. Journal Entry (Voucher / Transaction Header)
-- ============================================================================
CREATE TYPE doc_status AS ENUM ('draft', 'posted', 'cancelled');

CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    entry_number VARCHAR(100) NOT NULL,
    posting_date DATE NOT NULL,
    status doc_status NOT NULL DEFAULT 'draft',
    total_debit NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    total_credit NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    remarks TEXT,
    posted_at TIMESTAMPTZ,
    posted_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_entry_number UNIQUE (company_id, entry_number)
);

CREATE INDEX idx_journal_entries_date ON journal_entries(company_id, posting_date);
CREATE INDEX idx_journal_entries_status ON journal_entries(company_id, status);

-- ============================================================================
-- 4. General Ledger (gl_entry) - Partitioned by Range on posting_date
-- ============================================================================
CREATE TABLE IF NOT EXISTS gl_entries (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    posting_date DATE NOT NULL,
    voucher_type VARCHAR(50) NOT NULL, -- 'Journal Entry', 'Sales Invoice', 'Purchase Invoice', 'Payment'
    voucher_id UUID NOT NULL,
    voucher_no VARCHAR(100) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    account_code VARCHAR(50) NOT NULL,
    party_type VARCHAR(50),             -- 'Customer', 'Supplier', 'Employee'
    party_id UUID,
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    
    -- Transaction Currency Amounts (Zero Float Policy: NUMERIC 19, 6)
    debit NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (debit >= 0),
    credit NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (credit >= 0),
    exchange_rate NUMERIC(19, 6) NOT NULL DEFAULT 1.000000 CHECK (exchange_rate > 0),
    
    -- Functional / Base Currency Amounts
    debit_base NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (debit_base >= 0),
    credit_base NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (credit_base >= 0),
    
    remarks TEXT,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, posting_date),
    CONSTRAINT chk_gl_debit_credit_exclusive CHECK (NOT (debit > 0 AND credit > 0)),
    CONSTRAINT chk_gl_base_exclusive CHECK (NOT (debit_base > 0 AND credit_base > 0))
) PARTITION BY RANGE (posting_date);

-- Indexes on partitioned master
CREATE INDEX idx_gl_entries_account_date ON gl_entries(company_id, account_id, posting_date);
CREATE INDEX idx_gl_entries_voucher ON gl_entries(company_id, voucher_type, voucher_id);

-- ============================================================================
-- Partitions Setup (Current & Upcoming Years + Default Partition)
-- ============================================================================
CREATE TABLE IF NOT EXISTS gl_entries_2026_h1 PARTITION OF gl_entries
    FOR VALUES FROM ('2026-01-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS gl_entries_2026_h2 PARTITION OF gl_entries
    FOR VALUES FROM ('2026-07-01') TO ('2027-01-01');

CREATE TABLE IF NOT EXISTS gl_entries_2027_h1 PARTITION OF gl_entries
    FOR VALUES FROM ('2027-01-01') TO ('2027-07-01');

CREATE TABLE IF NOT EXISTS gl_entries_default PARTITION OF gl_entries DEFAULT;

-- ============================================================================
-- 5. Immutability Enforcer (Tryton Rule: Strict Financial Consistency)
-- Ledger rows can NEVER be modified or deleted once inserted!
-- Corrections MUST be done via reverse contra entries.
-- ============================================================================
CREATE OR REPLACE FUNCTION fn_enforce_gl_immutability()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'UPDATE') THEN
        RAISE EXCEPTION 'Financial Ledger Violation: gl_entry rows are append-only and cannot be updated. Use a reversal contra entry instead.';
    ELSIF (TG_OP = 'DELETE') THEN
        RAISE EXCEPTION 'Financial Ledger Violation: gl_entry rows are immutable and cannot be deleted.';
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_gl_immutability
BEFORE UPDATE OR DELETE ON gl_entries
FOR EACH ROW EXECUTE FUNCTION fn_enforce_gl_immutability();

-- ============================================================================
-- 6. High-Speed Background Job Queue (FOR UPDATE SKIP LOCKED)
-- Eliminates Redis/RabbitMQ overhead for real-time background processing
-- ============================================================================
CREATE TYPE job_status AS ENUM ('queued', 'processing', 'completed', 'failed');

CREATE TABLE IF NOT EXISTS background_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status job_status NOT NULL DEFAULT 'queued',
    priority INT NOT NULL DEFAULT 0,
    run_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(100),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_queue ON background_jobs (status, run_at, priority DESC, created_at ASC)
WHERE status = 'queued';
