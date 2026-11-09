-- 000012_horilla_hr_loans_recruitment_and_assets.up.sql
-- Horilla HRMS Advanced Pillars:
-- 1. Employee Loans & Advance Salary (Auto-EMI Salary Slip Deductions & Tryton GL Posting)
-- 2. Recruitment & Candidate ATS Pipeline (Job Openings, Hiring Stages & 1-Click Employee Promotion)
-- 3. Company Asset Custody & Allocation (Laptops, Devices, Equipment & Offboarding Return Checklist)

-- ============================================================================
-- 1. Employee Loans & Advances (Horilla: payroll.LoanAccount)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_loans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES company_employee(id) ON DELETE RESTRICT,
    loan_number VARCHAR(50) NOT NULL,
    loan_type VARCHAR(30) NOT NULL DEFAULT 'loan', -- 'loan', 'advanced_salary', 'fine'
    title VARCHAR(255) NOT NULL,
    principal_amount NUMERIC(19, 6) NOT NULL,
    monthly_installment NUMERIC(19, 6) NOT NULL,
    total_installments INT NOT NULL DEFAULT 1,
    paid_installments INT NOT NULL DEFAULT 0,
    disbursement_date DATE NOT NULL,
    start_date DATE NOT NULL,
    settled BOOLEAN NOT NULL DEFAULT FALSE,
    settled_date TIMESTAMPTZ,
    journal_entry_id UUID REFERENCES gl_entries(id),
    voucher_no VARCHAR(100),
    docstatus INT NOT NULL DEFAULT 0, -- 0=Draft, 1=Disbursed/Active, 2=Settled, 3=Cancelled
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_loan_number UNIQUE (company_id, loan_number)
);

CREATE INDEX idx_hr_loans_emp ON hr_loans(company_id, employee_id, settled);

CREATE TABLE IF NOT EXISTS hr_loan_repayments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id UUID NOT NULL REFERENCES hr_loans(id) ON DELETE CASCADE,
    salary_slip_id UUID REFERENCES hr_salary_slips(id) ON DELETE SET NULL,
    installment_number INT NOT NULL,
    amount NUMERIC(19, 6) NOT NULL,
    payment_date DATE NOT NULL,
    voucher_no VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_loan_repayments_loan ON hr_loan_repayments(loan_id);

-- ============================================================================
-- 2. Recruitment & ATS Pipeline (Horilla: recruitment.Recruitment & recruitment.Candidate)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_job_openings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    job_title VARCHAR(255) NOT NULL,
    department_id UUID REFERENCES hr_departments(id) ON DELETE SET NULL,
    designation_id UUID REFERENCES hr_designations(id) ON DELETE SET NULL,
    vacancies INT NOT NULL DEFAULT 1,
    experience_years INT NOT NULL DEFAULT 0,
    employment_type VARCHAR(50) NOT NULL DEFAULT 'Full-Time', -- 'Full-Time', 'Part-Time', 'Contract', 'Internship'
    target_date DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'Active', -- 'Draft', 'Active', 'Closed'
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hr_candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    job_opening_id UUID NOT NULL REFERENCES hr_job_openings(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    stage VARCHAR(50) NOT NULL DEFAULT 'Applied', -- 'Applied', 'Screening', 'Interview', 'Offer', 'Hired', 'Rejected'
    rating INT NOT NULL DEFAULT 3, -- 1 to 5 stars
    expected_salary NUMERIC(19, 6) NOT NULL DEFAULT 0,
    current_company VARCHAR(255),
    resume_url VARCHAR(500),
    notes TEXT,
    converted_employee_id UUID REFERENCES company_employee(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hr_candidates_job_stage ON hr_candidates(company_id, job_opening_id, stage);

-- ============================================================================
-- 3. Company Asset Custody & Allocation (Horilla: asset.Asset & asset.AssetAssignment)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    category_name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hr_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    category_id UUID NOT NULL REFERENCES hr_asset_categories(id) ON DELETE RESTRICT,
    asset_name VARCHAR(255) NOT NULL,
    asset_code VARCHAR(50) NOT NULL,
    serial_number VARCHAR(100) NOT NULL,
    purchase_date DATE NOT NULL,
    purchase_cost NUMERIC(19, 6) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'Available', -- 'Available', 'Allocated', 'Under Maintenance', 'Retired'
    warranty_expiry DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_asset_code UNIQUE (company_id, asset_code)
);

CREATE INDEX idx_hr_assets_company_status ON hr_assets(company_id, status);

CREATE TABLE IF NOT EXISTS hr_asset_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES hr_assets(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES company_employee(id) ON DELETE RESTRICT,
    allocated_date DATE NOT NULL,
    return_due_date DATE,
    returned_date DATE,
    condition_on_alloc VARCHAR(100) NOT NULL DEFAULT 'New / Excellent',
    condition_on_return VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'Active', -- 'Active', 'Returned', 'Damaged'
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hr_asset_alloc_emp ON hr_asset_allocations(employee_id, status);
