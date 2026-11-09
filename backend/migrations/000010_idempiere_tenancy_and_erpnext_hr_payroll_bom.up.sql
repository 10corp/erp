-- 000010_idempiere_tenancy_and_erpnext_hr_payroll_bom.up.sql
-- 1. iDempiere-style Multi-Tenancy (tenant_id / AD_Client_ID & company_id / AD_Org_ID)
-- 2. Tryton Core SQL Compatibility Aliases (account_account, account_move, account_move_line, stock_location, stock_move)
-- 3. ERPNext-style HR & Payroll Subsystem (docstatus: 0=Draft, 1=Submitted, 2=Cancelled)
-- 4. ERPNext-style Manufacturing & Bill of Materials (BOM & Work Order)

-- ============================================================================
-- 1. iDempiere Multi-Tenant Client Isolation
-- ============================================================================
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Default Root Tenant
INSERT INTO tenants (id, code, name, is_active)
VALUES ('00000000-0000-0000-0000-000000000001', 'TENANT-001', 'Enterprise ERP Global Holding Tenant', TRUE)
ON CONFLICT (id) DO NOTHING;

-- Enforce tenant_id on core transactional tables (iDempiere standard)
ALTER TABLE companies ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE party_party ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE warehouses ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE gl_entries ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE stock_ledger_entries ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001';

CREATE INDEX IF NOT EXISTS idx_companies_tenant ON companies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_party_tenant ON party_party(tenant_id);
CREATE INDEX IF NOT EXISTS idx_accounts_tenant ON accounts(tenant_id);

-- ============================================================================
-- 2. Tryton Explicit SQL Views (Strict Accounting & Double-Entry Stock)
-- ============================================================================
CREATE OR REPLACE VIEW account_account AS SELECT * FROM accounts;
CREATE OR REPLACE VIEW account_move AS SELECT * FROM journal_entries;
CREATE OR REPLACE VIEW account_move_line AS SELECT * FROM gl_entries;
CREATE OR REPLACE VIEW stock_location AS SELECT * FROM warehouses;
CREATE OR REPLACE VIEW stock_move AS SELECT * FROM stock_ledger_entries;

-- ============================================================================
-- 3. ERPNext HR & Payroll Subsystem (docstatus: 0=Draft, 1=Submitted, 2=Cancelled)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES hr_departments(id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_dept_code UNIQUE (company_id, code)
);

CREATE TABLE IF NOT EXISTS hr_designations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_desig_code UNIQUE (company_id, code)
);

CREATE TABLE IF NOT EXISTS hr_leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    max_days_allowed INT NOT NULL DEFAULT 14,
    is_paid BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_leave_name UNIQUE (company_id, name)
);

CREATE TABLE IF NOT EXISTS hr_leave_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES company_employee(id) ON DELETE RESTRICT,
    leave_type_id UUID NOT NULL REFERENCES hr_leave_types(id) ON DELETE RESTRICT,
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    total_days INT NOT NULL CHECK (total_days > 0),
    reason TEXT NOT NULL,
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)), -- 0=Draft, 1=Approved/Submitted, 2=Cancelled
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hr_salary_structures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    base_salary NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    house_rent NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    medical_allowance NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    conveyance NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    provident_fund_deduction NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    tax_deduction NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_salary_structure UNIQUE (company_id, name)
);

CREATE TABLE IF NOT EXISTS hr_salary_slips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    slip_number VARCHAR(100) NOT NULL,
    employee_id UUID NOT NULL REFERENCES company_employee(id) ON DELETE RESTRICT,
    posting_date DATE NOT NULL,
    month SMALLINT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INT NOT NULL,
    
    -- Zero-Float Earnings & Deductions
    gross_pay NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    total_deduction NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    net_pay NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    
    -- ERPNext Document Lifecycle (docstatus)
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    journal_entry_id UUID REFERENCES journal_entries(id) ON DELETE SET NULL, -- Linked Auto-Posted GL Move
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_salary_slip_period UNIQUE (company_id, employee_id, month, year)
);

CREATE INDEX idx_salary_slips_emp_period ON hr_salary_slips(company_id, employee_id, year, month);
CREATE INDEX idx_salary_slips_docstatus ON hr_salary_slips(company_id, docstatus);

-- ============================================================================
-- 4. ERPNext Manufacturing & Bill of Materials (BOM & Work Orders)
-- ============================================================================
CREATE TABLE IF NOT EXISTS manufacturing_boms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    bom_number VARCHAR(100) NOT NULL,
    item_code VARCHAR(100) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    quantity NUMERIC(19, 6) NOT NULL DEFAULT 1.000000 CHECK (quantity > 0),
    uom VARCHAR(20) NOT NULL DEFAULT 'Nos',
    
    -- Zero-Float Cost Accumulator
    total_raw_material_cost NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    operating_cost NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    total_cost NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    
    -- ERPNext Document Lifecycle
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_bom_number UNIQUE (company_id, bom_number)
);

CREATE TABLE IF NOT EXISTS manufacturing_bom_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bom_id UUID NOT NULL REFERENCES manufacturing_boms(id) ON DELETE CASCADE,
    raw_item_code VARCHAR(100) NOT NULL,
    raw_item_name VARCHAR(255) NOT NULL,
    qty NUMERIC(19, 6) NOT NULL CHECK (qty > 0),
    uom VARCHAR(20) NOT NULL DEFAULT 'Nos',
    rate NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    amount NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    scrap_percentage NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bom_items_parent ON manufacturing_bom_items(bom_id);

CREATE TABLE IF NOT EXISTS manufacturing_work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    work_order_no VARCHAR(100) NOT NULL,
    bom_id UUID NOT NULL REFERENCES manufacturing_boms(id) ON DELETE RESTRICT,
    production_item_code VARCHAR(100) NOT NULL,
    qty_to_produce NUMERIC(19, 6) NOT NULL CHECK (qty_to_produce > 0),
    produced_qty NUMERIC(19, 6) NOT NULL DEFAULT 0.000000,
    
    -- Tryton Double-Entry Stock Location Routing
    source_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT, -- Raw Materials Store
    wip_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,    -- Work-In-Progress Floor
    target_warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT, -- Finished Goods Store
    
    planned_start_date DATE NOT NULL,
    planned_end_date DATE,
    actual_end_date DATE,
    
    -- ERPNext Document Lifecycle
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)),
    status VARCHAR(50) NOT NULL DEFAULT 'Draft', -- 'Draft', 'Submitted', 'In-Progress', 'Completed', 'Cancelled'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_work_order UNIQUE (company_id, work_order_no)
);

CREATE INDEX idx_work_orders_company_status ON manufacturing_work_orders(company_id, docstatus, status);
