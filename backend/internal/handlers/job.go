package handlers

import (
	"net/http"

	"visekai/backend/internal/middleware"
	"visekai/backend/internal/models"
	"visekai/backend/internal/services"
	"visekai/backend/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// JobHandler handles OCR job-related requests
type JobHandler struct {
	jobService *services.JobService
	validator  *validator.Validator
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobService *services.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		validator:  validator.New(),
	}
}

// SubmitJob handles OCR job submission
func (h *JobHandler) SubmitJob(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse request
	var req models.OCRJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			"Invalid request body",
			nil,
		))
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			err.Error(),
			nil,
		))
		return
	}

	// Create submission request
	submission := models.JobSubmissionRequest{
		DocumentID:     req.DocumentID,
		OCRMode:        req.OCRMode,
		ResolutionMode: req.ResolutionMode,
		Priority:       req.Priority,
	}

	// Submit job
	job, err := h.jobService.SubmitJob(c.Request.Context(), submission, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"JOB_001",
			err.Error(),
			nil,
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		job,
		"OCR job submitted successfully",
	))
}

// SubmitBatchJob handles batch OCR job submission
func (h *JobHandler) SubmitBatchJob(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse request
	var req models.BatchOCRJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			"Invalid request body",
			nil,
		))
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			err.Error(),
			nil,
		))
		return
	}

	// Submit jobs for each document
	var jobs []*models.OCRJob
	var errors []string

	for _, documentID := range req.DocumentIDs {
		submission := models.JobSubmissionRequest{
			DocumentID:     documentID,
			OCRMode:        req.OCRMode,
			ResolutionMode: req.ResolutionMode,
			Priority:       0, // Batch jobs have default priority
		}

		job, err := h.jobService.SubmitJob(c.Request.Context(), submission, userID)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		jobs = append(jobs, job)
	}

	response := gin.H{
		"jobs":    jobs,
		"success": len(jobs),
		"failed":  len(errors),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		response,
		"Batch OCR jobs submitted",
	))
}

// ListJobs handles listing user's OCR jobs
func (h *JobHandler) ListJobs(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse pagination
	var req models.JobListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = models.JobListRequest{
			Page:    1,
			PerPage: 20,
		}
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	// Get jobs
	jobs, pagination, err := h.jobService.ListJobs(c.Request.Context(), userID, req.Page, req.PerPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"SYS_006",
			"Failed to list jobs",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.PaginatedResponse{
			Items:      jobs,
			Pagination: *pagination,
		},
		"Jobs retrieved successfully",
	))
}

// GetJob handles getting a single OCR job
func (h *JobHandler) GetJob(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse job ID
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_008",
			"Invalid job ID",
			nil,
		))
		return
	}

	// Get job
	job, err := h.jobService.GetJob(c.Request.Context(), jobID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"RES_003",
			"Job not found",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		job,
		"Job retrieved successfully",
	))
}

// CancelJob handles cancelling an OCR job
func (h *JobHandler) CancelJob(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse job ID
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_008",
			"Invalid job ID",
			nil,
		))
		return
	}

	// Cancel job
	err = h.jobService.CancelJob(c.Request.Context(), jobID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"JOB_002",
			err.Error(),
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		nil,
		"Job cancelled successfully",
	))
}

// DeleteJob handles deleting an OCR job
func (h *JobHandler) DeleteJob(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse job ID
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_008",
			"Invalid job ID",
			nil,
		))
		return
	}

	// Delete job
	err = h.jobService.DeleteJob(c.Request.Context(), jobID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"JOB_003",
			err.Error(),
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		nil,
		"Job deleted successfully",
	))
}

// GetJobResult handles getting the result of an OCR job
func (h *JobHandler) GetJobResult(c *gin.Context) {
	// Get authenticated user
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Parse job ID
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_008",
			"Invalid job ID",
			nil,
		))
		return
	}

	// Get result
	result, err := h.jobService.GetJobResult(c.Request.Context(), jobID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"RES_004",
			"Result not found",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		result,
		"Result retrieved successfully",
	))
}
