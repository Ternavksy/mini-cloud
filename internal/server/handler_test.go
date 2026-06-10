package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mini-cloud/internal/filecollections"
	"mini-cloud/internal/models"
)

func TestFileRoutesUseUnifiedResponseFormat(t *testing.T) {
	store := mockFileStore{
		saveFunc: func(id string, r io.Reader) error {
			data, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			if string(data) != "hello" {
				t.Fatalf("saved data = %q, want %q", string(data), "hello")
			}
			return nil
		},
		getFunc: func(id string) (io.ReadCloser, error) {
			if id != "file-1" {
				t.Fatalf("get id = %q, want file-1", id)
			}
			return io.NopCloser(strings.NewReader("hello")), nil
		},
		deleteFunc: func(id string) error {
			if id != "file-1" {
				t.Fatalf("delete id = %q, want file-1", id)
			}
			return nil
		},
		listFunc: func() ([]string, error) {
			return []string{"file-1", "__file_collections/index.json"}, nil
		},
	}
	router := NewRouter(store, emptyCollectionService())

	uploadBody, contentType := multipartBody(t, "file", "test.txt", "hello")
	uploadRecorder := httptest.NewRecorder()
	uploadRequest := httptest.NewRequest(http.MethodPost, "/upload", uploadBody)
	uploadRequest.Header.Set("Content-Type", contentType)
	router.ServeHTTP(uploadRecorder, uploadRequest)
	assertStatus(t, uploadRecorder, http.StatusOK)
	assertJSONHasKey(t, uploadRecorder.Body.String(), "data")

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/files", nil))
	assertStatus(t, listRecorder, http.StatusOK)
	assertJSONHasKey(t, listRecorder.Body.String(), "data")
	if strings.Contains(listRecorder.Body.String(), "__file_collections") {
		t.Fatalf("List response leaked metadata key: %s", listRecorder.Body.String())
	}

	downloadRecorder := httptest.NewRecorder()
	router.ServeHTTP(downloadRecorder, httptest.NewRequest(http.MethodGet, "/download/file-1", nil))
	assertStatus(t, downloadRecorder, http.StatusOK)
	if downloadRecorder.Body.String() != "hello" {
		t.Fatalf("download body = %q, want hello", downloadRecorder.Body.String())
	}

	deleteRecorder := httptest.NewRecorder()
	router.ServeHTTP(deleteRecorder, httptest.NewRequest(http.MethodDelete, "/files/file-1", nil))
	assertStatus(t, deleteRecorder, http.StatusOK)
	assertJSONHasKey(t, deleteRecorder.Body.String(), "data")
}

func TestCollectionRoutesUseDomainTypesAndUnifiedResponses(t *testing.T) {
	collection := models.FileCollection{
		BaseModel: models.BaseModel{
			ID:        "collection-1",
			CreatedAt: time.Now().UTC(),
		},
		Name:    "album",
		FileIDs: []models.FileID{"file-1"},
	}

	collections := mockCollectionService{
		listFunc: func() ([]models.FileCollection, error) {
			return []models.FileCollection{collection}, nil
		},
		createFunc: func(name string, fileIDs []models.FileID) (models.FileCollection, error) {
			if name != "album" {
				t.Fatalf("collection name = %q, want album", name)
			}
			if len(fileIDs) != 1 || fileIDs[0] != "file-1" {
				t.Fatalf("file ids = %v, want [file-1]", fileIDs)
			}
			return collection, nil
		},
		getFunc: func(id models.CollectionID) (models.FileCollection, error) {
			if id != "collection-1" {
				t.Fatalf("collection id = %q, want collection-1", id)
			}
			return collection, nil
		},
		addFileFunc: func(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
			if collectionID != "collection-1" || fileID != "file-2" {
				t.Fatalf("add args = %q %q, want collection-1 file-2", collectionID, fileID)
			}
			collection.FileIDs = append(collection.FileIDs, fileID)
			return collection, nil
		},
		removeFileFunc: func(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
			if collectionID != "collection-1" || fileID != "file-1" {
				t.Fatalf("remove args = %q %q, want collection-1 file-1", collectionID, fileID)
			}
			return collection, nil
		},
		deleteFunc: func(id models.CollectionID) error {
			if id != "collection-1" {
				t.Fatalf("delete collection id = %q, want collection-1", id)
			}
			return nil
		},
	}

	router := NewRouter(emptyFileStore(), collections)
	requests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/collections", "", http.StatusOK},
		{http.MethodPost, "/collections", `{"name":"album","file_ids":["file-1"]}`, http.StatusCreated},
		{http.MethodGet, "/collections/collection-1", "", http.StatusOK},
		{http.MethodPost, "/collections/collection-1/files/file-2", "", http.StatusOK},
		{http.MethodDelete, "/collections/collection-1/files/file-1", "", http.StatusOK},
		{http.MethodDelete, "/collections/collection-1", "", http.StatusOK},
	}

	for _, request := range requests {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(request.method, request.path, strings.NewReader(request.body))
		if request.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		router.ServeHTTP(recorder, req)
		assertStatus(t, recorder, request.status)
		assertJSONHasKey(t, recorder.Body.String(), "data")
	}
}

func TestCollectionNotFoundUsesErrorEnvelope(t *testing.T) {
	collections := emptyCollectionService()
	collections.getFunc = func(id models.CollectionID) (models.FileCollection, error) {
		return models.FileCollection{}, filecollections.ErrCollectionNotFound
	}

	router := NewRouter(emptyFileStore(), collections)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/collections/missing", nil))

	assertStatus(t, recorder, http.StatusNotFound)
	assertJSONHasKey(t, recorder.Body.String(), "error")
}

func TestListFilesErrorUsesErrorEnvelope(t *testing.T) {
	store := emptyFileStore()
	store.listFunc = func() ([]string, error) {
		return nil, errors.New("storage unavailable")
	}

	router := NewRouter(store, emptyCollectionService())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/files", nil))

	assertStatus(t, recorder, http.StatusInternalServerError)
	assertJSONHasKey(t, recorder.Body.String(), "error")
}

func emptyFileStore() mockFileStore {
	return mockFileStore{
		saveFunc: func(id string, r io.Reader) error {
			return nil
		},
		getFunc: func(id string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		},
		deleteFunc: func(id string) error {
			return nil
		},
		listFunc: func() ([]string, error) {
			return nil, nil
		},
	}
}

func emptyCollectionService() mockCollectionService {
	return mockCollectionService{
		listFunc: func() ([]models.FileCollection, error) {
			return nil, nil
		},
		createFunc: func(name string, fileIDs []models.FileID) (models.FileCollection, error) {
			return models.FileCollection{}, nil
		},
		getFunc: func(id models.CollectionID) (models.FileCollection, error) {
			return models.FileCollection{}, nil
		},
		addFileFunc: func(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
			return models.FileCollection{}, nil
		},
		removeFileFunc: func(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
			return models.FileCollection{}, nil
		},
		deleteFunc: func(id models.CollectionID) error {
			return nil
		},
	}
}

func multipartBody(t *testing.T, fieldName, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return body, writer.FormDataContentType()
}

func assertStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, want, recorder.Body.String())
	}
}

func assertJSONHasKey(t *testing.T, body, key string) {
	t.Helper()

	var response map[string]any
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("response is not json: %v; body = %s", err, body)
	}
	if _, ok := response[key]; !ok {
		t.Fatalf("response missing key %q: %s", key, body)
	}
}
