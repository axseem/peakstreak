package storage

import (
	"context"
	"io"
)

// FileStorage defines the interface for blob storage operations.
type FileStorage interface {
	// Save stores a file and returns its public-facing URL.
	Save(ctx context.Context, file io.Reader, filename string) (string, error)
	// Delete removes a file given its public-facing URL.
	Delete(ctx context.Context, url string) error
}
