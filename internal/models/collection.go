package models

import "time"

type CollectionID string

type BaseModel struct {
	ID        CollectionID `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
}

type FileCollection struct {
	BaseModel
	Name    string   `json:"name"`
	FileIDs []FileID `json:"file_ids"`
}
