package models

import "time"

type ShareLink struct {
	ID           string     `json:"id" db:"id"`
	FileID       string     `json:"file_id" db:"file_id"`
	ShareToken   string     `json:"share_token" db:"share_token"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}
