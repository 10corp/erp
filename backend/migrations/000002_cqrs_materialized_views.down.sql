-- 000002_cqrs_materialized_views.down.sql
DROP PROCEDURE IF EXISTS sp_refresh_financial_kpis();
DROP MATERIALIZED VIEW IF EXISTS mv_ceo_financial_kpis;
DROP MATERIALIZED VIEW IF EXISTS mv_account_balances;
