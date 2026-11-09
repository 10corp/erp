-- 000009_gl_subledger_and_interbranch.up.sql
-- General Ledger Sub-Ledger Integrity, Inter-Branch Clearing & Consolidated Trial Balance

-- 1. Composite Index for vw_company_hierarchy Recursive CTE Optimization
CREATE INDEX IF NOT EXISTS idx_companies_parent_id_composite ON companies(parent_id, id);

-- 2. Standard Inter-Branch Clearing Accounts (Due To / Due From)
-- Ensure default accounts 1990 and 2990 exist for corporate companies
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT id FROM companies WHERE parent_id IS NULL LOOP
        -- 1990: Due From Branches (Current Asset)
        INSERT INTO accounts (id, company_id, code, name, root_type, is_group, currency, is_active)
        VALUES (gen_random_uuid(), r.id, '1990', 'Due From Branches (Inter-Company Clearing Asset)', 'Asset', FALSE, 'BDT', TRUE)
        ON CONFLICT DO NOTHING;

        -- 2990: Due To Branches (Current Liability)
        INSERT INTO accounts (id, company_id, code, name, root_type, is_group, currency, is_active)
        VALUES (gen_random_uuid(), r.id, '2990', 'Due To Branches (Inter-Company Clearing Liability)', 'Liability', FALSE, 'BDT', TRUE)
        ON CONFLICT DO NOTHING;
    END LOOP;
END $$;

-- 3. Composite Index on gl_entries for Branch and Party Sub-ledger Lookups
CREATE INDEX IF NOT EXISTS idx_gl_entries_company_party_date 
ON gl_entries(company_id, party_id, posting_date) 
WHERE party_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_gl_entries_account_company 
ON gl_entries(account_id, company_id, posting_date);

-- 4. Sub-ledger Integrity Constraint Check Function
-- Enforces that Accounts Receivable (1020) and Accounts Payable (2001) MUST have party_id
CREATE OR REPLACE FUNCTION fn_validate_subledger_party()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.account_code IN ('1020', '2001') AND NEW.party_id IS NULL THEN
        RAISE EXCEPTION 'Tryton Sub-Ledger Invariant: Account % (Receivable/Payable) strictly requires party_id', NEW.account_code;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_validate_subledger_party ON gl_entries;
CREATE TRIGGER trg_validate_subledger_party
BEFORE INSERT ON gl_entries
FOR EACH ROW
EXECUTE FUNCTION fn_validate_subledger_party();
