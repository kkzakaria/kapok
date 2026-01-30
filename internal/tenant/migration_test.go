package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationStatusConstants(t *testing.T) {
	assert.Equal(t, MigrationStatus("pending"), MigrationPending)
	assert.Equal(t, MigrationStatus("replicating"), MigrationReplicating)
	assert.Equal(t, MigrationStatus("verifying"), MigrationVerifying)
	assert.Equal(t, MigrationStatus("cutover"), MigrationCutover)
	assert.Equal(t, MigrationStatus("completed"), MigrationCompleted)
	assert.Equal(t, MigrationStatus("rolled_back"), MigrationRolledBack)
	assert.Equal(t, MigrationStatus("failed"), MigrationFailed)
}

func TestMigrationRecordStruct(t *testing.T) {
	record := MigrationRecord{
		ID:            "mig-1",
		TenantID:      "tenant-1",
		FromIsolation: "schema",
		ToIsolation:   "database",
		Status:        MigrationPending,
	}

	assert.Equal(t, "mig-1", record.ID)
	assert.Equal(t, "schema", record.FromIsolation)
	assert.Equal(t, MigrationPending, record.Status)
}
