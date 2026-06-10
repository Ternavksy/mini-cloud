package storage

import (
	"fmt"
	"os"
)

func NewStorage() (Storage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	switch storageType {
	case "s3":
		endpoint := os.Getenv("S3_ENDPOINT")
		accessKey := os.Getenv("S3_ACCESS_KEY")
		secretKey := os.Getenv("S3_SECRET_KEY")
		bucketName := os.Getenv("S3_BUCKET_NAME")

		if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
			return nil, fmt.Errorf("missing S3 configuration: endpoint, access_key, secret_key, or bucket_name")
		}

		return NewS3Storage(endpoint, accessKey, secretKey, bucketName)

	case "local":
		fallthrough
	default:
		basePath := os.Getenv("LOCAL_STORAGE_PATH")
		if basePath == "" {
			basePath = "./storage"
		}
		return New(basePath), nil
	}
}
