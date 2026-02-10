package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/nelfander/Playingfield/internal/domain"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string // The base URL used to access files (e.g. http://localhost:9000/pf-uploads)
	logger     *slog.Logger
}

// NewS3Storage creates a new storage provider
func NewS3Storage(client *s3.Client, bucketName, publicURL string, logger *slog.Logger) *S3Storage {
	// create "child" logger that automatically adds the bucket name to EVERY log
	childLogger := logger.With("bucket", bucketName)
	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
		logger:     childLogger,
	}
}

// content is io.reader so that the data passes through without loading it all in memory
func (s *S3Storage) UploadFile(ctx context.Context, fileName string, content io.Reader) (*domain.UploadResult, error) {
	// generate a unique key so users don't overwrite each other's files
	key := fmt.Sprintf("%s-%s", uuid.New().String(), fileName)

	// upload the file to the bucket
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   content,
	})
	if err != nil {
		s.logger.Error("upload failed", "error", err)
		return nil, err
	}

	// construct the result
	return &domain.UploadResult{
		Key: key,
		URL: fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, key),
	}, nil
}

func (s *S3Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}
