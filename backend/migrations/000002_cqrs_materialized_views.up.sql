-- 000002_cqrs_materialized_views.up.sql
-- CQRS Architecture: Blazing Fast Read Models for CEO Dashboards & Financial Reports

-- ============================================================================
-- 1. Account Balances Materialized View
-- Instant P&L and Balance Sheet aggregation across millions of partitioned rows
-- ============================================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_account_balances AS
SELECT
    a.company_id,
    a.id AS account_id,
    a.code AS account_code,
    a.name AS account_name,
    a.root_type,
    a.currency,
    COALESCE(SUM(g.debit_base), 0.000000) AS total_debit_base,
    COALESCE(SUM(g.credit_base), 0.000000) AS total_credit_base,
    CASE
        WHEN a.root_type IN ('Asset', 'Expense') THEN
            COALESCE(SUM(g.debit_base), 0.000000) - COALESCE(SUM(g.credit_base), 0.000000)
        ELSE
            COALESCE(SUM(g.credit_base), 0.000000) - COALESCE(SUM(g.debit_base), 0.000000)
    END AS net_balance_base,
    COUNT(g.id) AS total_transactions,
    MAX(g.posting_date) AS last_posting_date
FROM accounts a
LEFT JOIN gl_entries g ON a.id = g.account_id AND g.is_cancelled = FALSE
WHERE a.is_active = TRUE
GROUP BY a.company_id, a.id, a.code, a.name, a.root_type, a.currency;

-- Unique index required for REFRESH MATERIALIZED VIEW CONCURRENTLY
CREATE UNIQUE INDEX idx_mv_account_balances_unique ON mv_account_balances(account_id);
CREATE INDEX idx_mv_account_balances_company ON mv_account_balances(company_id, root_type);

-- ============================================================================
-- 2. CEO Real-Time Executive KPI Dashboard View
-- Aggregated Revenue, Expenses, Net Profit, Cash Position
-- ============================================================================
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_ceo_financial_kpis AS
SELECT
    company_id,
    COALESCE(SUM(CASE WHEN root_type = 'Income' THEN net_balance_base ELSE 0 END), 0.000000) AS total_revenue,
    COALESCE(SUM(CASE WHEN root_type = 'Expense' THEN net_balance_base ELSE 0 END), 0.000000) AS total_expense,
    (
        COALESCE(SUM(CASE WHEN root_type = 'Income' THEN net_balance_base ELSE 0 END), 0.000000) -
        COALESCE(SUM(CASE WHEN root_type = 'Expense' THEN net_balance_base ELSE 0 END), 0.000000)
    ) AS net_profit,
    COALESCE(SUM(CASE WHEN root_type = 'Asset' THEN net_balance_base ELSE 0 END), 0.000000) AS total_assets,
    COALESCE(SUM(CASE WHEN root_type = 'Liability' THEN net_balance_base ELSE 0 END), 0.000000) AS total_liabilities,
    NOW() AS calculated_at
FROM mv_account_balances
GROUP BY company_id;

CREATE UNIQUE INDEX idx_mv_ceo_kpis_company ON mv_ceo_financial_kpis(company_id);

-- ============================================================================
-- 3. Concurrent Refresh Stored Procedure
-- Can be triggered by Go Worker without blocking any reads
-- ============================================================================
CREATE OR REPLACE PROCEDURE sp_refresh_financial_kpis()
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_account_balances;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_ceo_financial_kpis;
END;
$$;
