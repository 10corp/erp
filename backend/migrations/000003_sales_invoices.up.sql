-- 000003_sales_invoices.up.sql
-- ERPNext Schema + Tryton Immutability for Sales Invoicing

CREATE TABLE IF NOT EXISTS sales_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    invoice_number VARCHAR(100) NOT NULL,
    customer_id UUID,
    customer_name VARCHAR(255) NOT NULL,
    posting_date DATE NOT NULL,
    due_date DATE NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    conversion_rate NUMERIC(19, 6) NOT NULL DEFAULT 1.000000 CHECK (conversion_rate > 0),
    
    -- Monetary totals in Transaction Currency
    net_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (net_total >= 0),
    tax_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_total >= 0),
    grand_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (grand_total >= 0),
    
    -- Monetary totals in Base Currency (Functional Ledger Currency)
    base_net_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_net_total >= 0),
    base_tax_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_tax_total >= 0),
    base_grand_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_grand_total >= 0),
    
    -- Ledger Mapping Accounts
    debit_to_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT, -- Accounts Receivable
    income_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,   -- Sales Revenue
    tax_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,               -- VAT / Tax Payable
    
    -- Tryton Strict State Machine (0: Draft, 1: Submitted/Posted, 2: Cancelled)
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    remarks TEXT,
    posted_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_invoice_number UNIQUE (company_id, invoice_number)
);

CREATE INDEX idx_sales_invoices_company_date ON sales_invoices(company_id, posting_date);
CREATE INDEX idx_sales_invoices_status ON sales_invoices(company_id, docstatus);

CREATE TABLE IF NOT EXISTS sales_invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    item_code VARCHAR(100) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    description TEXT,
    qty NUMERIC(19, 6) NOT NULL CHECK (qty > 0),
    rate NUMERIC(19, 6) NOT NULL CHECK (rate >= 0),
    amount NUMERIC(19, 6) NOT NULL CHECK (amount >= 0),
    tax_rate NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_rate >= 0),
    tax_amount NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_amount >= 0),
    base_rate NUMERIC(19, 6) NOT NULL CHECK (base_rate >= 0),
    base_amount NUMERIC(19, 6) NOT NULL CHECK (base_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_invoice_items_parent ON sales_invoice_items(parent_id);
