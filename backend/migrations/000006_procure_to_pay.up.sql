-- 000006_procure_to_pay.up.sql
-- Procure-to-Pay (P2P): Purchase Orders, Purchase Invoices & Goods Receipts

-- Purchase Orders
CREATE TABLE IF NOT EXISTS purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    po_number VARCHAR(100) NOT NULL,
    supplier_id UUID,
    supplier_name VARCHAR(255) NOT NULL,
    transaction_date DATE NOT NULL,
    delivery_date DATE,
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    conversion_rate NUMERIC(19, 6) NOT NULL DEFAULT 1.000000 CHECK (conversion_rate > 0),
    
    net_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (net_total >= 0),
    tax_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_total >= 0),
    grand_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (grand_total >= 0),
    
    base_net_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_net_total >= 0),
    base_tax_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_tax_total >= 0),
    base_grand_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_grand_total >= 0),
    
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    status VARCHAR(50) NOT NULL DEFAULT 'Draft',
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_po_number UNIQUE (company_id, po_number)
);

CREATE INDEX idx_purchase_orders_supplier ON purchase_orders(company_id, supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(company_id, docstatus);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    item_code VARCHAR(100) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    description TEXT,
    warehouse_id UUID REFERENCES warehouses(id) ON DELETE RESTRICT,
    qty NUMERIC(19, 6) NOT NULL CHECK (qty > 0),
    rate NUMERIC(19, 6) NOT NULL CHECK (rate >= 0),
    amount NUMERIC(19, 6) NOT NULL CHECK (amount >= 0),
    tax_rate NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_rate >= 0),
    tax_amount NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_amount >= 0),
    received_qty NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (received_qty >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_po_items_parent ON purchase_order_items(parent_id);

-- Purchase Invoices (Accounts Payable Liability)
CREATE TABLE IF NOT EXISTS purchase_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    invoice_number VARCHAR(100) NOT NULL,
    supplier_invoice_no VARCHAR(100),
    supplier_id UUID,
    supplier_name VARCHAR(255) NOT NULL,
    posting_date DATE NOT NULL,
    due_date DATE NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    conversion_rate NUMERIC(19, 6) NOT NULL DEFAULT 1.000000 CHECK (conversion_rate > 0),
    
    net_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (net_total >= 0),
    tax_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (tax_total >= 0),
    grand_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (grand_total >= 0),
    
    base_net_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_net_total >= 0),
    base_tax_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_tax_total >= 0),
    base_grand_total NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (base_grand_total >= 0),
    
    credit_to_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT, -- Accounts Payable (2001)
    expense_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,   -- Stock in Hand (1030) or Expense
    tax_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,               -- Input VAT / Tax Receivable
    
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    remarks TEXT,
    posted_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_purchase_invoice_no UNIQUE (company_id, invoice_number)
);

CREATE INDEX idx_purchase_invoices_date ON purchase_invoices(company_id, posting_date);
CREATE INDEX idx_purchase_invoices_status ON purchase_invoices(company_id, docstatus);

CREATE TABLE IF NOT EXISTS purchase_invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES purchase_invoices(id) ON DELETE CASCADE,
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

CREATE INDEX idx_purchase_invoice_items_parent ON purchase_invoice_items(parent_id);
