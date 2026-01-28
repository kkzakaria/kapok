package rbac

import (
	"fmt"

	"github.com/rs/zerolog"
)

// Role represents a role in the system
type Role struct {
	Name        string
	Description string
	Permissions []Permission
}

// Permission represents a specific permission
type Permission struct {
	Object string
	Action string
}

// Default roles and their permissions
var (
	// RoleAdmin has full permissions
	RoleAdmin = Role{
		Name:        "admin",
		Description: "Administrator with full access",
		Permissions: []Permission{
			{Object: "*", Action: "*"},
		},
	}

	// RoleDeveloper has read/write permissions
	RoleDeveloper = Role{
		Name:        "developer",
		Description: "Developer with read and write access",
		Permissions: []Permission{
			{Object: "database", Action: "read"},
			{Object: "database", Action: "write"},
			{Object: "api", Action: "read"},
			{Object: "api", Action: "write"},
			{Object: "schema", Action: "read"},
			{Object: "schema", Action: "write"},
		},
	}

	// RoleViewer has read-only permissions
	RoleViewer = Role{
		Name:        "viewer",
		Description: "Viewer with read-only access",
		Permissions: []Permission{
			{Object: "database", Action: "read"},
			{Object: "api", Action: "read"},
			{Object: "schema", Action: "read"},
		},
	}
)

// DefaultRoles returns all default roles
func DefaultRoles() []Role {
	return []Role{RoleAdmin, RoleDeveloper, RoleViewer}
}

// BootstrapDefaultPolicies initializes default roles and permissions
func BootstrapDefaultPolicies(enforcer *Enforcer, tenantID string, logger zerolog.Logger) error {
	logger.Info().
		Str("tenant", tenantID).
		Msg("bootstrapping default RBAC policies")

	roles := DefaultRoles()

	for _, role := range roles {
		for _, perm := range role.Permissions {
			err := enforcer.AddPolicy(role.Name, perm.Object, perm.Action, tenantID)
			if err != nil {
				return fmt.Errorf("failed to add policy for role %s: %w", role.Name, err)
			}

			logger.Debug().
				Str("role", role.Name).
				Str("object", perm.Object).
				Str("action", perm.Action).
				Str("tenant", tenantID).
				Msg("policy added")
		}
	}

	// Add role hierarchy: admin > developer > viewer
	// This means admin inherits all developer permissions, and developer inherits all viewer permissions
	if err := enforcer.AddRoleForUser("admin", "developer"); err != nil {
		return fmt.Errorf("failed to add admin -> developer hierarchy: %w", err)
	}
	if err := enforcer.AddRoleForUser("developer", "viewer"); err != nil {
		return fmt.Errorf("failed to add developer -> viewer hierarchy: %w", err)
	}

	logger.Info().
		Str("tenant", tenantID).
		Int("roles", len(roles)).
		Msg("default RBAC policies bootstrapped successfully")

	return nil
}

// GetRoleByName retrieves a role by name
func GetRoleByName(name string) (Role, error) {
	for _, role := range DefaultRoles() {
		if role.Name == name {
			return role, nil
		}
	}
	return Role{}, fmt.Errorf("role not found: %s", name)
}

// ValidateRole checks if a role name is valid
func ValidateRole(name string) bool {
	for _, role := range DefaultRoles() {
		if role.Name == name {
			return true
		}
	}
	return false
}
