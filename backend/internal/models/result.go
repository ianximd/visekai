package models

import (
	"time"

	"github.com/google/uuid"
)

// OCRResult represents the result of an OCR job
type OCRResult struct {
	ID               uuid.UUID      `json:"id"`
	JobID            uuid.UUID      `json:"job_id"`
	DocumentID       uuid.UUID      `json:"document_id"`
	RawText          string         `json:"raw_text"`
	MarkdownText     string         `json:"markdown_text"`
	JSONData         map[string]any `json:"json_data,omitempty"`
	ConfidenceScore  float64        `json:"confidence_score"`
	ProcessingTimeMs int            `json:"processing_time_ms"`
	NumPages         int            `json:"num_pages"`
	CreatedAt        time.Time      `json:"created_at"`
}

// ResultExportFormat represents the export format for OCR results
type ResultExportFormat string

const (
	ExportFormatMarkdown ResultExportFormat = "markdown"
	ExportFormatJSON     ResultExportFormat = "json"
	ExportFormatText     ResultExportFormat = "text"
	ExportFormatPDF      ResultExportFormat = "pdf"
	ExportFormatDOCX     ResultExportFormat = "docx"
)

// ResultExportRequest represents the data needed to export a result
type ResultExportRequest struct {
	Format ResultExportFormat `json:"format" validate:"required,oneof=markdown json text pdf docx"`
}
