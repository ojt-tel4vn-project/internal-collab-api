package storage

import (
	"context"
	"io"
)

type StorageService interface {
	UploadFile(ctx context.Context, path string, file io.Reader) (string, error)
}
