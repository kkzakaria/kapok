package migrationtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckNames(t *testing.T) {
	assert.Equal(t, "row_count", (&RowCountCheck{}).Name())
	assert.Equal(t, "checksum", (&ChecksumCheck{}).Name())
	assert.Equal(t, "schema_compare", (&SchemaCompareCheck{}).Name())
}
