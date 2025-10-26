package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Storage handles file storage operations
type Storage struct {
	basePath string
}

// NewStorage creates a new storage instance
func NewStorage(basePath string) (*Storage, error) {
	// Ensure base path exists
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{
		basePath: basePath,
	}, nil
}

// SaveFile saves an uploaded file to storage
func (s *Storage) SaveFile(file *multipart.FileHeader, userID uuid.UUID) (filePath string, fileHash string, err error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Create user directory
	userDir := filepath.Join(s.basePath, "documents", userID.String())
	err = os.MkdirAll(userDir, 0755)
	if err != nil {
		return "", "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// Create destination file
	destPath := filepath.Join(userDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Calculate hash while copying
	hash := sha256.New()
	multiWriter := io.MultiWriter(dst, hash)

	// Copy file
	_, err = io.Copy(multiWriter, src)
	if err != nil {
		os.Remove(destPath) // Clean up on error
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	fileHash = fmt.Sprintf("%x", hash.Sum(nil))
	return destPath, fileHash, nil
}

// DeleteFile deletes a file from storage
func (s *Storage) DeleteFile(filePath string) error {
	// Verify file is within basePath (security check)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	absBasePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute base path: %w", err)
	}

	if !strings.HasPrefix(absPath, absBasePath) {
		return fmt.Errorf("file path outside storage directory")
	}

	// Delete file
	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func (s *Storage) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// GetFilePath returns the full path for a file
func (s *Storage) GetFilePath(relativePath string) string {
	return filepath.Join(s.basePath, relativePath)
}

// ValidateFileType checks if the file type is allowed
func ValidateFileType(filename string, allowedExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// GetMimeType returns the MIME type based on file extension
func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".tif":  "image/tiff",
		".pdf":  "application/pdf",
		".webp": "image/webp",
	}

	mimeType, ok := mimeTypes[ext]
	if !ok {
		return "application/octet-stream"
	}
	return mimeType
}
