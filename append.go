package vtktar

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/klauspost/compress/zstd"
)

// Append appends VTK files to the destination VTKTAR archive
func Append(destFilename string, filenames []string) error {
	file, err := os.OpenFile(destFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer file.Close()

	w := tar.NewWriter(file)
	buffer := &bytes.Buffer{}

	for _, filename := range filenames {
		fi, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("stat %s: %w", filename, err)
		}
		log.Printf("writing %s", fi.Name())

		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}

		// Fill the buffer with compressed file contents
		zw, err := zstd.NewWriter(buffer)
		if err != nil {
			return fmt.Errorf("create xz writer: %w", err)
		}
		_, err = io.Copy(zw, f)
		if err != nil {
			return fmt.Errorf("copy data to xz writer: %w", err)
		}
		zw.Close()

		// Write header
		header := &tar.Header{
			Name: fi.Name(),
			Size: int64(buffer.Len()),
			Mode: 0600,
		}
		if err := w.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header: %w", err)
		}
		// Write compressed data to the vtktar
		_, err = io.Copy(w, buffer)
		buffer.Reset()
		if err == io.EOF {
			continue
		}
		if err != nil {
			return fmt.Errorf("copy data to vtktar: %w", err)
		}
	}
	// Flush the tar writer but don't write the footer
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing tar writer: %w", err)
	}
	return nil
}
