// Code generated from api/openapi.yaml; DO NOT EDIT.

package api

import "time"

type APIError struct {
	Message string `json:"message"`
}

type CollectionsEnvelope struct {
	Data CollectionsResponse `json:"data"`
}

type CollectionsResponse struct {
	Collections []FileCollectionResponse `json:"collections"`
}

type CreateCollectionRequest struct {
	FileIDs []string `json:"file_ids"`
	Name    string   `json:"name"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type FileCollectionEnvelope struct {
	Data FileCollectionResponse `json:"data"`
}

type FileCollectionResponse struct {
	CreatedAt time.Time `json:"created_at"`
	FileIDs   []string  `json:"file_ids"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
}

type FilesEnvelope struct {
	Data FilesResponse `json:"data"`
}

type FilesResponse struct {
	Files []string `json:"files"`
}

type MessageEnvelope struct {
	Data MessageResponse `json:"data"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type UploadFileEnvelope struct {
	Data UploadFileResponse `json:"data"`
}

type UploadFileResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
