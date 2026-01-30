package tenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinStrings(t *testing.T) {
	assert.Equal(t, "", joinStrings(nil, ", "))
	assert.Equal(t, "a", joinStrings([]string{"a"}, ", "))
	assert.Equal(t, "a, b, c", joinStrings([]string{"a", "b", "c"}, ", "))
}
