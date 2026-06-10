package storage

import (
	"io"
	"strings"
	"testing"
)

func TestLocalStorageSaveGetListDelete(t *testing.T) {
	store := New(t.TempDir())

	if err := store.Save("folder/file-1", strings.NewReader("hello")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	files, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(files) != 1 || files[0] != "folder/file-1" {
		t.Fatalf("List() files = %v, want [folder/file-1]", files)
	}

	file, err := store.Get("folder/file-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("Get() data = %q, want %q", string(data), "hello")
	}

	if err := store.Delete("folder/file-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get("folder/file-1"); err == nil {
		t.Fatal("Get() after delete error = nil, want error")
	}
}
