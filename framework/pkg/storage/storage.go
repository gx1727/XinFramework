package storage

import (
	"context"
	"io"
)

// Storage defines the interface for object storage operations
type Storage interface {
	Put(ctx context.Context, key string, file io.Reader) (string, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)
}
