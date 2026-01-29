package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FilesystemStore implements Store using the local filesystem.
type FilesystemStore struct {
	basePath string
}

// NewFilesystemStore creates a new filesystem-backed store.
func NewFilesystemStore(basePath string) (*FilesystemStore, error) {
	if err := os.MkdirAll(basePath, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &FilesystemStore{basePath: basePath}, nil
}

func (s *FilesystemStore) fullPath(key string) string {
	return filepath.Join(s.basePath, key)
}

func (s *FilesystemStore) Upload(_ context.Context, key string, r io.Reader) error {
	p := s.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return fmt.Errorf("failed to create parent dirs: %w", err)
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (s *FilesystemStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(s.fullPath(key))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}

func (s *FilesystemStore) Delete(_ context.Context, key string) error {
	if err := os.Remove(s.fullPath(key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *FilesystemStore) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(s.fullPath(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
