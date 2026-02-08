package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"videodownload/internal/model"
	"videodownload/pkg/logger"

	"go.uber.org/zap"
)

// VideoService handles video metadata extraction
type VideoService struct {
	pythonWorkerURL string
	httpClient      *http.Client
}

// NewVideoService creates a new video service
func NewVideoService(host string, port int, timeout int) *VideoService {
	return &VideoService{
		pythonWorkerURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// GetVideoInfo fetches video information from yt-dlp worker
func (s *VideoService) GetVideoInfo(videoURL string) (*model.VideoInfo, error) {
	endpoint := s.pythonWorkerURL + "/api/info"

	reqBody := map[string]string{"url": videoURL}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.Logger.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Logger.Error("Failed to fetch video info", zap.Error(err), zap.String("url", videoURL))
		return nil, fmt.Errorf("failed to fetch video info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Warn("Non-OK status from python worker", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("python worker returned status %d", resp.StatusCode)
	}

	var metadata model.VideoMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		logger.Logger.Error("Failed to decode response", zap.Error(err))
		return nil, err
	}

	videoInfo := s.parseMetadata(metadata)
	logger.Logger.Info("Video info retrieved", zap.String("title", videoInfo.Title), zap.Int("formats", len(videoInfo.Formats)))
	return videoInfo, nil
}

// parseMetadata converts raw metadata to VideoInfo
func (s *VideoService) parseMetadata(metadata model.VideoMetadata) *model.VideoInfo {
	formats := []model.FormatOption{}

	for _, fmt := range metadata.Formats {
		format := s.parseFormat(fmt)
		if format != nil {
			formats = append(formats, *format)
		}
	}

	return &model.VideoInfo{
		URL:          metadata.URL,
		Title:        metadata.Title,
		Duration:     int(metadata.Duration),
		ThumbnailURL: metadata.Thumbnail,
		Uploader:     metadata.Uploader,
		Formats:      formats,
	}
}

// parseFormat converts raw format data to FormatOption
func (s *VideoService) parseFormat(rawFmt map[string]interface{}) *model.FormatOption {
	formatID, _ := rawFmt["format_id"].(string)
	ext, _ := rawFmt["ext"].(string)

	// Skip non-downloadable formats
	if ext == "" {
		return nil
	}

	format := &model.FormatOption{
		FormatID:  formatID,
		Extension: ext,
	}

	if v, ok := rawFmt["format"].(string); ok {
		format.Format = v
	}
	if v, ok := rawFmt["resolution"].(string); ok {
		format.Resolution = v
	}
	if v, ok := rawFmt["vcodec"].(string); ok {
		format.VideoCodec = v
	}
	if v, ok := rawFmt["acodec"].(string); ok {
		format.AudioCodec = v
	}
	if v, ok := rawFmt["filesize"].(float64); ok {
		format.FileSize = int64(v)
	}
	if v, ok := rawFmt["fps"].(float64); ok {
		format.Fps = int(v)
	}

	format.Quality = s.determineQuality(format)
	format.OfficialName = s.buildOfficialName(format)

	return format
}

// determineQuality determines video quality based on resolution
func (s *VideoService) determineQuality(format *model.FormatOption) string {
	if format.VideoCodec == "" || format.VideoCodec == "none" {
		return "Audio"
	}

	switch format.Resolution {
	case "3840x2160", "4k":
		return "4K"
	case "2560x1440":
		return "1440p"
	case "1920x1080":
		return "FHD"
	case "1280x720":
		return "HD"
	case "854x480":
		return "480p"
	case "640x360":
		return "360p"
	default:
		return format.Resolution
	}
}

// buildOfficialName builds a readable format name
func (s *VideoService) buildOfficialName(format *model.FormatOption) string {
	if format.Quality == "Audio" {
		return fmt.Sprintf("%s - %s", format.Quality, format.AudioCodec)
	}
	return fmt.Sprintf("%s (%s) - %s + %s", format.Quality, format.Resolution, format.VideoCodec, format.AudioCodec)
}
