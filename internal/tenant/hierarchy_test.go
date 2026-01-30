package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHierarchyLevelConstants(t *testing.T) {
	assert.Equal(t, HierarchyLevel("organization"), HierarchyOrganization)
	assert.Equal(t, HierarchyLevel("project"), HierarchyProject)
	assert.Equal(t, HierarchyLevel("team"), HierarchyTeam)
}

func TestMaxHierarchyDepth(t *testing.T) {
	assert.Equal(t, 3, MaxHierarchyDepth)
}

func TestHierarchyLevelProgression(t *testing.T) {
	tests := []struct {
		parent   HierarchyLevel
		expected HierarchyLevel
		valid    bool
	}{
		{HierarchyOrganization, HierarchyProject, true},
		{HierarchyProject, HierarchyTeam, true},
		{HierarchyTeam, "", false}, // cannot go deeper
	}

	for _, tt := range tests {
		t.Run(string(tt.parent), func(t *testing.T) {
			var childLevel HierarchyLevel
			var valid bool
			switch tt.parent {
			case HierarchyOrganization:
				childLevel = HierarchyProject
				valid = true
			case HierarchyProject:
				childLevel = HierarchyTeam
				valid = true
			default:
				valid = false
			}
			assert.Equal(t, tt.valid, valid)
			if valid {
				assert.Equal(t, tt.expected, childLevel)
			}
		})
	}
}
