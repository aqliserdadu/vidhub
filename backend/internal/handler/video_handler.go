package handler

import (
	"net/http"

	"videodownload/internal/model"
	"videodownload/internal/service"
	"videodownload/pkg/logger"
	"videodownload/pkg/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// VideoHandler handles video-related requests
type VideoHandler struct {
	videoService *service.VideoService
	cfg          *model.Config
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(vs *service.VideoService, cfg *model.Config) *VideoHandler {
	return &VideoHandler{
		videoService: vs,
		cfg:          cfg,
	}
}

// GetVideoInfo handles GET /api/video/info
func (h *VideoHandler) GetVideoInfo(c *gin.Context) {
	videoURL := c.Query("url")

	if videoURL == "" {
		logger.Logger.Warn("Empty URL provided")
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_url",
			Message: "Video URL is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate URL
	if !validator.ValidateURL(videoURL, h.cfg.Security.AllowedDomains) {
		logger.Logger.Warn("Invalid URL domain",
			zap.String("url", videoURL),
			zap.Strings("allowed_domains", h.cfg.Security.AllowedDomains))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_domain",
			Message: "URL domain is not allowed",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get video info from service
	videoInfo, err := h.videoService.GetVideoInfo(videoURL)
	if err != nil {
		logger.Logger.Error("Failed to get video info", zap.Error(err), zap.String("url", videoURL))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to fetch video information",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, videoInfo)
}

// HealthCheck handles GET /health
func (h *VideoHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "video-downloader",
	})
}
