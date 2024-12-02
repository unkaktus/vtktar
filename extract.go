package vtktar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

// Extract extracts the VTKTAR archive into the destination
func Extract(destination, filename string) error {
	di, err := os.Stat(destination)
	if err != nil {
		return fmt.Errorf("stat destination: %w", err)
	}
	if !di.IsDir() {
		return fmt.Errorf("destination is not a directory")
	}
	file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	tr := tar.NewReader(file)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("iterating through the tar archive: %w", err)
		}
		r, err := zstd.NewReader(tr)
		if err != nil {
			return fmt.Errorf("create xz reader: %w", err)
		}
		destFilename := filepath.Join(destination, hdr.Name)
		destFile, err := os.OpenFile(destFilename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("open destination file: %w", err)
		}

		if _, err := io.Copy(destFile, r); err != nil {
			return fmt.Errorf("copy contents to the destination file: %w", err)
		}
	}
	return nil
}
