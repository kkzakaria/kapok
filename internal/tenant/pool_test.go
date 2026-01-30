package tenant

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestPoolManager(t *testing.T) {
	logger := zerolog.Nop()
	pm := NewPoolManager(logger)

	assert.Equal(t, 0, pm.Count())

	// Get non-existent
	_, ok := pm.Get("tenant-1")
	assert.False(t, ok)

	// Remove non-existent (should not panic)
	assert.NotPanics(t, func() {
		pm.Remove("tenant-1")
	})

	// CloseAll on empty
	assert.NotPanics(t, func() {
		pm.CloseAll()
	})
}
