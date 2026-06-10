package server

//go:generate go run mockgen.go

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mini-cloud/internal/api"
	"mini-cloud/internal/filecollections"
	"mini-cloud/internal/models"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type fileStore interface {
	Save(id string, r io.Reader) error
	Get(id string) (io.ReadCloser, error)
	Delete(id string) error
	List() ([]string, error)
}

type collectionService interface {
	List() ([]models.FileCollection, error)
	Create(name string, fileIDs []models.FileID) (models.FileCollection, error)
	Get(id models.CollectionID) (models.FileCollection, error)
	AddFile(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error)
	RemoveFile(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error)
	Delete(id models.CollectionID) error
}

type Handler struct {
	store       fileStore
	collections collectionService
}

func NewHandler(store fileStore, collections collectionService) *Handler {
	return &Handler{
		store:       store,
		collections: collections,
	}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	defer file.Close()

	id := models.FileID(uuid.New().String())

	err = h.store.Save(string(id), file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.UploadFileEnvelope{
		Data: api.UploadFileResponse{
			ID:   string(id),
			Name: header.Filename,
		},
	})
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "file id is required", nil)
		return
	}

	file, err := h.store.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found", err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+id)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	files, err := h.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	files = filterCollectionMetadata(files)
	writeJSON(w, http.StatusOK, api.FilesEnvelope{
		Data: api.FilesResponse{Files: files},
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "file id is required", nil)
		return
	}

	err := h.store.Delete(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.MessageEnvelope{
		Data: api.MessageResponse{Message: "file deleted"},
	})
}

func (h *Handler) ListCollections(w http.ResponseWriter, r *http.Request) {
	collections, err := h.collections.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.CollectionsEnvelope{
		Data: api.CollectionsResponse{Collections: toCollectionResponses(collections)},
	})
}

func (h *Handler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var request api.CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	collection, err := h.collections.Create(request.Name, toFileIDs(request.FileIDs))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusCreated, api.FileCollectionEnvelope{
		Data: toCollectionResponse(collection),
	})
}

func (h *Handler) GetCollection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	collection, err := h.collections.Get(models.CollectionID(id))
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.FileCollectionEnvelope{
		Data: toCollectionResponse(collection),
	})
}

func (h *Handler) AddFileToCollection(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")
	fileID := chi.URLParam(r, "fileID")

	collection, err := h.collections.AddFile(models.CollectionID(collectionID), models.FileID(fileID))
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.FileCollectionEnvelope{
		Data: toCollectionResponse(collection),
	})
}

func (h *Handler) RemoveFileFromCollection(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")
	fileID := chi.URLParam(r, "fileID")

	collection, err := h.collections.RemoveFile(models.CollectionID(collectionID), models.FileID(fileID))
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.FileCollectionEnvelope{
		Data: toCollectionResponse(collection),
	})
}

func (h *Handler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.collections.Delete(models.CollectionID(id))
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, api.MessageEnvelope{
		Data: api.MessageResponse{Message: "collection deleted"},
	})
}

func filterCollectionMetadata(files []string) []string {
	filtered := files[:0]
	for _, file := range files {
		if !filecollections.IsMetadataKey(file) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func writeJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Printf("request error: status=%d message=%q err=%v", status, message, err)
	} else {
		log.Printf("request error: status=%d message=%q", status, message)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorEnvelope{
		Error: api.APIError{Message: message},
	})
}

func toFileIDs(ids []string) []models.FileID {
	fileIDs := make([]models.FileID, 0, len(ids))
	for _, id := range ids {
		fileIDs = append(fileIDs, models.FileID(id))
	}
	return fileIDs
}

func toCollectionResponses(collections []models.FileCollection) []api.FileCollectionResponse {
	responses := make([]api.FileCollectionResponse, 0, len(collections))
	for _, collection := range collections {
		responses = append(responses, toCollectionResponse(collection))
	}
	return responses
}

func toCollectionResponse(collection models.FileCollection) api.FileCollectionResponse {
	return api.FileCollectionResponse{
		ID:        string(collection.ID),
		Name:      collection.Name,
		FileIDs:   toStrings(collection.FileIDs),
		CreatedAt: collection.CreatedAt,
	}
}

func toStrings(ids []models.FileID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, string(id))
	}
	return values
}
