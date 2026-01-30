package tenant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecisionEngineGetPendingApprovals(t *testing.T) {
	// Test that empty engine returns empty map
	engine := &DecisionEngine{
		pendingApprovals: make(map[string]string),
		cooldowns:        make(map[string]time.Time),
	}

	approvals := engine.GetPendingApprovals()
	assert.Empty(t, approvals)

	// Add pending approval
	engine.pendingApprovals["tenant-1"] = "database"
	approvals = engine.GetPendingApprovals()
	assert.Len(t, approvals, 1)
	assert.Equal(t, "database", approvals["tenant-1"])

	// Verify it returns a copy
	approvals["tenant-2"] = "schema"
	assert.Len(t, engine.pendingApprovals, 1)
}
