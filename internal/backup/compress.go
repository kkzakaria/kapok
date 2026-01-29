package backup

import (
	"compress/gzip"
	"fmt"
	"io"
)

// Compress gzip-compresses data from src and writes to dst.
func Compress(dst io.Writer, src io.Reader) error {
	gw := gzip.NewWriter(dst)
	if _, err := io.Copy(gw, src); err != nil {
		gw.Close()
		return fmt.Errorf("failed to compress: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("failed to finalize compression: %w", err)
	}
	return nil
}

// Decompress gzip-decompresses data from src and writes to dst.
func Decompress(dst io.Writer, src io.Reader) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	if _, err := io.Copy(dst, gr); err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}
	return nil
}
