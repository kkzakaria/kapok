-- Kapok Control Database Schema
-- This migration creates the foundational tables for the multi-tenant platform

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ===========================================================================
-- Tenants Table
-- ===========================================================================
-- Stores metadata for each tenant in the platform
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    schema_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT tenants_status_check CHECK (status IN ('active', 'provisioning', 'suspended', 'deleted', 'migrating'))
);

-- Index for filtering by status
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

-- Index for lookup by name
CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);

-- ===========================================================================
-- Casbin RBAC Rule Table
-- ===========================================================================
-- Stores role-based access control policies for the platform
CREATE TABLE IF NOT EXISTS casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(10),    -- Policy type: p (policy) or g (grouping/role)
    v0 VARCHAR(256),      -- Subject (user ID or role name)
    v1 VARCHAR(256),      -- Object (resource name)
    v2 VARCHAR(256),      -- Action (read, write, delete, etc.)
    v3 VARCHAR(256),      -- Tenant ID (for tenant-scoped permissions)
    v4 VARCHAR(256),      -- Optional: additional context
    v5 VARCHAR(256)       -- Optional: additional context
);

-- Composite index for efficient policy lookups
CREATE INDEX IF NOT EXISTS idx_casbin_rule ON casbin_rule(ptype, v0, v1, v2, v3);

-- ===========================================================================
-- Audit Log Table
-- ===========================================================================
-- Immutable audit trail for all critical operations
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,                    -- NULL for platform-level events
    user_id VARCHAR(256),              -- User who performed the action
    action VARCHAR(100) NOT NULL,      -- Action performed (e.g., 'tenant.create', 'user.login')
    resource VARCHAR(256),             -- Resource affected (e.g., 'tenant:123', 'user:abc')
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB,                    -- Additional context (IP address, user agent, etc.)
    
    CONSTRAINT audit_log_tenant_fk FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
);

-- Indexes for efficient audit log queries
CREATE INDEX IF NOT EXISTS idx_audit_log_tenant ON audit_log(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action, timestamp DESC);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_audit_log_metadata ON audit_log USING GIN (metadata);

-- ===========================================================================
-- Tenants Table Extensions (control-plane)
-- ===========================================================================
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug VARCHAR(100);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS isolation_level VARCHAR(20) DEFAULT 'schema';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS storage_used_bytes BIGINT DEFAULT 0;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS last_activity TIMESTAMP;

-- Epic 9: Hierarchy columns
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES tenants(id);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS hierarchy_level VARCHAR(20) DEFAULT 'organization';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS path TEXT DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_tenants_parent ON tenants(parent_id);
CREATE INDEX IF NOT EXISTS idx_tenants_path ON tenants(path);

-- Epic 9: DB-per-tenant columns
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS database_name VARCHAR(100) DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS database_host VARCHAR(256) DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS database_port INT DEFAULT 0;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS tier VARCHAR(20) DEFAULT 'standard';

-- ===========================================================================
-- Users Table (authentication)
-- ===========================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(256) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    roles TEXT NOT NULL DEFAULT 'user',
    tenant_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ===========================================================================
-- Trigger: Update tenants.updated_at
-- ===========================================================================
-- Automatically update the updated_at timestamp when a tenant record changes
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenants_updated_at_trigger
BEFORE UPDATE ON tenants
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- ===========================================================================
-- Initial Audit Log Entry
-- ===========================================================================
-- Record the creation of the control database
-- ===========================================================================
-- Tenant Quotas Table (Epic 9)
-- ===========================================================================
CREATE TABLE IF NOT EXISTS tenant_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id),
    max_storage_bytes BIGINT DEFAULT 0,
    max_connections INT DEFAULT 0,
    max_qps FLOAT DEFAULT 0,
    max_children INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ===========================================================================
-- Tenant Migrations Table (Epic 9)
-- ===========================================================================
CREATE TABLE IF NOT EXISTS tenant_migrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    from_isolation VARCHAR(20) NOT NULL,
    to_isolation VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    rollback_before TIMESTAMP,
    error_message TEXT DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tenant_migrations_tenant ON tenant_migrations(tenant_id);

-- ===========================================================================
-- Initial Audit Log Entry
-- ===========================================================================
-- Record the creation of the control database
INSERT INTO audit_log (action, resource, metadata)
VALUES (
    'database.initialize',
    'control_database',
    '{"version": "1.0.0", "description": "Control database schema initialized"}'::jsonb
);
