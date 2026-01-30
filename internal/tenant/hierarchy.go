package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// HierarchyManager manages parent-child tenant relationships
type HierarchyManager struct {
	db     *database.DB
	logger zerolog.Logger
}

// NewHierarchyManager creates a new HierarchyManager
func NewHierarchyManager(db *database.DB, logger zerolog.Logger) *HierarchyManager {
	return &HierarchyManager{
		db:     db,
		logger: logger,
	}
}

// GetChildren returns direct children of the given tenant
func (h *HierarchyManager) GetChildren(ctx context.Context, parentID string) ([]*Tenant, error) {
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE parent_id = $1 ORDER BY created_at", tenantSelectColumns)

	rows, err := h.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query children: %w", err)
	}
	defer rows.Close()

	var children []*Tenant
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child tenant: %w", err)
		}
		children = append(children, t)
	}
	return children, rows.Err()
}

// GetAncestors returns all ancestors of a tenant from immediate parent to root
func (h *HierarchyManager) GetAncestors(ctx context.Context, tenantID string) ([]*Tenant, error) {
	// Use the path to find ancestors
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE id = $1", tenantSelectColumns)
	t, err := scanTenant(h.db.QueryRowContext(ctx, query, tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if t.Path == "" || t.ParentID == nil {
		return nil, nil
	}

	// Parse path to get ancestor IDs (path format: /rootID/parentID/selfID)
	parts := strings.Split(strings.Trim(t.Path, "/"), "/")
	if len(parts) <= 1 {
		return nil, nil
	}

	// Exclude self from ancestors
	ancestorIDs := parts[:len(parts)-1]
	if len(ancestorIDs) == 0 {
		return nil, nil
	}

	// Build query for ancestors
	placeholders := make([]string, len(ancestorIDs))
	args := make([]interface{}, len(ancestorIDs))
	for i, id := range ancestorIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	ancestorQuery := fmt.Sprintf("SELECT %s FROM tenants WHERE id IN (%s) ORDER BY path",
		tenantSelectColumns, strings.Join(placeholders, ","))

	rows, err := h.db.QueryContext(ctx, ancestorQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ancestors: %w", err)
	}
	defer rows.Close()

	var ancestors []*Tenant
	for rows.Next() {
		a, err := scanTenant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ancestor: %w", err)
		}
		ancestors = append(ancestors, a)
	}
	return ancestors, rows.Err()
}

// GetSubtree returns all descendants of a tenant (uses path prefix matching)
func (h *HierarchyManager) GetSubtree(ctx context.Context, tenantID string) ([]*Tenant, error) {
	// First get the tenant to know its path
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE id = $1", tenantSelectColumns)
	t, err := scanTenant(h.db.QueryRowContext(ctx, query, tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Find all tenants whose path starts with this tenant's path (excluding self)
	subtreeQuery := fmt.Sprintf("SELECT %s FROM tenants WHERE path LIKE $1 AND id != $2 ORDER BY path",
		tenantSelectColumns)

	rows, err := h.db.QueryContext(ctx, subtreeQuery, t.Path+"/%", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subtree: %w", err)
	}
	defer rows.Close()

	var descendants []*Tenant
	for rows.Next() {
		d, err := scanTenant(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan descendant: %w", err)
		}
		descendants = append(descendants, d)
	}
	return descendants, rows.Err()
}

// ValidateDepth checks whether a new child can be created under the given parent
func (h *HierarchyManager) ValidateDepth(ctx context.Context, parentID string) error {
	query := fmt.Sprintf("SELECT %s FROM tenants WHERE id = $1", tenantSelectColumns)
	parent, err := scanTenant(h.db.QueryRowContext(ctx, query, parentID))
	if err != nil {
		return fmt.Errorf("failed to get parent tenant: %w", err)
	}

	depth := len(strings.Split(strings.Trim(parent.Path, "/"), "/"))
	if depth >= MaxHierarchyDepth {
		return fmt.Errorf("maximum hierarchy depth (%d) reached", MaxHierarchyDepth)
	}
	return nil
}

// CountChildren returns the number of direct children for a tenant
func (h *HierarchyManager) CountChildren(ctx context.Context, parentID string) (int, error) {
	var count int
	err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tenants WHERE parent_id = $1", parentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count children: %w", err)
	}
	return count, nil
}
