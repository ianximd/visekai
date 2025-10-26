package models

import (
	"time"

	"github.com/google/uuid"
)

// Document represents a uploaded document
type Document struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Filename         string     `json:"filename"`
	OriginalFilename string     `json:"original_filename"`
	FilePath         string     `json:"file_path"`
	FileSize         int64      `json:"file_size"`
	MimeType         string     `json:"mime_type"`
	FileHash         string     `json:"file_hash"`
	NumPages         int        `json:"num_pages"`
	ThumbnailPath    *string    `json:"thumbnail_path,omitempty"`
	UploadedAt       time.Time  `json:"uploaded_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

// DocumentUploadRequest represents the metadata for a document upload
type DocumentUploadRequest struct {
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
}

// DocumentListRequest represents pagination and filter parameters
type DocumentListRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PerPage  int    `json:"per_page" validate:"min=1,max=100"`
	SortBy   string `json:"sort_by" validate:"omitempty,oneof=uploaded_at filename file_size"`
	SortDesc bool   `json:"sort_desc"`
}
