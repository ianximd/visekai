package services

import (
	"context"
	"fmt"
	"time"

	"visekai/backend/internal/models"
	"visekai/backend/internal/ocr"
	"visekai/backend/internal/repository"
	"visekai/backend/pkg/logger"

	"github.com/google/uuid"
)

// JobService handles OCR job operations
type JobService struct {
	jobRepo      *repository.JobRepository
	resultRepo   *repository.ResultRepository
	documentRepo *repository.DocumentRepository
	ocrClient    *ocr.Client
}

// NewJobService creates a new job service
func NewJobService(
	jobRepo *repository.JobRepository,
	resultRepo *repository.ResultRepository,
	documentRepo *repository.DocumentRepository,
	ocrClient *ocr.Client,
) *JobService {
	return &JobService{
		jobRepo:      jobRepo,
		resultRepo:   resultRepo,
		documentRepo: documentRepo,
		ocrClient:    ocrClient,
	}
}

// SubmitJob creates a new OCR job
func (s *JobService) SubmitJob(ctx context.Context, req models.JobSubmissionRequest, userID uuid.UUID) (*models.OCRJob, error) {
	// Verify document exists and belongs to user
	document, err := s.documentRepo.GetByID(ctx, req.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if document.UserID != userID {
		return nil, fmt.Errorf("unauthorized: document does not belong to user")
	}

	// Create job
	job := &models.OCRJob{
		DocumentID:     req.DocumentID,
		UserID:         userID,
		OCRMode:        req.OCRMode,
		ResolutionMode: req.ResolutionMode,
		Priority:       req.Priority,
		MaxRetries:     3,
		RetryCount:     0,
		Metadata:       req.Metadata,
	}

	err = s.jobRepo.Create(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	logger.Info("OCR job submitted", "job_id", job.ID, "document_id", job.DocumentID, "user_id", userID)

	// Start processing asynchronously
	go s.processJob(context.Background(), job.ID)

	return job, nil
}

// GetJob retrieves a job by ID
func (s *JobService) GetJob(ctx context.Context, jobID uuid.UUID, userID uuid.UUID) (*models.OCRJob, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if job.UserID != userID {
		return nil, fmt.Errorf("unauthorized: job does not belong to user")
	}

	return job, nil
}

// ListJobs retrieves jobs for a user with pagination
func (s *JobService) ListJobs(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.OCRJob, *models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	jobs, total, err := s.jobRepo.GetByUserID(ctx, userID, page, perPage)
	if err != nil {
		return nil, nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	pagination := &models.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return jobs, pagination, nil
}

// CancelJob cancels a pending or processing job
func (s *JobService) CancelJob(ctx context.Context, jobID uuid.UUID, userID uuid.UUID) error {
	// Get job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	// Verify ownership
	if job.UserID != userID {
		return fmt.Errorf("unauthorized: job does not belong to user")
	}

	// Check if job can be cancelled
	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed || job.Status == models.JobStatusCancelled {
		return fmt.Errorf("cannot cancel job with status: %s", job.Status)
	}

	// Update status
	err = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusCancelled, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	logger.Info("OCR job cancelled", "job_id", jobID, "user_id", userID)

	return nil
}

// DeleteJob deletes a completed or failed job
func (s *JobService) DeleteJob(ctx context.Context, jobID uuid.UUID, userID uuid.UUID) error {
	// Get job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	// Verify ownership
	if job.UserID != userID {
		return fmt.Errorf("unauthorized: job does not belong to user")
	}

	// Check if job can be deleted
	if job.Status == models.JobStatusPending || job.Status == models.JobStatusProcessing {
		return fmt.Errorf("cannot delete active job, cancel it first")
	}

	// Delete job (cascade will delete results)
	err = s.jobRepo.Delete(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	logger.Info("OCR job deleted", "job_id", jobID, "user_id", userID)

	return nil
}

// GetJobResult retrieves the result for a job
func (s *JobService) GetJobResult(ctx context.Context, jobID uuid.UUID, userID uuid.UUID) (*models.OCRResult, error) {
	// Verify job ownership
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if job.UserID != userID {
		return nil, fmt.Errorf("unauthorized: job does not belong to user")
	}

	// Get result
	result, err := s.resultRepo.GetByJobID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// processJob processes an OCR job asynchronously
func (s *JobService) processJob(ctx context.Context, jobID uuid.UUID) {
	logger.Info("Starting OCR job processing", "job_id", jobID)

	// Get job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		logger.Error("Failed to get job", "job_id", jobID, "error", err)
		return
	}

	// Check if job is still pending
	if job.Status != models.JobStatusPending {
		logger.Warn("Job is not pending, skipping", "job_id", jobID, "status", job.Status)
		return
	}

	// Update status to processing
	err = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusProcessing, nil)
	if err != nil {
		logger.Error("Failed to update job status", "job_id", jobID, "error", err)
		return
	}

	// Get document
	document, err := s.documentRepo.GetByID(ctx, job.DocumentID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get document: %v", err)
		_ = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusFailed, &errorMsg)
		logger.Error("Failed to get document", "job_id", jobID, "document_id", job.DocumentID, "error", err)
		return
	}

	// Process document with OCR service
	startTime := time.Now()
	ocrResponse, err := s.ocrClient.ProcessDocument(ctx, document.FilePath, job.OCRMode, job.ResolutionMode)
	if err != nil {
		errorMsg := fmt.Sprintf("OCR processing failed: %v", err)
		_ = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusFailed, &errorMsg)

		// Check if we should retry
		if job.RetryCount < job.MaxRetries {
			_ = s.jobRepo.IncrementRetryCount(ctx, jobID)
			_ = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusPending, nil)
			logger.Warn("OCR processing failed, will retry", "job_id", jobID, "retry_count", job.RetryCount+1, "error", err)

			// Retry after a delay
			time.Sleep(10 * time.Second)
			go s.processJob(context.Background(), jobID)
		} else {
			logger.Error("OCR processing failed after max retries", "job_id", jobID, "error", err)
		}
		return
	}

	processingTime := time.Since(startTime)
	logger.Info("OCR processing completed", "job_id", jobID, "processing_time", processingTime)

	// Save result
	result := &models.OCRResult{
		JobID:            jobID,
		DocumentID:       job.DocumentID,
		RawText:          ocrResponse.Text,
		MarkdownText:     ocrResponse.Markdown,
		JSONData:         ocrResponse.StructuredData,
		ConfidenceScore:  ocrResponse.Confidence,
		ProcessingTimeMs: ocrResponse.ProcessingTime,
		NumPages:         ocrResponse.NumPages,
	}

	err = s.resultRepo.Create(ctx, result)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to save result: %v", err)
		_ = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusFailed, &errorMsg)
		logger.Error("Failed to save result", "job_id", jobID, "error", err)
		return
	}

	// Update job status to completed
	err = s.jobRepo.UpdateStatus(ctx, jobID, models.JobStatusCompleted, nil)
	if err != nil {
		logger.Error("Failed to update job status to completed", "job_id", jobID, "error", err)
		return
	}

	logger.Info("OCR job completed successfully", "job_id", jobID, "result_id", result.ID)
}

// GetPendingJobs retrieves pending jobs for processing
func (s *JobService) GetPendingJobs(ctx context.Context, limit int) ([]*models.OCRJob, error) {
	return s.jobRepo.GetPendingJobs(ctx, limit)
}

// ProcessNextJob processes the next pending job in the queue
func (s *JobService) ProcessNextJob(ctx context.Context) error {
	jobs, err := s.GetPendingJobs(ctx, 1)
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil // No jobs to process
	}

	go s.processJob(context.Background(), jobs[0].ID)
	return nil
}
