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

// DocumentRepository handles document database operations
type DocumentRepository struct {
	db *pgxpool.Pool
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new document in the database
func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (
			id, user_id, filename, original_filename, file_path,
			file_size, mime_type, file_hash, num_pages, thumbnail_path, uploaded_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	doc.ID = uuid.New()
	doc.UploadedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		doc.ID,
		doc.UserID,
		doc.Filename,
		doc.OriginalFilename,
		doc.FilePath,
		doc.FileSize,
		doc.MimeType,
		doc.FileHash,
		doc.NumPages,
		doc.ThumbnailPath,
		doc.UploadedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	query := `
		SELECT id, user_id, filename, original_filename, file_path,
		       file_size, mime_type, file_hash, num_pages, thumbnail_path,
		       uploaded_at, deleted_at
		FROM documents
		WHERE id = $1 AND deleted_at IS NULL
	`

	var doc models.Document
	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.UserID,
		&doc.Filename,
		&doc.OriginalFilename,
		&doc.FilePath,
		&doc.FileSize,
		&doc.MimeType,
		&doc.FileHash,
		&doc.NumPages,
		&doc.ThumbnailPath,
		&doc.UploadedAt,
		&doc.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

// ListByUser retrieves documents for a specific user with pagination
func (r *DocumentRepository) ListByUser(ctx context.Context, userID uuid.UUID, req models.DocumentListRequest) ([]models.Document, int, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}
	if req.SortBy == "" {
		req.SortBy = "uploaded_at"
	}

	offset := (req.Page - 1) * req.PerPage
	order := "DESC"
	if !req.SortDesc {
		order = "ASC"
	}

	// Count total documents
	countQuery := `SELECT COUNT(*) FROM documents WHERE user_id = $1 AND deleted_at IS NULL`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents
	query := fmt.Sprintf(`
		SELECT id, user_id, filename, original_filename, file_path,
		       file_size, mime_type, file_hash, num_pages, thumbnail_path,
		       uploaded_at, deleted_at
		FROM documents
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, req.SortBy, order)

	rows, err := r.db.Query(ctx, query, userID, req.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
			&doc.ID,
			&doc.UserID,
			&doc.Filename,
			&doc.OriginalFilename,
			&doc.FilePath,
			&doc.FileSize,
			&doc.MimeType,
			&doc.FileHash,
			&doc.NumPages,
			&doc.ThumbnailPath,
			&doc.UploadedAt,
			&doc.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, total, nil
}

// SoftDelete soft deletes a document
func (r *DocumentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE documents SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// GetByHash retrieves a document by file hash (for deduplication)
func (r *DocumentRepository) GetByHash(ctx context.Context, hash string, userID uuid.UUID) (*models.Document, error) {
	query := `
		SELECT id, user_id, filename, original_filename, file_path,
		       file_size, mime_type, file_hash, num_pages, thumbnail_path,
		       uploaded_at, deleted_at
		FROM documents
		WHERE file_hash = $1 AND user_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var doc models.Document
	err := r.db.QueryRow(ctx, query, hash, userID).Scan(
		&doc.ID,
		&doc.UserID,
		&doc.Filename,
		&doc.OriginalFilename,
		&doc.FilePath,
		&doc.FileSize,
		&doc.MimeType,
		&doc.FileHash,
		&doc.NumPages,
		&doc.ThumbnailPath,
		&doc.UploadedAt,
		&doc.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not found is not an error for this use case
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document by hash: %w", err)
	}

	return &doc, nil
}
