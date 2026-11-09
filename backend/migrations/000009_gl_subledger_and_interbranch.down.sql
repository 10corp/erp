-- 000009_gl_subledger_and_interbranch.down.sql
DROP TRIGGER IF EXISTS trg_validate_subledger_party ON gl_entries;
DROP FUNCTION IF EXISTS fn_validate_subledger_party();
DROP INDEX IF EXISTS idx_gl_entries_account_company;
DROP INDEX IF EXISTS idx_gl_entries_company_party_date;
DROP INDEX IF EXISTS idx_companies_parent_id_composite;
