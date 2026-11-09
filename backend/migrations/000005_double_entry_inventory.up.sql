-- 000005_double_entry_inventory.up.sql
-- Tryton Double-Entry Stock & Perpetual Inventory (COGS Auto-Posting)

-- ============================================================================
-- 1. Warehouses & Double-Entry Locations (Internal, Supplier, Customer, Transit)
-- ============================================================================
CREATE TYPE warehouse_type_enum AS ENUM ('Internal', 'Supplier', 'Customer', 'Transit', 'Waste');

CREATE TABLE IF NOT EXISTS warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    warehouse_type warehouse_type_enum NOT NULL DEFAULT 'Internal',
    account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT, -- Default Stock Asset Account (1030)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_warehouse_code UNIQUE (company_id, code)
);

CREATE INDEX idx_warehouses_company ON warehouses(company_id, warehouse_type);

-- ============================================================================
-- 2. Item Warehouse Balances (Row-Level Locking & Moving Average Engine)
-- ============================================================================
CREATE TABLE IF NOT EXISTS item_warehouse_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    item_code VARCHAR(100) NOT NULL,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,
    current_qty NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (current_qty >= 0), -- Strict Negative Stock Prevention
    valuation_rate NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (valuation_rate >= 0),
    stock_value NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (stock_value >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_item_warehouse UNIQUE (company_id, item_code, warehouse_id)
);

CREATE INDEX idx_item_warehouse_balances ON item_warehouse_balances(company_id, item_code, warehouse_id);

-- ============================================================================
-- 3. Stock Ledger Entries (Append-Only Double-Entry Inventory Ledger)
-- ============================================================================
CREATE TABLE IF NOT EXISTS stock_ledger_entries (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    posting_datetime TIMESTAMPTZ NOT NULL,
    item_code VARCHAR(100) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    
    -- Tryton Double-Entry Locations
    from_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,
    to_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,
    
    voucher_type VARCHAR(50) NOT NULL, -- 'Delivery Note', 'Purchase Receipt', 'Stock Transfer'
    voucher_id UUID NOT NULL,
    voucher_no VARCHAR(100) NOT NULL,
    
    -- Moving Average Quantities & Values
    actual_qty NUMERIC(19, 6) NOT NULL,
    valuation_rate NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (valuation_rate >= 0),
    stock_value_difference NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    qty_after_transaction NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    stock_value_after_transaction NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    
    remarks TEXT,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, posting_datetime)
) PARTITION BY RANGE (posting_datetime);

-- Stock Ledger Partitions (2026, 2027, Default)
CREATE TABLE IF NOT EXISTS stock_ledger_entries_2026_h1 PARTITION OF stock_ledger_entries
    FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-07-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS stock_ledger_entries_2026_h2 PARTITION OF stock_ledger_entries
    FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS stock_ledger_entries_default PARTITION OF stock_ledger_entries DEFAULT;

CREATE INDEX idx_stock_ledger_item_date ON stock_ledger_entries(company_id, item_code, posting_datetime);
CREATE INDEX idx_stock_ledger_voucher ON stock_ledger_entries(company_id, voucher_type, voucher_id);

-- ============================================================================
-- 4. Delivery Notes (Perpetual COGS & Stock Dispatch Header)
-- ============================================================================
CREATE TABLE IF NOT EXISTS delivery_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    delivery_note_number VARCHAR(100) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    posting_date DATE NOT NULL,
    from_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,
    to_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT, -- Customer Virtual Location
    cogs_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,   -- 5010 - Cost of Goods Sold
    stock_asset_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT, -- 1030 - Stock In Hand
    total_qty NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (total_qty > 0),
    total_cogs_amount NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (total_cogs_amount >= 0),
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    remarks TEXT,
    posted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_delivery_number UNIQUE (company_id, delivery_note_number)
);

CREATE TABLE IF NOT EXISTS delivery_note_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES delivery_notes(id) ON DELETE CASCADE,
    item_code VARCHAR(100) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    qty NUMERIC(19, 6) NOT NULL CHECK (qty > 0),
    valuation_rate NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (valuation_rate >= 0),
    cogs_amount NUMERIC(19, 6) NOT NULL DEFAULT 0.000000 CHECK (cogs_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_items_parent ON delivery_note_items(parent_id);
