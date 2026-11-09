-- 000010_idempiere_tenancy_and_erpnext_hr_payroll_bom.down.sql
DROP TABLE IF EXISTS manufacturing_work_orders CASCADE;
DROP TABLE IF EXISTS manufacturing_bom_items CASCADE;
DROP TABLE IF EXISTS manufacturing_boms CASCADE;

DROP TABLE IF EXISTS hr_salary_slips CASCADE;
DROP TABLE IF EXISTS hr_salary_structures CASCADE;
DROP TABLE IF EXISTS hr_leave_applications CASCADE;
DROP TABLE IF EXISTS hr_leave_types CASCADE;
DROP TABLE IF EXISTS hr_designations CASCADE;
DROP TABLE IF EXISTS hr_departments CASCADE;

DROP VIEW IF EXISTS stock_move CASCADE;
DROP VIEW IF EXISTS stock_location CASCADE;
DROP VIEW IF EXISTS account_move_line CASCADE;
DROP VIEW IF EXISTS account_move CASCADE;
DROP VIEW IF EXISTS account_account CASCADE;

ALTER TABLE stock_ledger_entries DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE gl_entries DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE warehouses DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE party_party DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE companies DROP COLUMN IF EXISTS tenant_id;

DROP TABLE IF EXISTS tenants CASCADE;
