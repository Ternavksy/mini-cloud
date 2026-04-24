package server

import (
	"encoding/json"
	"io"
	"mini-cloud/internal/storage"
	"net/http"
	"path"

	"github.com/google/uuid"
)

var store storage.Storage

func InitStorage() {
	var err error
	store, err = storage.NewStorage()
	if err != nil {
		panic("failed to initialize storage: " + err.Error())
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	id := uuid.New().String()

	err = store.Save(id, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"id":   id,
		"name": header.Filename,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := path.Base(r.URL.Path)
	if id == "" || id == "/" {
		http.Error(w, "file id is required", http.StatusBadRequest)
		return
	}

	file, err := store.Get(id)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+id)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	files, err := store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"files": files})
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := path.Base(r.URL.Path)
	if id == "" || id == "/" {
		http.Error(w, "file id is required", http.StatusBadRequest)
		return
	}

	err := store.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "file deleted"})
}
