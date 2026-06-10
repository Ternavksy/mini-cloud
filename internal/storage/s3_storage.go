package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client     *minio.Client
	bucketName string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucketName string) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	_, err = client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to minio: %w", err)
	}

	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
		} else {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *S3Storage) Save(id string, r io.Reader) error {
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucketName, id, r, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to save object: %w", err)
	}
	return nil
}

func (s *S3Storage) Get(id string) (io.ReadCloser, error) {
	ctx := context.Background()
	object, err := s.client.GetObject(ctx, s.bucketName, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}

func (s *S3Storage) Delete(id string) error {
	ctx := context.Background()
	err := s.client.RemoveObject(ctx, s.bucketName, id, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (s *S3Storage) List() ([]string, error) {
	ctx := context.Background()
	var files []string

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}
