package server

import (
	"encoding/json"
	"mini-cloud/internal/storage"
	"net/http"

	"github.com/google/uuid"
)

var store = storage.New("./storage")

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

func ListHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("list"))
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("delete"))
}
