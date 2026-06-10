package filecollections

import (
	"testing"

	"mini-cloud/internal/models"
	"mini-cloud/internal/storage"
)

func TestServiceCreateListGetAndRemoveFile(t *testing.T) {
	store := storage.New(t.TempDir())
	service := NewService(store)

	created, err := service.Create("summer album", []models.FileID{"file-1", "file-2", "file-1", ""})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name != "summer album" {
		t.Fatalf("Create() name = %q, want %q", created.Name, "summer album")
	}
	if len(created.FileIDs) != 2 {
		t.Fatalf("Create() file count = %d, want 2", len(created.FileIDs))
	}

	collections, err := service.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(collections) != 1 {
		t.Fatalf("List() count = %d, want 1", len(collections))
	}

	got, err := service.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("Get() id = %q, want %q", got.ID, created.ID)
	}

	updated, err := service.AddFile(created.ID, "file-3")
	if err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if len(updated.FileIDs) != 3 || updated.FileIDs[2] != "file-3" {
		t.Fatalf("AddFile() file ids = %v, want file-3 appended", updated.FileIDs)
	}

	updated, err = service.RemoveFile(created.ID, "file-1")
	if err != nil {
		t.Fatalf("RemoveFile() error = %v", err)
	}
	if len(updated.FileIDs) != 2 || updated.FileIDs[0] != "file-2" || updated.FileIDs[1] != "file-3" {
		t.Fatalf("RemoveFile() file ids = %v, want [file-2 file-3]", updated.FileIDs)
	}

	if err := service.Delete(created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	collections, err = service.List()
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(collections) != 0 {
		t.Fatalf("List() after delete count = %d, want 0", len(collections))
	}
}
