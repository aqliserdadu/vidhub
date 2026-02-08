package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"videodownload/internal/model"
	"videodownload/internal/service"
	"videodownload/pkg/logger"
	"videodownload/pkg/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DownloadHandler handles download-related requests
type DownloadHandler struct {
	downloadService  *service.DownloadService
	quotaService     *service.QuotaService
	rateLimitService *service.RateLimitService
	cfg              *model.Config
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler(ds *service.DownloadService, cfg *model.Config, qs *service.QuotaService, rls *service.RateLimitService) *DownloadHandler {
	return &DownloadHandler{
		downloadService:  ds,
		quotaService:     qs,
		rateLimitService: rls,
		cfg:              cfg,
	}
}

// StartDownload handles POST /api/download
func (h *DownloadHandler) StartDownload(c *gin.Context) {
	var req model.DownloadRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("Invalid download request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate URL
	if !validator.ValidateURL(req.URL, h.cfg.Security.AllowedDomains) {
		logger.Logger.Warn("Invalid URL domain", zap.String("url", req.URL))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_domain",
			Message: "URL domain is not allowed",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate format ID
	if !validator.ValidateFormatID(req.FormatID) {
		logger.Logger.Warn("Invalid format ID", zap.String("format_id", req.FormatID))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_format",
			Message: "Invalid format ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// ✅ Validate file size BEFORE starting download
	// This prevents worker from processing oversized files
	if req.FileSize > 0 {
		maxSizeBytes := int64(h.cfg.Storage.MaxVideoSizeMB) * 1024 * 1024
		if req.FileSize > maxSizeBytes {
			maxSizeMB := h.cfg.Storage.MaxVideoSizeMB
			fileSizeMB := req.FileSize / (1024 * 1024)
			logger.Logger.Warn("File size exceeds limit",
				zap.Int64("file_size", req.FileSize),
				zap.Int64("max_size", maxSizeBytes),
				zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusRequestEntityTooLarge, model.ErrorResponse{
				Error:   "file_too_large",
				Message: fmt.Sprintf("File size exceeds maximum limit of %dMB. Requested size: %dMB.", maxSizeMB, fileSizeMB),
				Code:    http.StatusRequestEntityTooLarge,
			})
			return
		}
		logger.Logger.Debug("File size validation passed",
			zap.Int64("file_size", req.FileSize),
			zap.Int64("max_size", maxSizeBytes))
	}

	// Check quota before starting download
	clientIP := c.ClientIP()
	if h.cfg.Quota.Enabled {
		// We need to check quota - but we don't know file size yet
		// So we check if quota is even available (not completely exhausted)
		allowed, remainingMB := h.quotaService.CheckQuota(clientIP, 0)
		if !allowed && remainingMB == 0 {
			logger.Logger.Warn("Quota exhausted", zap.String("ip", clientIP))
			quotaInfo := h.quotaService.GetQuotaInfo(clientIP)
			c.JSON(http.StatusPaymentRequired, model.ErrorResponse{
				Error:   "quota_exhausted",
				Message: "Daily download quota exhausted. Please try again after quota reset.",
				Code:    http.StatusPaymentRequired,
			})
			c.Set("quota_info", quotaInfo)
			return
		}
		logger.Logger.Debug("Quota check passed", zap.String("ip", clientIP), zap.Int64("remaining_mb", remainingMB))
	}

	// Start download
	downloadResp, err := h.downloadService.Download(&req)
	if err != nil {
		logger.Logger.Error("Download failed", zap.Error(err), zap.String("url", req.URL))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "download_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Add quota usage if enabled
	if h.cfg.Quota.Enabled {
		// Get file size and add to quota
		fileSizeBytes, err := h.downloadService.GetFileSize(downloadResp.ID)
		if err == nil && fileSizeBytes > 0 {
			fileSizeMB := fileSizeBytes / (1024 * 1024)
			if fileSizeBytes%1024*1024 > 0 {
				fileSizeMB++ // Round up
			}
			h.quotaService.AddUsage(clientIP, fileSizeMB)
			logger.Logger.Debug("Quota usage added", zap.String("ip", clientIP), zap.Int64("size_mb", fileSizeMB))

			// Set quota info in response header
			quotaInfo := h.quotaService.GetQuotaInfo(clientIP)
			c.Set("quota_info", quotaInfo)
		}
	}

	c.JSON(http.StatusOK, downloadResp)
}

// GetFile handles GET /api/download/:id
func (h *DownloadHandler) GetFile(c *gin.Context) {
	fileID := c.Param("id")

	if fileID == "" {
		logger.Logger.Warn("Empty file ID")
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "File ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	file, err := h.downloadService.GetDownloadFile(fileID)
	if err != nil {
		logger.Logger.Warn("File not found", zap.String("file_id", fileID))
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "File not found or has expired",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check if file still exists
	if _, err := os.Stat(file.FilePath); err != nil {
		logger.Logger.Warn("File does not exist", zap.String("path", file.FilePath))
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "File no longer available",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Set proper Content-Disposition header with filename encoding
	// Use RFC 5987 for proper handling of unicode and special characters
	contentDisposition := buildContentDispositionHeader(file.Filename)
	c.Header("Content-Disposition", contentDisposition)
	c.Header("Content-Type", "application/octet-stream")
	c.File(file.FilePath)

	logger.Logger.Info("File downloaded by user",
		zap.String("file_id", fileID),
		zap.String("filename", file.Filename))
}

// buildContentDispositionHeader builds a proper Content-Disposition header
// with RFC 5987 encoding for unicode and special characters
func buildContentDispositionHeader(filename string) string {
	// Check if filename needs encoding (has non-ASCII or special characters)
	needsEncoding := false
	for _, r := range filename {
		if r > 127 || r == '"' || r == '\\' || r == ';' || r == ',' {
			needsEncoding = true
			break
		}
	}

	// Also check for spaces - they should be quoted at minimum
	if strings.ContainsAny(filename, " \t\n\r") {
		needsEncoding = true
	}

	if !needsEncoding {
		// Simple ASCII filename without special characters
		// Just quote it for safety
		return fmt.Sprintf(`attachment; filename="%s"`, filename)
	}

	// Use RFC 5987 encoding for unicode and special characters
	// Format: filename*=UTF-8''<percent-encoded-filename>
	encodedFilename := url.QueryEscape(filename)
	return fmt.Sprintf(`attachment; filename*=UTF-8''%s`, encodedFilename)
}
