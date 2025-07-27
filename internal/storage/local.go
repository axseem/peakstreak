package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements the FileStorage interface for the local disk.
type LocalStorage struct {
	basePath  string // The absolute path on disk
	publicURL string // The public URL prefix
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(basePath, publicURL string) *LocalStorage {
	return &LocalStorage{
		basePath:  basePath,
		publicURL: publicURL,
	}
}

// Save stores a file to the local disk.
func (s *LocalStorage) Save(ctx context.Context, file io.Reader, filename string) (string, error) {
	// Ensure the base directory exists.
	if err := os.MkdirAll(s.basePath, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create storage directory: %w", err)
	}

	filePath := filepath.Join(s.basePath, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("could not create file on server: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("could not copy file contents: %w", err)
	}

	// Return the public-facing URL.
	return s.publicURL + "/" + filename, nil
}

// Delete removes a file from the local disk based on its public URL.
func (s *LocalStorage) Delete(ctx context.Context, url string) error {
	if !strings.HasPrefix(url, s.publicURL) {
		return fmt.Errorf("URL '%s' does not match public URL prefix '%s'", url, s.publicURL)
	}

	// Extract the filename from the URL.
	filename := strings.TrimPrefix(url, s.publicURL)
	filePath := filepath.Join(s.basePath, filename)

	// Check if the file exists before trying to remove it.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, which is fine. The goal is for it to be gone.
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
