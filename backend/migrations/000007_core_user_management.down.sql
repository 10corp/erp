-- 000007_core_user_management.down.sql
DROP VIEW IF EXISTS vw_user_effective_permissions;
DROP TABLE IF EXISTS user_audit_logs CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS user_companies CASCADE;
DROP TABLE IF EXISTS users CASCADE;
