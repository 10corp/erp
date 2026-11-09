-- 000008_core_party_company.down.sql
DROP VIEW IF EXISTS vw_company_hierarchy;
ALTER TABLE users DROP COLUMN IF EXISTS employee_id;
ALTER TABLE users DROP COLUMN IF EXISTS main_company_id;
DROP TABLE IF EXISTS company_employee CASCADE;
DROP VIEW IF EXISTS company_company;
ALTER TABLE companies DROP COLUMN IF EXISTS footer;
ALTER TABLE companies DROP COLUMN IF EXISTS header;
ALTER TABLE companies DROP COLUMN IF EXISTS branch_code;
ALTER TABLE companies DROP COLUMN IF EXISTS is_branch;
ALTER TABLE companies DROP COLUMN IF EXISTS timezone;
ALTER TABLE companies DROP COLUMN IF EXISTS party_id;
ALTER TABLE companies DROP COLUMN IF EXISTS parent_id;
DROP TABLE IF EXISTS party_contact_mechanism CASCADE;
DROP TABLE IF EXISTS party_address CASCADE;
DROP TABLE IF EXISTS party_identifier CASCADE;
DROP VIEW IF EXISTS parties;
DROP TABLE IF EXISTS party_party CASCADE;
