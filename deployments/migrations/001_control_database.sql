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
    
    CONSTRAINT tenants_status_check CHECK (status IN ('active', 'provisioning', 'suspended', 'deleted'))
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
INSERT INTO audit_log (action, resource, metadata)
VALUES (
    'database.initialize',
    'control_database',
    '{"version": "1.0.0", "description": "Control database schema initialized"}'::jsonb
);
