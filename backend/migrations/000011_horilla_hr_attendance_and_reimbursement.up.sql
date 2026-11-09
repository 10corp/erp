-- 000011_horilla_hr_attendance_and_reimbursement.up.sql
-- Horilla HRMS-inspired Attendance Tracking, Shifts & Expense Reimbursements
-- 1. hr_shifts: Employee Work Shifts & Grace Periods
-- 2. hr_attendance_records: Daily Check-In/Out & Overtime
-- 3. hr_reimbursement_claims: Employee Expense Claims with Auto-GL Posting

-- ============================================================================
-- 1. Shifts Master (Horilla: base.EmployeeShift)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    grace_time_minutes INT NOT NULL DEFAULT 15,
    standard_hours NUMERIC(5, 2) NOT NULL DEFAULT 8.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_shift_code UNIQUE (company_id, code)
);

-- ============================================================================
-- 2. Attendance Records (Horilla: attendance.Attendance)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES company_employee(id) ON DELETE RESTRICT,
    attendance_date DATE NOT NULL,
    check_in TIMESTAMPTZ,
    check_out TIMESTAMPTZ,
    working_hours NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    overtime_hours NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'Present', -- 'Present', 'Late', 'Half Day', 'Absent'
    check_in_source VARCHAR(50) NOT NULL DEFAULT 'Web', -- 'Biometric', 'Web', 'GPS'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_employee_attendance_date UNIQUE (company_id, employee_id, attendance_date)
);

CREATE INDEX idx_attendance_date_emp ON hr_attendance_records(company_id, attendance_date, employee_id);

-- ============================================================================
-- 3. Expense Reimbursements (Horilla: payroll.Reimbursement)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr_reimbursement_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) DEFAULT '00000000-0000-0000-0000-000000000001',
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    claim_number VARCHAR(100) NOT NULL,
    employee_id UUID NOT NULL REFERENCES company_employee(id) ON DELETE RESTRICT,
    claim_date DATE NOT NULL,
    category VARCHAR(50) NOT NULL, -- 'Travel', 'Medical', 'Client Entertainment', 'Office Supplies'
    amount NUMERIC(19, 6) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'BDT',
    description TEXT NOT NULL,
    receipt_reference VARCHAR(255),
    
    -- Tryton & ERPNext docstatus
    docstatus SMALLINT NOT NULL DEFAULT 0 CHECK (docstatus IN (0, 1, 2)), -- 0=Draft, 1=Approved/GL Posted, 2=Cancelled
    journal_entry_id UUID REFERENCES journal_entries(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_company_claim_number UNIQUE (company_id, claim_number)
);

CREATE INDEX idx_reimbursements_emp ON hr_reimbursement_claims(company_id, employee_id, docstatus);
