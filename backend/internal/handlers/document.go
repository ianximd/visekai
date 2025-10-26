package handlers

import (
	"net/http"

	"visekai/backend/internal/middleware"
	"visekai/backend/internal/models"
	"visekai/backend/internal/repository"
	"visekai/backend/pkg/storage"
	"visekai/backend/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DocumentHandler handles document-related requests
type DocumentHandler struct {
	documentRepo *repository.DocumentRepository
	storage      *storage.Storage
	validator    *validator.Validator
	maxFileSize  int64
	allowedExts  []string
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(
	documentRepo *repository.DocumentRepository,
	storage *storage.Storage,
	maxFileSize int64,
	allowedExts []string,
) *DocumentHandler {
	return &DocumentHandler{
		documentRepo: documentRepo,
		storage:      storage,
		validator:    validator.New(),
		maxFileSize:  maxFileSize,
		allowedExts:  allowedExts,
	}
}

// Upload handles document upload
func (h *DocumentHandler) Upload(c *gin.Context) {
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

	// Parse multipart form
	err = c.Request.ParseMultipartForm(h.maxFileSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_003",
			"File too large or invalid multipart form",
			nil,
		))
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_004",
			"No file uploaded",
			nil,
		))
		return
	}

	// Validate file size
	if file.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_005",
			"File size exceeds maximum allowed size",
			nil,
		))
		return
	}

	// Validate file type
	if !storage.ValidateFileType(file.Filename, h.allowedExts) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_006",
			"File type not allowed",
			nil,
		))
		return
	}

	// Save file
	filePath, fileHash, err := h.storage.SaveFile(file, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"SYS_002",
			"Failed to save file",
			nil,
		))
		return
	}

	// Check for duplicate by hash
	existingDoc, err := h.documentRepo.GetByHash(c.Request.Context(), fileHash, userID)
	if err == nil && existingDoc != nil {
		// Delete the newly uploaded file since it's a duplicate
		_ = h.storage.DeleteFile(filePath)
		
		c.JSON(http.StatusOK, models.NewSuccessResponse(
			existingDoc,
			"File already exists (duplicate detected)",
		))
		return
	}

	// Create document record
	document := &models.Document{
		UserID:           userID,
		Filename:         filePath[len(h.storage.GetFilePath("")):], // Relative path
		OriginalFilename: file.Filename,
		FilePath:         filePath,
		FileSize:         file.Size,
		MimeType:         storage.GetMimeType(file.Filename),
		FileHash:         fileHash,
		NumPages:         1, // TODO: Extract actual page count for PDFs
	}

	err = h.documentRepo.Create(c.Request.Context(), document)
	if err != nil {
		// Clean up file on database error
		_ = h.storage.DeleteFile(filePath)
		
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"SYS_003",
			"Failed to create document record",
			nil,
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		document,
		"File uploaded successfully",
	))
}

// List handles listing user's documents
func (h *DocumentHandler) List(c *gin.Context) {
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
	var req models.DocumentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = models.DocumentListRequest{
			Page:    1,
			PerPage: 20,
			SortBy:  "uploaded_at",
		}
	}

	// Get documents
	documents, total, err := h.documentRepo.ListByUser(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"SYS_004",
			"Failed to list documents",
			nil,
		))
		return
	}

	// Calculate pagination
	totalPages := (total + req.PerPage - 1) / req.PerPage
	pagination := models.Pagination{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		models.PaginatedResponse{
			Items:      documents,
			Pagination: pagination,
		},
		"Documents retrieved successfully",
	))
}

// Get handles getting a single document
func (h *DocumentHandler) Get(c *gin.Context) {
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

	// Parse document ID
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_007",
			"Invalid document ID",
			nil,
		))
		return
	}

	// Get document
	document, err := h.documentRepo.GetByID(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"RES_002",
			"Document not found",
			nil,
		))
		return
	}

	// Verify ownership
	if document.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"AUTH_004",
			"Access denied",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		document,
		"Document retrieved successfully",
	))
}

// Delete handles deleting a document
func (h *DocumentHandler) Delete(c *gin.Context) {
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

	// Parse document ID
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_007",
			"Invalid document ID",
			nil,
		))
		return
	}

	// Get document
	document, err := h.documentRepo.GetByID(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"RES_002",
			"Document not found",
			nil,
		))
		return
	}

	// Verify ownership
	if document.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"AUTH_004",
			"Access denied",
			nil,
		))
		return
	}

	// Soft delete document
	err = h.documentRepo.SoftDelete(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"SYS_005",
			"Failed to delete document",
			nil,
		))
		return
	}

	// Note: We don't delete the actual file immediately for safety
	// A cleanup job can handle this later

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		nil,
		"Document deleted successfully",
	))
}
