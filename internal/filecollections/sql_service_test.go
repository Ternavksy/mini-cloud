package filecollections

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"mini-cloud/internal/database"
	"mini-cloud/internal/models"
)

func TestSQLServiceCreateListGetAddRemoveAndDelete(t *testing.T) {
	service := newTestSQLService(t)

	created, err := service.Create("album", []models.FileID{"file-1", "file-2", "file-1"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name != "album" {
		t.Fatalf("Create() name = %q, want album", created.Name)
	}
	if len(created.FileIDs) != 2 {
		t.Fatalf("Create() file ids = %v, want 2 unique ids", created.FileIDs)
	}

	listed, err := service.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("List() count = %d, want 1", len(listed))
	}

	got, err := service.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != created.ID || len(got.FileIDs) != 2 {
		t.Fatalf("Get() = %+v, want collection with two files", got)
	}

	updated, err := service.AddFile(created.ID, "file-3")
	if err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if len(updated.FileIDs) != 3 {
		t.Fatalf("AddFile() file ids = %v, want 3", updated.FileIDs)
	}

	updated, err = service.RemoveFile(created.ID, "file-1")
	if err != nil {
		t.Fatalf("RemoveFile() error = %v", err)
	}
	if len(updated.FileIDs) != 2 {
		t.Fatalf("RemoveFile() file ids = %v, want 2", updated.FileIDs)
	}

	if err := service.Delete(created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(created.ID); !errors.Is(err, ErrCollectionNotFound) {
		t.Fatalf("Get() after delete error = %v, want ErrCollectionNotFound", err)
	}
}

func TestSQLServiceValidatesInput(t *testing.T) {
	service := newTestSQLService(t)

	if _, err := service.Create("", nil); err == nil {
		t.Fatal("Create() error = nil, want error")
	}
	if _, err := service.AddFile("missing", ""); err == nil {
		t.Fatal("AddFile() blank file id error = nil, want error")
	}
	if err := service.Delete("missing"); !errors.Is(err, ErrCollectionNotFound) {
		t.Fatalf("Delete() error = %v, want ErrCollectionNotFound", err)
	}
}

func newTestSQLService(t *testing.T) *SQLService {
	t.Helper()

	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "metadata.db"))
	if err != nil {
		t.Fatalf("database.Open() error = %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	return NewSQLService(db)
}
