-- 000001_init_accounting_schema.down.sql
DROP TABLE IF EXISTS background_jobs CASCADE;
DROP TYPE IF EXISTS job_status CASCADE;

DROP TRIGGER IF EXISTS trg_gl_immutability ON gl_entries CASCADE;
DROP FUNCTION IF EXISTS fn_enforce_gl_immutability CASCADE;

DROP TABLE IF EXISTS gl_entries CASCADE;
DROP TABLE IF EXISTS journal_entries CASCADE;
DROP TYPE IF EXISTS doc_status CASCADE;

DROP TABLE IF EXISTS accounts CASCADE;
DROP TYPE IF EXISTS account_root_type CASCADE;

DROP TABLE IF EXISTS companies CASCADE;
