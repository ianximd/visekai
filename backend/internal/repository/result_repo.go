package repository

import (
	"context"
	"fmt"
	"time"

	"visekai/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ResultRepository handles OCR result database operations
type ResultRepository struct {
	db *pgxpool.Pool
}

// NewResultRepository creates a new result repository
func NewResultRepository(db *pgxpool.Pool) *ResultRepository {
	return &ResultRepository{db: db}
}

// Create creates a new OCR result
func (r *ResultRepository) Create(ctx context.Context, result *models.OCRResult) error {
	query := `
		INSERT INTO ocr_results (
			id, job_id, document_id, raw_text, markdown_text, json_data,
			confidence_score, processing_time_ms, num_pages, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	result.ID = uuid.New()
	result.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		result.ID,
		result.JobID,
		result.DocumentID,
		result.RawText,
		result.MarkdownText,
		result.JSONData,
		result.ConfidenceScore,
		result.ProcessingTimeMs,
		result.NumPages,
		result.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create result: %w", err)
	}

	return nil
}

// GetByID retrieves a result by ID
func (r *ResultRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OCRResult, error) {
	query := `
		SELECT id, job_id, document_id, raw_text, markdown_text, json_data,
			   confidence_score, processing_time_ms, num_pages, created_at
		FROM ocr_results
		WHERE id = $1
	`

	var result models.OCRResult
	err := r.db.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.JobID,
		&result.DocumentID,
		&result.RawText,
		&result.MarkdownText,
		&result.JSONData,
		&result.ConfidenceScore,
		&result.ProcessingTimeMs,
		&result.NumPages,
		&result.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("result not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	return &result, nil
}

// GetByJobID retrieves a result by job ID
func (r *ResultRepository) GetByJobID(ctx context.Context, jobID uuid.UUID) (*models.OCRResult, error) {
	query := `
		SELECT id, job_id, document_id, raw_text, markdown_text, json_data,
			   confidence_score, processing_time_ms, num_pages, created_at
		FROM ocr_results
		WHERE job_id = $1
	`

	var result models.OCRResult
	err := r.db.QueryRow(ctx, query, jobID).Scan(
		&result.ID,
		&result.JobID,
		&result.DocumentID,
		&result.RawText,
		&result.MarkdownText,
		&result.JSONData,
		&result.ConfidenceScore,
		&result.ProcessingTimeMs,
		&result.NumPages,
		&result.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("result not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	return &result, nil
}

// GetByDocumentID retrieves results by document ID
func (r *ResultRepository) GetByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*models.OCRResult, error) {
	query := `
		SELECT id, job_id, document_id, raw_text, markdown_text, json_data,
			   confidence_score, processing_time_ms, num_pages, created_at
		FROM ocr_results
		WHERE document_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get results: %w", err)
	}
	defer rows.Close()

	var results []*models.OCRResult
	for rows.Next() {
		var result models.OCRResult
		err := rows.Scan(
			&result.ID,
			&result.JobID,
			&result.DocumentID,
			&result.RawText,
			&result.MarkdownText,
			&result.JSONData,
			&result.ConfidenceScore,
			&result.ProcessingTimeMs,
			&result.NumPages,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, &result)
	}

	return results, nil
}

// Update updates an existing result
func (r *ResultRepository) Update(ctx context.Context, result *models.OCRResult) error {
	query := `
		UPDATE ocr_results
		SET raw_text = $1, markdown_text = $2, json_data = $3,
		    confidence_score = $4, processing_time_ms = $5, num_pages = $6
		WHERE id = $7
	`

	res, err := r.db.Exec(ctx, query,
		result.RawText,
		result.MarkdownText,
		result.JSONData,
		result.ConfidenceScore,
		result.ProcessingTimeMs,
		result.NumPages,
		result.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update result: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("result not found")
	}

	return nil
}

// Delete deletes a result
func (r *ResultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ocr_results WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete result: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("result not found")
	}

	return nil
}
