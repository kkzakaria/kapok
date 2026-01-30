package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemStoreCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFilesystemStore(dir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test/backup.sql.gz"
	data := []byte("CREATE TABLE test (id INT);")

	// Upload
	err = store.Upload(ctx, key, bytes.NewReader(data))
	require.NoError(t, err)

	// Exists
	exists, err := store.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Download
	rc, err := store.Download(ctx, key)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	rc.Close()
	require.NoError(t, err)
	assert.Equal(t, data, got)

	// Delete
	err = store.Delete(ctx, key)
	require.NoError(t, err)

	exists, err = store.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFilesystemStoreDeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFilesystemStore(dir)
	require.NoError(t, err)

	err = store.Delete(context.Background(), "nonexistent")
	assert.NoError(t, err)
}

func TestFilesystemStoreExistsNonExistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFilesystemStore(dir)
	require.NoError(t, err)

	exists, err := store.Exists(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}
