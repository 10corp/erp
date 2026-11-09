-- 000004_payment_entries.up.sql
-- Payment Entry & Forex Realized Gain/Loss Engine
-- Strictly Zero-Float, Multi-Currency Dual-Column, Tryton Immutability

CREATE TABLE IF NOT EXISTS payment_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    payment_number VARCHAR(100) NOT NULL,
    payment_type VARCHAR(20) NOT NULL DEFAULT 'Receive' CHECK (payment_type IN ('Receive', 'Pay')),
    party_type VARCHAR(50) NOT NULL DEFAULT 'Customer',
    party_id UUID,
    party_name VARCHAR(255) NOT NULL,
    posting_date DATE NOT NULL,
    
    -- Ledger Accounts
    paid_to_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,   -- Bank / Cash
    paid_from_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT, -- Accounts Receivable / Payable
    
    -- Currency & Payment Date Exchange Rate
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    exchange_rate NUMERIC(19, 6) NOT NULL DEFAULT 1.000000 CHECK (exchange_rate > 0),
    
    -- Monetary Amounts
    paid_amount NUMERIC(19, 6) NOT NULL CHECK (paid_amount > 0),
    base_paid_amount NUMERIC(19, 6) NOT NULL CHECK (base_paid_amount > 0),
    allocated_amount NUMERIC(19, 6) NOT NULL CHECK (allocated_amount >= 0),
    
    -- Invoice Link & Forex Tracking
    reference_invoice_id UUID REFERENCES sales_invoices(id) ON DELETE RESTRICT,
    reference_invoice_number VARCHAR(100),
    reference_invoice_rate NUMERIC(19, 6) DEFAULT 1.000000 CHECK (reference_invoice_rate > 0),
    
    -- Forex Gain/Loss: positive = gain, negative = loss
    forex_gain_loss NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    forex_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    
    -- Tryton Strict State Machine (0: Draft, 1: Submitted/Posted, 2: Cancelled)
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    remarks TEXT,
    posted_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_payment_number UNIQUE (company_id, payment_number)
);

CREATE INDEX idx_payment_entries_company_date ON payment_entries(company_id, posting_date);
CREATE INDEX idx_payment_entries_status ON payment_entries(company_id, docstatus);
CREATE INDEX idx_payment_entries_invoice ON payment_entries(reference_invoice_id);
