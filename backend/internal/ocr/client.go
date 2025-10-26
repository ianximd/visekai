package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"visekai/backend/internal/models"
	"visekai/backend/pkg/logger"
)

// Client handles communication with the OCR service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new OCR client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // OCR can take time
		},
	}
}

// OCRRequest represents a request to the OCR service
type OCRRequest struct {
	Mode       string `json:"mode"`        // document, handwritten, general, figure
	Resolution string `json:"resolution"`  // tiny, small, base, large, gundam
}

// OCRResponse represents a response from the OCR service
type OCRResponse struct {
	Success        bool                   `json:"success"`
	Text           string                 `json:"text"`
	Markdown       string                 `json:"markdown"`
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime int                    `json:"processing_time_ms"`
	NumPages       int                    `json:"num_pages"`
	Error          string                 `json:"error,omitempty"`
}

// ProcessDocument sends a document to the OCR service for processing
func (c *Client) ProcessDocument(ctx context.Context, filePath string, ocrMode models.OCRMode, resolutionMode models.ResolutionMode) (*OCRResponse, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Add OCR parameters
	_ = writer.WriteField("mode", string(ocrMode))
	_ = writer.WriteField("resolution", string(resolutionMode))

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/ocr/process", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	logger.Info("Sending OCR request", "url", url, "file", filepath.Base(filePath), "mode", ocrMode, "resolution", resolutionMode)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Error("OCR service returned error", "status", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("OCR service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var ocrResp OCRResponse
	err = json.Unmarshal(respBody, &ocrResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !ocrResp.Success {
		return nil, fmt.Errorf("OCR processing failed: %s", ocrResp.Error)
	}

	logger.Info("OCR processing completed", "confidence", ocrResp.Confidence, "processing_time_ms", ocrResp.ProcessingTime)

	return &ocrResp, nil
}

// HealthCheck checks if the OCR service is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OCR service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// GetStatus gets the status of the OCR service
func (c *Client) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/status", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var status map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return status, nil
}
