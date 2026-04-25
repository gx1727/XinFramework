-- ============================================
-- Search Performance Indexes
-- ILIKE fuzzy search optimization with gin_trgm
-- Note: Skips user/tenant indexes if user lacks table ownership
-- ============================================

SET client_encoding = 'UTF8';

-- Enable pg_trgm extension for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Users table indexes (skip if no privileges - table may be owned by infra)
DO $$
BEGIN
    CREATE INDEX IF NOT EXISTS idx_users_code_trgm ON users USING gin (code gin_trgm_ops);
EXCEPTION WHEN insufficient_privileGE THEN
    RAISE NOTICE 'Skipped idx_users_code_trgm: no privileges on users table';
END $$;

DO $$
BEGIN
    CREATE INDEX IF NOT EXISTS idx_users_real_name_trgm ON users USING gin (real_name gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_users_real_name_trgm: no privileges on users table';
END $$;

DO $$
BEGIN
    CREATE INDEX IF NOT EXISTS idx_users_phone_trgm ON users USING gin (phone gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_users_phone_trgm: no privileges on users table';
END $$;

-- Tenants table indexes
DO $$
BEGIN
    CREATE INDEX IF NOT EXISTS idx_tenants_name_trgm ON tenants USING gin (name gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_tenants_name_trgm: no privileges on tenants table';
END $$;

DO $$
BEGIN
    CREATE INDEX IF NOT EXISTS idx_tenants_code_trgm ON tenants USING gin (code gin_trgm_ops);
EXCEPTION WHEN insufficient_privilege THEN
    RAISE NOTICE 'Skipped idx_tenants_code_trgm: no privileges on tenants table';
END $$;

-- Framework-owned tables - always create
CREATE INDEX IF NOT EXISTS idx_auth_sessions_session_expires
    ON auth_sessions (session_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_menus_tenant_active
    ON menus (tenant_id) WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_user_roles_user_active
    ON user_roles (user_id) WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_permissions_role_resource
    ON permissions (role_id, resource_type, resource_code);
