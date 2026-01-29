package backup

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"short", "hello"},
		{"repetitive", "abcabcabcabcabcabcabcabcabcabc"},
		{"sql-like", "CREATE TABLE foo (id INT); INSERT INTO foo VALUES (1);"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var compressed bytes.Buffer
			err := Compress(&compressed, bytes.NewReader([]byte(tt.data)))
			require.NoError(t, err)

			var decompressed bytes.Buffer
			err = Decompress(&decompressed, bytes.NewReader(compressed.Bytes()))
			require.NoError(t, err)

			assert.Equal(t, tt.data, decompressed.String())
		})
	}
}

func TestCompressReducesSize(t *testing.T) {
	// Highly compressible data
	data := bytes.Repeat([]byte("AAAA"), 1000)

	var compressed bytes.Buffer
	require.NoError(t, Compress(&compressed, bytes.NewReader(data)))

	assert.Less(t, compressed.Len(), len(data))
}

func TestDecompressInvalidData(t *testing.T) {
	var buf bytes.Buffer
	err := Decompress(&buf, bytes.NewReader([]byte("not gzip data")))
	assert.Error(t, err)
}
