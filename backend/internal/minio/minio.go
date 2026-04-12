// Package minio provides MinIO client initialization.
package minio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStore struct {
	client *minio.Client
	bucket string
}

// NewClient creates a new MinIO client.
func NewClient(ctx context.Context, endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("MinIO endpoint is required")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Verify connection
	exists, err := client.BucketExists(ctx, "docvault-documents")
	if err != nil {
		// Bucket might not exist yet, which is okay for initial setup
		slog.Warn("bucket check failed (may not exist yet)", "error", err)
	} else if exists {
		slog.Info("MinIO bucket exists", "bucket", "docvault-documents")
	}

	slog.Info("MinIO client initialized", "endpoint", endpoint)

	return client, nil
}

// EnsureBucket ensures the docvault bucket exists.
func EnsureBucket(ctx context.Context, client *minio.Client, bucketName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		slog.Info("bucket created", "bucket", bucketName)
	}

	return nil
}

func NewObjectStore(client *minio.Client, bucket string) *ObjectStore {
	return &ObjectStore{client: client, bucket: bucket}
}

func (s *ObjectStore) PutObject(ctx context.Context, object string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, object, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object to minio: %w", err)
	}
	return nil
}

func (s *ObjectStore) PresignGetObject(ctx context.Context, object string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, object, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("failed to presign minio object: %w", err)
	}
	return presignedURL.String(), nil
}
