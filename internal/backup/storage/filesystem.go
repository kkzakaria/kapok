package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FilesystemStore implements Store using the local filesystem.
type FilesystemStore struct {
	basePath string
}

// NewFilesystemStore creates a new filesystem-backed store.
func NewFilesystemStore(basePath string) (*FilesystemStore, error) {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storage path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &FilesystemStore{basePath: abs}, nil
}

func (s *FilesystemStore) fullPath(key string) (string, error) {
	p := filepath.Join(s.basePath, key)
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	// Ensure the resolved path stays within basePath to prevent path traversal.
	if !strings.HasPrefix(abs, s.basePath+string(filepath.Separator)) && abs != s.basePath {
		return "", fmt.Errorf("path traversal detected: %s escapes base %s", key, s.basePath)
	}
	return abs, nil
}

func (s *FilesystemStore) Upload(_ context.Context, key string, r io.Reader) error {
	p, err := s.fullPath(key)
	if err != nil {
		return err
	}
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
	p, err := s.fullPath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}

func (s *FilesystemStore) Delete(_ context.Context, key string) error {
	p, err := s.fullPath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *FilesystemStore) Exists(_ context.Context, key string) (bool, error) {
	p, err := s.fullPath(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
