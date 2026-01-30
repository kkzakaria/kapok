package tenant

import (
	"context"
	"testing"

	"github.com/kapok/kapok/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrategyRegistry(t *testing.T) {
	reg := NewStrategyRegistry()

	// Register should not panic
	assert.NotPanics(t, func() {
		reg.Register(&mockStrategy{typ: "schema"})
		reg.Register(&mockStrategy{typ: "database"})
	})

	// Get existing
	s, err := reg.Get("schema")
	require.NoError(t, err)
	assert.Equal(t, "schema", s.Type())

	// Get non-existing
	_, err = reg.Get("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown isolation strategy")
}

type mockStrategy struct {
	typ string
}

func (m *mockStrategy) Provision(_ context.Context, _ *Tenant) error   { return nil }
func (m *mockStrategy) Deprovision(_ context.Context, _ *Tenant) error { return nil }
func (m *mockStrategy) GetConnection(_ context.Context, _ *Tenant) (*database.DB, error) {
	return nil, nil
}
func (m *mockStrategy) Type() string { return m.typ }
