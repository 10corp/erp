-- 000008_core_party_company.up.sql
-- Tryton-inspired Universal Party & Multi-Branch Company Architecture
-- 1. party_party (Universal Legal Entity)
-- 2. party_identifier (TIN, BIN, VAT, NID, Trade License)
-- 3. party_address (Billing, Shipping, Multi-location)
-- 4. party_contact_mechanism (Phone, Mobile, Email, WhatsApp)
-- 5. company_company enhancements (Multi-branch / Subsidiary tree)
-- 6. company_employee (Employee Profile linked to party_party)
-- 7. users binding (Active Branch, Main Company, Employee binding)

-- ============================================================================
-- 1. Universal Party Table (party.party)
-- ============================================================================
CREATE TABLE IF NOT EXISTS party_party (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_customer BOOLEAN NOT NULL DEFAULT FALSE,
    is_supplier BOOLEAN NOT NULL DEFAULT FALSE,
    is_employee BOOLEAN NOT NULL DEFAULT FALSE,
    is_company BOOLEAN NOT NULL DEFAULT FALSE,
    lang VARCHAR(10) NOT NULL DEFAULT 'en',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_party_code ON party_party(code);
CREATE INDEX idx_party_name_lower ON party_party(LOWER(name));
CREATE INDEX idx_party_roles ON party_party(is_customer, is_supplier, is_employee, is_company);

-- Convenient alias view for concise Go & SQL access
CREATE OR REPLACE VIEW parties AS SELECT * FROM party_party;

-- ============================================================================
-- 2. Party Identifiers (party.identifier - BIN, TIN, VAT, NID)
-- ============================================================================
CREATE TABLE IF NOT EXISTS party_identifier (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id UUID NOT NULL REFERENCES party_party(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'BIN', 'TIN', 'VAT', 'NID', 'TRADE_LICENSE', 'PASSPORT'
    code VARCHAR(100) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_party_identifier UNIQUE (party_id, type, code)
);

CREATE INDEX idx_party_identifier_lookup ON party_identifier(type, code);

-- ============================================================================
-- 3. Party Addresses (party.address - Invoicing, Delivery)
-- ============================================================================
CREATE TABLE IF NOT EXISTS party_address (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id UUID NOT NULL REFERENCES party_party(id) ON DELETE CASCADE,
    name VARCHAR(255), -- Attention / Care of
    street TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    postal_code VARCHAR(30),
    country VARCHAR(100) NOT NULL DEFAULT 'Bangladesh',
    is_invoice BOOLEAN NOT NULL DEFAULT TRUE,
    is_delivery BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_party_address_party ON party_address(party_id);

-- ============================================================================
-- 4. Party Contact Mechanisms (party.contact_mechanism - Email, Phone, etc.)
-- ============================================================================
CREATE TABLE IF NOT EXISTS party_contact_mechanism (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id UUID NOT NULL REFERENCES party_party(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL, -- 'phone', 'mobile', 'email', 'website', 'whatsapp', 'other'
    value VARCHAR(255) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_party_contact_mechanism_party ON party_contact_mechanism(party_id, type);

-- ============================================================================
-- 5. Company Multi-Branch & Subsidiary Tree Enhancements
-- ============================================================================
ALTER TABLE companies ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES companies(id) ON DELETE RESTRICT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS party_id UUID REFERENCES party_party(id) ON DELETE RESTRICT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Dhaka';
ALTER TABLE companies ADD COLUMN IF NOT EXISTS is_branch BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS branch_code VARCHAR(50);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS header TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS footer TEXT;

CREATE INDEX IF NOT EXISTS idx_companies_parent ON companies(parent_id);
CREATE INDEX IF NOT EXISTS idx_companies_party ON companies(party_id);

-- Tryton-compatible view alias
CREATE OR REPLACE VIEW company_company AS SELECT * FROM companies;

-- ============================================================================
-- 6. Company Employee (company.employee linked to party_party)
-- ============================================================================
CREATE TABLE IF NOT EXISTS company_employee (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id UUID NOT NULL REFERENCES party_party(id) ON DELETE RESTRICT,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    employee_code VARCHAR(50) NOT NULL,
    designation VARCHAR(100),
    department VARCHAR(100),
    supervisor_id UUID REFERENCES company_employee(id) ON DELETE SET NULL,
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_employee_code UNIQUE (company_id, employee_code)
);

CREATE INDEX idx_employee_party ON company_employee(party_id);
CREATE INDEX idx_employee_company ON company_employee(company_id);

-- ============================================================================
-- 7. Users Table Binding (Active Branch, Main Company, Employee Binding)
-- ============================================================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS employee_id UUID REFERENCES company_employee(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS main_company_id UUID REFERENCES companies(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_users_employee ON users(employee_id);
CREATE INDEX IF NOT EXISTS idx_users_main_company ON users(main_company_id);

-- ============================================================================
-- 8. Multi-Branch Hierarchy Recursive CTE View
-- ============================================================================
CREATE OR REPLACE VIEW vw_company_hierarchy AS
WITH RECURSIVE branch_tree AS (
    -- Root Holding Companies
    SELECT 
        c.id AS company_id,
        c.name AS company_name,
        c.id AS root_company_id,
        c.name AS root_company_name,
        c.parent_id,
        c.is_branch,
        c.branch_code,
        1 AS hierarchy_level,
        ARRAY[c.id] AS path_ids,
        ARRAY[c.name::text] AS path_names
    FROM companies c
    WHERE c.parent_id IS NULL
    
    UNION ALL
    
    -- Sub-branches & subsidiaries
    SELECT 
        child.id AS company_id,
        child.name AS company_name,
        bt.root_company_id,
        bt.root_company_name,
        child.parent_id,
        child.is_branch,
        child.branch_code,
        bt.hierarchy_level + 1 AS hierarchy_level,
        bt.path_ids || child.id,
        bt.path_names || child.name::text
    FROM companies child
    JOIN branch_tree bt ON child.parent_id = bt.company_id
)
SELECT * FROM branch_tree;
