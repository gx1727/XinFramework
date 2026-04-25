-- ============================================
-- Search Performance Indexes
-- ILIKE fuzzy search optimization with gin_trgm
-- ============================================

SET client_encoding = 'UTF8';

-- Enable pg_trgm extension for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Users table indexes for ILIKE search on code, real_name, phone
CREATE INDEX IF NOT EXISTS idx_users_code_trgm ON users USING gin (code gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_real_name_trgm ON users USING gin (real_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_phone_trgm ON users USING gin (phone gin_trgm_ops);

-- Tenants table indexes for ILIKE search on name, code
CREATE INDEX IF NOT EXISTS idx_tenants_name_trgm ON tenants USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tenants_code_trgm ON tenants USING gin (code gin_trgm_ops);

-- auth_sessions composite index for validate query
CREATE INDEX IF NOT EXISTS idx_auth_sessions_session_expires
    ON auth_sessions (session_id, expires_at);

-- menus table composite index for GetUserMenus query
CREATE INDEX IF NOT EXISTS idx_menus_tenant_active
    ON menus (tenant_id) WHERE is_deleted = FALSE;

-- user_roles composite index for GetUserMenus/GetUserResources queries
CREATE INDEX IF NOT EXISTS idx_user_roles_user_active
    ON user_roles (user_id) WHERE is_deleted = FALSE;

-- permissions composite index for resource lookup
CREATE INDEX IF NOT EXISTS idx_permissions_role_resource
    ON permissions (role_id, resource_type, resource_code);
