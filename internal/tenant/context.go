package tenant

import (
	"context"
	"fmt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// tenantIDKey is the context key for tenant ID
	tenantIDKey contextKey = "tenant_id"
	// tenantKey is the context key for full tenant object
	tenantKey contextKey = "tenant"
)

// WithTenantID adds a tenant ID to the context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// GetTenantID retrieves the tenant ID from the context
func GetTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(tenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}

// MustGetTenantID retrieves the tenant ID from context or panics
// Use this only in contexts where tenant ID is guaranteed to exist
func MustGetTenantID(ctx context.Context) string {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		panic(fmt.Sprintf("tenant ID not found in context: %v", err))
	}
	return tenantID
}

// WithTenant adds a full tenant object to the context
func WithTenant(ctx context.Context, tenant *Tenant) context.Context {
	ctx = context.WithValue(ctx, tenantKey, tenant)
	// Also set the tenant ID for convenience
	if tenant != nil {
		ctx = WithTenantID(ctx, tenant.ID)
	}
	return ctx
}

// GetTenant retrieves the full tenant object from the context
func GetTenant(ctx context.Context) (*Tenant, error) {
	tenant, ok := ctx.Value(tenantKey).(*Tenant)
	if !ok || tenant == nil {
		return nil, fmt.Errorf("tenant not found in context")
	}
	return tenant, nil
}

// HasTenantID checks if the context contains a tenant ID
func HasTenantID(ctx context.Context) bool {
	_, err := GetTenantID(ctx)
	return err == nil
}
