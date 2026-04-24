package server

import (
	"net/http"

	"github.com/go-chi/chi"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/upload", UploadHandler)
	r.Get("/files", ListHandler)
	r.Get("/download/{id}", DownloadHandler)
	r.Delete("/files/{id}", DeleteHandler)

	return r
}
