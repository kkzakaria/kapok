package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackupValidate(t *testing.T) {
	tests := []struct {
		name    string
		backup  Backup
		wantErr string
	}{
		{
			name:    "missing tenant_id",
			backup:  Backup{SchemaName: "tenant_x"},
			wantErr: "tenant_id is required",
		},
		{
			name:    "missing schema_name",
			backup:  Backup{TenantID: "abc"},
			wantErr: "schema_name is required",
		},
		{
			name:    "invalid status",
			backup:  Backup{TenantID: "abc", SchemaName: "t", Status: "bogus"},
			wantErr: "invalid status",
		},
		{
			name:    "invalid type",
			backup:  Backup{TenantID: "abc", SchemaName: "t", Type: "bogus"},
			wantErr: "invalid type",
		},
		{
			name:    "invalid trigger",
			backup:  Backup{TenantID: "abc", SchemaName: "t", Trigger: "bogus"},
			wantErr: "invalid trigger",
		},
		{
			name:   "valid minimal",
			backup: Backup{TenantID: "abc", SchemaName: "t"},
		},
		{
			name:   "valid full",
			backup: Backup{TenantID: "abc", SchemaName: "t", Status: StatusCompleted, Type: TypeFull, Trigger: TriggerScheduled},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.backup.Validate()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
