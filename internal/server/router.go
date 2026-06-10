package server

import (
	"net/http"

	"github.com/go-chi/chi"
)

func NewRouter(store fileStore, collections collectionService) http.Handler {
	r := chi.NewRouter()
	handler := NewHandler(store, collections)

	r.Post("/upload", handler.Upload)
	r.Get("/files", handler.List)
	r.Get("/download/{id}", handler.Download)
	r.Delete("/files/{id}", handler.Delete)
	r.Get("/collections", handler.ListCollections)
	r.Post("/collections", handler.CreateCollection)
	r.Get("/collections/{id}", handler.GetCollection)
	r.Delete("/collections/{id}", handler.DeleteCollection)
	r.Post("/collections/{id}/files/{fileID}", handler.AddFileToCollection)
	r.Delete("/collections/{id}/files/{fileID}", handler.RemoveFileFromCollection)

	return r
}
