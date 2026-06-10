package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"mini-cloud/internal/database"
	"mini-cloud/internal/filecollections"
	"mini-cloud/internal/storage"
)

func TestIntegrationFileAndCollectionFlow(t *testing.T) {
	tempDir := t.TempDir()
	store := storage.New(filepath.Join(tempDir, "files"))

	db, err := database.Open(context.Background(), filepath.Join(tempDir, "metadata.db"))
	if err != nil {
		t.Fatalf("database.Open() error = %v", err)
	}
	defer db.Close()

	router := NewRouter(store, filecollections.NewSQLService(db))

	uploadBody, uploadContentType := multipartBody(t, "file", "note.txt", "hello")
	uploadRecorder := httptest.NewRecorder()
	uploadRequest := httptest.NewRequest(http.MethodPost, "/upload", uploadBody)
	uploadRequest.Header.Set("Content-Type", uploadContentType)
	router.ServeHTTP(uploadRecorder, uploadRequest)
	assertStatus(t, uploadRecorder, http.StatusOK)

	fileID := jsonPath(t, uploadRecorder.Body.String(), "data", "id")
	if fileID == "" {
		t.Fatalf("upload response missing file id: %s", uploadRecorder.Body.String())
	}

	createRecorder := httptest.NewRecorder()
	createBody := `{"name":"album","file_ids":["` + fileID + `"]}`
	createRequest := httptest.NewRequest(http.MethodPost, "/collections", strings.NewReader(createBody))
	createRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, createRequest)
	assertStatus(t, createRecorder, http.StatusCreated)

	collectionID := jsonPath(t, createRecorder.Body.String(), "data", "id")
	if collectionID == "" {
		t.Fatalf("create collection response missing id: %s", createRecorder.Body.String())
	}

	getCollectionRecorder := httptest.NewRecorder()
	router.ServeHTTP(getCollectionRecorder, httptest.NewRequest(http.MethodGet, "/collections/"+collectionID, nil))
	assertStatus(t, getCollectionRecorder, http.StatusOK)

	downloadRecorder := httptest.NewRecorder()
	router.ServeHTTP(downloadRecorder, httptest.NewRequest(http.MethodGet, "/download/"+fileID, nil))
	assertStatus(t, downloadRecorder, http.StatusOK)
	if downloadRecorder.Body.String() != "hello" {
		t.Fatalf("download body = %q, want hello", downloadRecorder.Body.String())
	}

	removeRecorder := httptest.NewRecorder()
	router.ServeHTTP(removeRecorder, httptest.NewRequest(http.MethodDelete, "/collections/"+collectionID+"/files/"+fileID, nil))
	assertStatus(t, removeRecorder, http.StatusOK)

	deleteCollectionRecorder := httptest.NewRecorder()
	router.ServeHTTP(deleteCollectionRecorder, httptest.NewRequest(http.MethodDelete, "/collections/"+collectionID, nil))
	assertStatus(t, deleteCollectionRecorder, http.StatusOK)
}

func jsonPath(t *testing.T, body string, keys ...string) string {
	t.Helper()

	var value any
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		t.Fatalf("invalid json response: %v; body = %s", err, body)
	}

	for _, key := range keys {
		object, ok := value.(map[string]any)
		if !ok {
			return ""
		}
		value = object[key]
	}

	result, _ := value.(string)
	return result
}
