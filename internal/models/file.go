package models

import "time"

type File struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	FilePath    string     `json:"file_path" db:"file_path"`
	Size        int64      `json:"size" db:"size"`
	IsEncrypted bool       `json:"is_encrypted" db:"is_encrypted"`
	StorageKey  string     `json:"storage_key" db:"storage_key"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`

	Status     string     `json:"status" db:"status"`          
	UploadedAt *time.Time `json:"uploaded_at" db:"uploaded_at"`
}
