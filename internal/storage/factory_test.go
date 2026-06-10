package storage

import "testing"

func TestNewStorageDefaultsToLocalStorage(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "")
	t.Setenv("LOCAL_STORAGE_PATH", t.TempDir())

	store, err := NewStorage()
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}
	if _, ok := store.(*LocalStorage); !ok {
		t.Fatalf("NewStorage() type = %T, want *LocalStorage", store)
	}
}

func TestNewStorageReturnsErrorForMissingS3Config(t *testing.T) {
	t.Setenv("STORAGE_TYPE", "s3")
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("S3_ACCESS_KEY", "")
	t.Setenv("S3_SECRET_KEY", "")
	t.Setenv("S3_BUCKET_NAME", "")

	if _, err := NewStorage(); err == nil {
		t.Fatal("NewStorage() error = nil, want error")
	}
}
