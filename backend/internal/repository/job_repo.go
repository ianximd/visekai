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

// JobRepository handles OCR job database operations
type JobRepository struct {
	db *pgxpool.Pool
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new OCR job
func (r *JobRepository) Create(ctx context.Context, job *models.OCRJob) error {
	query := `
		INSERT INTO ocr_jobs (
			id, document_id, user_id, status, ocr_mode, resolution_mode,
			priority, retry_count, max_retries, progress_percentage, created_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	job.ID = uuid.New()
	job.Status = models.JobStatusPending
	job.CreatedAt = time.Now()
	job.ProgressPercentage = 0

	_, err := r.db.Exec(ctx, query,
		job.ID,
		job.DocumentID,
		job.UserID,
		job.Status,
		job.OCRMode,
		job.ResolutionMode,
		job.Priority,
		job.RetryCount,
		job.MaxRetries,
		job.ProgressPercentage,
		job.CreatedAt,
		job.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by ID
func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OCRJob, error) {
	query := `
		SELECT id, document_id, user_id, status, ocr_mode, resolution_mode,
			   priority, retry_count, max_retries, progress_percentage,
			   created_at, started_at, completed_at, error_message, metadata
		FROM ocr_jobs
		WHERE id = $1
	`

	var job models.OCRJob
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.DocumentID,
		&job.UserID,
		&job.Status,
		&job.OCRMode,
		&job.ResolutionMode,
		&job.Priority,
		&job.RetryCount,
		&job.MaxRetries,
		&job.ProgressPercentage,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
		&job.Metadata,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// GetByUserID retrieves all jobs for a user with pagination
func (r *JobRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.OCRJob, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	countQuery := `SELECT COUNT(*) FROM ocr_jobs WHERE user_id = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Get jobs
	query := `
		SELECT id, document_id, user_id, status, ocr_mode, resolution_mode,
			   priority, retry_count, max_retries, progress_percentage,
			   created_at, started_at, completed_at, error_message, metadata
		FROM ocr_jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.OCRJob
	for rows.Next() {
		var job models.OCRJob
		err := rows.Scan(
			&job.ID,
			&job.DocumentID,
			&job.UserID,
			&job.Status,
			&job.OCRMode,
			&job.ResolutionMode,
			&job.Priority,
			&job.RetryCount,
			&job.MaxRetries,
			&job.ProgressPercentage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.ErrorMessage,
			&job.Metadata,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, total, nil
}

// UpdateStatus updates the status of a job
func (r *JobRepository) UpdateStatus(ctx context.Context, jobID uuid.UUID, status models.JobStatus, errorMessage *string) error {
	var query string
	var args []interface{}

	now := time.Now()

	switch status {
	case models.JobStatusProcessing:
		query = `
			UPDATE ocr_jobs
			SET status = $1, started_at = $2
			WHERE id = $3
		`
		args = []interface{}{status, now, jobID}

	case models.JobStatusCompleted, models.JobStatusFailed, models.JobStatusCancelled:
		if errorMessage != nil {
			query = `
				UPDATE ocr_jobs
				SET status = $1, completed_at = $2, error_message = $3, progress_percentage = $4
				WHERE id = $5
			`
			progress := 0
			if status == models.JobStatusCompleted {
				progress = 100
			}
			args = []interface{}{status, now, *errorMessage, progress, jobID}
		} else {
			query = `
				UPDATE ocr_jobs
				SET status = $1, completed_at = $2, progress_percentage = $3
				WHERE id = $4
			`
			progress := 100
			args = []interface{}{status, now, progress, jobID}
		}

	default:
		query = `UPDATE ocr_jobs SET status = $1 WHERE id = $2`
		args = []interface{}{status, jobID}
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// UpdateProgress updates the progress percentage of a job
func (r *JobRepository) UpdateProgress(ctx context.Context, jobID uuid.UUID, progress int) error {
	query := `UPDATE ocr_jobs SET progress_percentage = $1 WHERE id = $2`

	result, err := r.db.Exec(ctx, query, progress, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// IncrementRetryCount increments the retry count for a job
func (r *JobRepository) IncrementRetryCount(ctx context.Context, jobID uuid.UUID) error {
	query := `UPDATE ocr_jobs SET retry_count = retry_count + 1 WHERE id = $1`

	result, err := r.db.Exec(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// GetPendingJobs retrieves all pending jobs ordered by priority and creation time
func (r *JobRepository) GetPendingJobs(ctx context.Context, limit int) ([]*models.OCRJob, error) {
	query := `
		SELECT id, document_id, user_id, status, ocr_mode, resolution_mode,
			   priority, retry_count, max_retries, progress_percentage,
			   created_at, started_at, completed_at, error_message, metadata
		FROM ocr_jobs
		WHERE status = $1
		ORDER BY priority DESC, created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, models.JobStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.OCRJob
	for rows.Next() {
		var job models.OCRJob
		err := rows.Scan(
			&job.ID,
			&job.DocumentID,
			&job.UserID,
			&job.Status,
			&job.OCRMode,
			&job.ResolutionMode,
			&job.Priority,
			&job.RetryCount,
			&job.MaxRetries,
			&job.ProgressPercentage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.ErrorMessage,
			&job.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// Delete deletes a job
func (r *JobRepository) Delete(ctx context.Context, jobID uuid.UUID) error {
	query := `DELETE FROM ocr_jobs WHERE id = $1`

	result, err := r.db.Exec(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// GetJobsByStatus retrieves jobs by status with pagination
func (r *JobRepository) GetJobsByStatus(ctx context.Context, userID uuid.UUID, status models.JobStatus, page, perPage int) ([]*models.OCRJob, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	countQuery := `SELECT COUNT(*) FROM ocr_jobs WHERE user_id = $1 AND status = $2`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID, status).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Get jobs
	query := `
		SELECT id, document_id, user_id, status, ocr_mode, resolution_mode,
			   priority, retry_count, max_retries, progress_percentage,
			   created_at, started_at, completed_at, error_message, metadata
		FROM ocr_jobs
		WHERE user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, userID, status, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.OCRJob
	for rows.Next() {
		var job models.OCRJob
		err := rows.Scan(
			&job.ID,
			&job.DocumentID,
			&job.UserID,
			&job.Status,
			&job.OCRMode,
			&job.ResolutionMode,
			&job.Priority,
			&job.RetryCount,
			&job.MaxRetries,
			&job.ProgressPercentage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&job.ErrorMessage,
			&job.Metadata,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, total, nil
}
