package models

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of an OCR job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// OCRMode represents the OCR processing mode
type OCRMode string

const (
	OCRModeDocument    OCRMode = "document"
	OCRModeHandwritten OCRMode = "handwritten"
	OCRModeGeneral     OCRMode = "general"
	OCRModeFigure      OCRMode = "figure"
)

// ResolutionMode represents the OCR resolution mode
type ResolutionMode string

const (
	ResolutionTiny   ResolutionMode = "tiny"
	ResolutionSmall  ResolutionMode = "small"
	ResolutionBase   ResolutionMode = "base"
	ResolutionLarge  ResolutionMode = "large"
	ResolutionGundam ResolutionMode = "gundam"
)

// OCRJob represents an OCR processing job
type OCRJob struct {
	ID                 uuid.UUID      `json:"id"`
	DocumentID         uuid.UUID      `json:"document_id"`
	UserID             uuid.UUID      `json:"user_id"`
	Status             JobStatus      `json:"status"`
	OCRMode            OCRMode        `json:"ocr_mode"`
	ResolutionMode     ResolutionMode `json:"resolution_mode"`
	Priority           int            `json:"priority"`
	RetryCount         int            `json:"retry_count"`
	MaxRetries         int            `json:"max_retries"`
	ProgressPercentage int            `json:"progress_percentage"`
	CreatedAt          time.Time      `json:"created_at"`
	StartedAt          *time.Time     `json:"started_at,omitempty"`
	CompletedAt        *time.Time     `json:"completed_at,omitempty"`
	ErrorMessage       *string        `json:"error_message,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// OCRJobRequest represents the data needed to submit an OCR job
type OCRJobRequest struct {
	DocumentID     uuid.UUID      `json:"document_id" validate:"required"`
	OCRMode        OCRMode        `json:"ocr_mode" validate:"required,oneof=document handwritten general figure"`
	ResolutionMode ResolutionMode `json:"resolution_mode" validate:"required,oneof=tiny small base large gundam"`
	Priority       int            `json:"priority" validate:"min=0,max=10"`
}

// JobSubmissionRequest represents internal job submission data
type JobSubmissionRequest struct {
	DocumentID     uuid.UUID
	OCRMode        OCRMode
	ResolutionMode ResolutionMode
	Priority       int
	Metadata       map[string]any
}

// BatchOCRJobRequest represents the data needed to submit batch OCR jobs
type BatchOCRJobRequest struct {
	DocumentIDs    []uuid.UUID    `json:"document_ids" validate:"required,min=1,max=50"`
	OCRMode        OCRMode        `json:"ocr_mode" validate:"required"`
	ResolutionMode ResolutionMode `json:"resolution_mode" validate:"required"`
}

// JobListRequest represents pagination and filter parameters for jobs
type JobListRequest struct {
	Page     int       `json:"page" validate:"min=1"`
	PerPage  int       `json:"per_page" validate:"min=1,max=100"`
	Status   JobStatus `json:"status" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	SortBy   string    `json:"sort_by" validate:"omitempty,oneof=created_at status priority"`
	SortDesc bool      `json:"sort_desc"`
}
