package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mini-cloud/internal/filecollections"
	"mini-cloud/internal/storage"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	store       storage.Storage
	collections filecollections.Service
}

type createCollectionRequest struct {
	Name    string   `json:"name"`
	FileIDs []string `json:"file_ids"`
}

func NewHandler(store storage.Storage, collections filecollections.Service) *Handler {
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

	id := uuid.New().String()

	err = h.store.Save(id, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":   id,
		"name": header.Filename,
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
	writeJSON(w, http.StatusOK, map[string][]string{"files": files})
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

	writeJSON(w, http.StatusOK, map[string]string{"message": "file deleted"})
}

func (h *Handler) ListCollections(w http.ResponseWriter, r *http.Request) {
	collections, err := h.collections.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"collections": collections})
}

func (h *Handler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var request createCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	collection, err := h.collections.Create(request.Name, request.FileIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusCreated, collection)
}

func (h *Handler) GetCollection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	collection, err := h.collections.Get(id)
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, collection)
}

func (h *Handler) AddFileToCollection(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")
	fileID := chi.URLParam(r, "fileID")

	collection, err := h.collections.AddFile(collectionID, fileID)
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, collection)
}

func (h *Handler) RemoveFileFromCollection(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "id")
	fileID := chi.URLParam(r, "fileID")

	collection, err := h.collections.RemoveFile(collectionID, fileID)
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, collection)
}

func (h *Handler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.collections.Delete(id)
	if err != nil {
		if errors.Is(err, filecollections.ErrCollectionNotFound) {
			writeError(w, http.StatusNotFound, "collection not found", err)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "collection deleted"})
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
	http.Error(w, message, status)
}
