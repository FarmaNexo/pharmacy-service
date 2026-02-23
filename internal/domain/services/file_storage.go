// internal/domain/services/file_storage.go
package services

import (
	"context"
	"io"
)

type FileStorage interface {
	Upload(ctx context.Context, bucket, key string, reader io.Reader, contentType string) (string, error)
	Delete(ctx context.Context, bucket, key string) error
}
