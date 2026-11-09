-- 000007_core_user_management.up.sql
-- Core User & Identity Management Subsystem (Tryton-inspired Architecture)
-- Multi-Company Tenancy, Hierarchical Roles, Bitmask Permissions & Audit Trail

-- ============================================================================
-- 1. Users Core Table (res.user)
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_superuser BOOLEAN NOT NULL DEFAULT FALSE, -- Tryton administrator (access to all models)
    current_company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    bitmask_permissions SMALLINT NOT NULL DEFAULT 1 CHECK (bitmask_permissions BETWEEN 0 AND 15),
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT uq_users_email UNIQUE (email)
);

CREATE INDEX idx_users_username_lower ON users(LOWER(username));
CREATE INDEX idx_users_email_lower ON users(LOWER(email));
CREATE INDEX idx_users_active ON users(is_active);

-- ============================================================================
-- 2. Multi-Company Tenancy (Tryton: res.user-company.company)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_companies (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, company_id)
);

CREATE INDEX idx_user_companies_company ON user_companies(company_id);

-- ============================================================================
-- 3. Hierarchical Roles & Groups (Tryton: res.group / user_role)
-- ============================================================================
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES roles(id) ON DELETE RESTRICT, -- Role Tree inheritance
    bitmask_permissions SMALLINT NOT NULL DEFAULT 1 CHECK (bitmask_permissions BETWEEN 0 AND 15),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_roles_parent ON roles(parent_id);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- ============================================================================
-- 4. Active Sessions (Token Management & Revocation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash of bearer/JWT session token
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_lookup ON user_sessions(token_hash, revoked_at, expires_at);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);

-- ============================================================================
-- 5. Security & Authentication Audit Log (GAAP/IFRS Immutability)
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL, -- LOGIN_SUCCESS, LOGIN_FAILED, ROLE_ASSIGNED, etc.
    ip_address INET,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_audit_logs_user_date ON user_audit_logs(user_id, created_at DESC);
CREATE INDEX idx_user_audit_logs_action ON user_audit_logs(action);

-- ============================================================================
-- 6. Role Hierarchy CTE View (Sub-Microsecond Bitmask Resolution)
-- ============================================================================
CREATE OR REPLACE VIEW vw_user_effective_permissions AS
WITH RECURSIVE role_tree AS (
    -- Direct assigned roles
    SELECT 
        ur.user_id,
        r.id AS role_id,
        r.code AS role_code,
        r.bitmask_permissions
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id AND r.is_active = TRUE
    
    UNION
    
    -- Inherited roles from parent
    SELECT 
        rt.user_id,
        p.id AS role_id,
        p.code AS role_code,
        p.bitmask_permissions
    FROM role_tree rt
    JOIN roles child ON rt.role_id = child.id
    JOIN roles p ON child.parent_id = p.id AND p.is_active = TRUE
)
SELECT 
    u.id AS user_id,
    u.username,
    u.email,
    u.is_superuser,
    CASE 
        WHEN u.is_superuser THEN 15 -- Superuser has all bitmasks (1 | 2 | 4 | 8)
        ELSE (u.bitmask_permissions | COALESCE(BIT_OR(rt.bitmask_permissions), 0))
    END AS effective_permissions
FROM users u
LEFT JOIN role_tree rt ON u.id = rt.user_id
WHERE u.is_active = TRUE
GROUP BY u.id, u.username, u.email, u.is_superuser, u.bitmask_permissions;
