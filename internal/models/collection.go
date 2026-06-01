package models

import "time"

type FileCollection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	FileIDs   []string  `json:"file_ids"`
	CreatedAt time.Time `json:"created_at"`
}
