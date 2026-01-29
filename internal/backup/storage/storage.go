package storage

import (
	"context"
	"io"
)

// Store defines the interface for backup storage backends.
type Store interface {
	// Upload stores data from r at the given key.
	Upload(ctx context.Context, key string, r io.Reader) error

	// Download returns a ReadCloser for the data at the given key.
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the object at the given key.
	Delete(ctx context.Context, key string) error

	// Exists checks whether an object exists at the given key.
	Exists(ctx context.Context, key string) (bool, error)
}
