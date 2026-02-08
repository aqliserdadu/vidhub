package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"videodownload/internal/model"
	"videodownload/internal/storage"
	"videodownload/pkg/logger"

	"go.uber.org/zap"
)

// DownloadService handles video downloads
type DownloadService struct {
	pythonWorkerURL string
	httpClient      *http.Client
	storageManager  *storage.Manager
}

// NewDownloadService creates a new download service
func NewDownloadService(host string, port int, timeout int, sm *storage.Manager) *DownloadService {
	return &DownloadService{
		pythonWorkerURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		storageManager: sm,
	}
}

// Download starts downloading a video
func (s *DownloadService) Download(req *model.DownloadRequest) (*model.DownloadResponse, error) {
	// Validate file size before downloading
	endpoint := s.pythonWorkerURL + "/api/download"

	reqBody := map[string]string{
		"url":       req.URL,
		"format_id": req.FormatID,
		"quality":   req.Quality,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.Logger.Error("Failed to create download request", zap.Error(err))
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		logger.Logger.Error("Download failed", zap.Error(err), zap.String("url", req.URL))
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Warn("Failed download response", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Try to read Python worker response (may include filename in JSON)
	var pythonResponse *model.PythonWorkerDownloadResponse
	var filename string
	var fileDataBytes []byte

	// First, try to read full body into buffer
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("Failed to read response body", zap.Error(err))
		return nil, err
	}

	// Try to parse as JSON first (Python worker may return filename in JSON)
	if jsonErr := json.Unmarshal(respBodyBytes, &pythonResponse); jsonErr == nil && pythonResponse != nil && pythonResponse.Filename != "" {
		// Python worker returned JSON with filename
		filename = pythonResponse.Filename
		logger.Logger.Debug("Filename extracted from Python worker JSON response",
			zap.String("filename", filename))
	} else {
		// Fallback: extract from Content-Disposition header or use default
		if cd := resp.Header.Get("Content-Disposition"); cd != "" {
			filename = s.extractFilenameFromHeader(cd)
		}

		if filename == "" {
			filename = "video_download.mp4"
			logger.Logger.Warn("Could not extract filename, using default",
				zap.String("content_disposition", resp.Header.Get("Content-Disposition")),
				zap.String("default_filename", filename))
		} else {
			logger.Logger.Debug("Filename extracted from Content-Disposition header",
				zap.String("filename", filename))
		}

		// If we parsed as JSON and got this far, respBodyBytes is the JSON, not the file data
		// This shouldn't happen in normal flow, but safeguard against it
		if string(respBodyBytes[:min(50, len(respBodyBytes))]) == "{" {
			logger.Logger.Error("Response appears to be JSON error, not file data")
			return nil, fmt.Errorf("invalid response from Python worker")
		}
	}

	fileDataBytes = respBodyBytes

	// Note: Filename is already truncated by Python worker, don't truncate again
	// Double truncation can cause issues with UTF-8 characters
	logger.Logger.Info("Download from Python worker completed",
		zap.String("filename", filename),
		zap.Int("size_bytes", len(fileDataBytes)))

	// Validate file size
	if !s.storageManager.ValidateFileSize(int64(len(fileDataBytes))) {
		logger.Logger.Warn("File size exceeds limit", zap.String("filename", filename), zap.Int("size", len(fileDataBytes)))
		return nil, fmt.Errorf("file size exceeds maximum limit of %dMB", 300)
	}

	// Save file
	if err := s.storageManager.EnsureDownloadDir(); err != nil {
		logger.Logger.Error("Failed to create download directory", zap.Error(err))
		return nil, err
	}

	downloadPath := s.storageManager.GetDownloadPath(filename)
	if err := os.WriteFile(downloadPath, fileDataBytes, 0644); err != nil {
		logger.Logger.Error("Failed to write file", zap.Error(err), zap.String("filename", filename))
		return nil, err
	}
	logger.Logger.Info("File saved to disk",
		zap.String("path", downloadPath),
		zap.String("filename", filename),
		zap.Int64("size_bytes", int64(len(fileDataBytes))))

	// Generate download response
	downloadID := fmt.Sprintf("%d", time.Now().UnixNano())
	file := &model.DownloadedFile{
		Filename: filename,
		FilePath: downloadPath,
		Size:     int64(len(fileDataBytes)),
		URL:      req.URL,
	}

	if err := s.storageManager.SaveFile(downloadID, file); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(s.storageManager.GetFileTTL()) * time.Second).Unix()

	logger.Logger.Info("Download completed and tracked",
		zap.String("download_id", downloadID),
		zap.String("filename", filename),
		zap.Int64("expires_at", expiresAt))

	return &model.DownloadResponse{
		ID:           downloadID,
		Title:        filename,
		DownloadLink: fmt.Sprintf("/api/download/%s", downloadID),
		ExpiresAt:    expiresAt,
	}, nil
}

// extractFilenameFromHeader extracts filename from Content-Disposition header
func (s *DownloadService) extractFilenameFromHeader(cd string) string {
	// Use mime.ParseMediaType for proper RFC 2183 parsing
	_, params, err := mime.ParseMediaType(cd)

	if err == nil {
		// Try to get filename* (RFC 5987 - UTF-8 encoded) first
		if fn, ok := params["filename*"]; ok && fn != "" {
			// RFC 5987 format: UTF-8''encodedfilename
			if strings.HasPrefix(fn, "UTF-8''") {
				encodedFn := fn[7:] // Remove "UTF-8''" prefix
				// Decode the RFC 5987 encoded filename
				decodedFn, decodeErr := queryUnescape(encodedFn)
				if decodeErr == nil && decodedFn != "" {
					return filepath.Base(decodedFn)
				}
			}
		}

		// Fallback to filename (RFC 2183) if filename* failed or not present
		if fn, ok := params["filename"]; ok && fn != "" {
			return filepath.Base(fn)
		}
	} else {
		// Parsing failed, try legacy simple split approach
		parts := strings.Split(cd, "filename=")
		if len(parts) > 1 {
			filename := strings.Trim(parts[1], "\"")
			return filepath.Base(filename)
		}
	}

	return ""
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// queryUnescape decodes URL-encoded strings (used for RFC 5987 filenames)
func queryUnescape(s string) (string, error) {
	// Decode %XX sequences and handle + as space for application/x-www-form-urlencoded
	// But RFC 5987 doesn't use + for space, so we use a different approach
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			if i+2 < len(s) {
				// Parse hex value
				hex := s[i+1 : i+3]
				var b byte
				_, err := fmt.Sscanf(hex, "%x", &b)
				if err != nil {
					return "", fmt.Errorf("invalid hex in encoded filename: %v", err)
				}
				result.WriteByte(b)
				i += 2
			} else {
				result.WriteByte(s[i])
			}
		default:
			result.WriteByte(s[i])
		}
	}
	return result.String(), nil
}

// GetDownloadFile retrieves a downloaded file for streaming
func (s *DownloadService) GetDownloadFile(fileID string) (*model.DownloadedFile, error) {
	file := s.storageManager.GetFile(fileID)
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	return file, nil
}

// GetFileSize returns the size of a downloaded file
func (s *DownloadService) GetFileSize(fileID string) (int64, error) {
	file := s.storageManager.GetFile(fileID)
	if file == nil {
		return 0, fmt.Errorf("file not found")
	}

	return file.Size, nil
}
