package migrationtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatencyCheckName(t *testing.T) {
	c := &LatencyCheck{MaxLatencyMS: 100}
	assert.Equal(t, "latency", c.Name())
}

func TestLatencyCheckDefaults(t *testing.T) {
	c := &LatencyCheck{}
	assert.Equal(t, int64(0), c.MaxLatencyMS) // defaults to 100ms in Run
}
