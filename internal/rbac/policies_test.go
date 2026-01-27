package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRoles(t *testing.T) {
	roles := DefaultRoles()
	
	assert.Len(t, roles, 3, "should have 3 default roles")
	
	roleNames := make(map[string]bool)
	for _, role := range roles {
		roleNames[role.Name] = true
	}
	
	assert.True(t, roleNames["admin"])
	assert.True(t, roleNames["developer"])
	assert.True(t, roleNames["viewer"])
}

func TestRoleAdmin(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin.Name)
	assert.NotEmpty(t, RoleAdmin.Description)
	assert.NotEmpty(t, RoleAdmin.Permissions)
	
	// Admin should have wildcard permissions
	hasWildcard := false
	for _, perm := range RoleAdmin.Permissions {
		if perm.Object == "*" && perm.Action == "*" {
			hasWildcard = true
			break
		}
	}
	assert.True(t, hasWildcard, "admin should have wildcard permissions")
}

func TestRoleDeveloper(t *testing.T) {
	assert.Equal(t, "developer", RoleDeveloper.Name)
	assert.NotEmpty(t, RoleDeveloper.Description)
	assert.NotEmpty(t, RoleDeveloper.Permissions)
	
	// Developer should have read/write but not full wildcard
	hasRead := false
	hasWrite := false
	for _, perm := range RoleDeveloper.Permissions {
		if perm.Action == "read" {
			hasRead = true
		}
		if perm.Action == "write" {
			hasWrite = true
		}
	}
	assert.True(t, hasRead, "developer should have read permissions")
	assert.True(t, hasWrite, "developer should have write permissions")
}

func TestRoleViewer(t *testing.T) {
	assert.Equal(t, "viewer", RoleViewer.Name)
	assert.NotEmpty(t, RoleViewer.Description)
	assert.NotEmpty(t, RoleViewer.Permissions)
	
	// Viewer should only have read permissions
	for _, perm := range RoleViewer.Permissions {
		assert.Equal(t, "read", perm.Action, "viewer should only have read action")
	}
}

func TestGetRoleByName_Success(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		expected string
	}{
		{"get admin", "admin", "admin"},
		{"get developer", "developer", "developer"},
		{"get viewer", "viewer", "viewer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := GetRoleByName(tt.roleName)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, role.Name)
		})
	}
}

func TestGetRoleByName_NotFound(t *testing.T) {
	_, err := GetRoleByName("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role not found")
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		expected bool
	}{
		{"valid admin", "admin", true},
		{"valid developer", "developer", true},
		{"valid viewer", "viewer", true},
		{"invalid role", "superadmin", false},
		{"empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRole(tt.roleName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoleHierarchy(t *testing.T) {
	// Verify that role permissions follow hierarchy
	adminPerms := len(RoleAdmin.Permissions)
	devPerms := len(RoleDeveloper.Permissions)
	viewerPerms := len(RoleViewer.Permissions)
	
	// Admin has wildcard permission
	assert.Greater(t, adminPerms, 0, "admin should have at least one permission")
	assert.GreaterOrEqual(t, devPerms, viewerPerms, 
		"developer should have at least as many permissions as viewer")
}

func TestPermissionStructure(t *testing.T) {
	// Test that permissions have valid structure
	roles := DefaultRoles()
	
	for _, role := range roles {
		for _, perm := range role.Permissions {
			assert.NotEmpty(t, perm.Object, "permission object should not be empty for role %s", role.Name)
			assert.NotEmpty(t, perm.Action, "permission action should not be empty for role %s", role.Name)
		}
	}
}
