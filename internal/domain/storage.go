package domain

import (
	"context"
	"io"
)

// UploadResult represents the data we get back after a successful upload
type UploadResult struct {
	URL      string // The link to access the file
	Key      string // The unique name inside the bucket
	Size     int64
	MimeType string
}

type StorageProvider interface {
	UploadFile(ctx context.Context, fileName string, content io.Reader) (*UploadResult, error)
	DeleteFile(ctx context.Context, key string) error
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, string, error)
}
