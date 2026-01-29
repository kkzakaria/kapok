package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracingProvider_Shutdown(t *testing.T) {
	// Test shutdown with nil provider
	tp := &TracingProvider{}
	err := tp.Shutdown(context.Background())
	assert.NoError(t, err)
}
