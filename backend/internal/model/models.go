package model

import "time"

// VideoInfo contains metadata about a video
type VideoInfo struct {
	URL          string         `json:"url"`
	Title        string         `json:"title"`
	Duration     int            `json:"duration"`
	ThumbnailURL string         `json:"thumbnail_url"`
	Uploader     string         `json:"uploader"`
	Formats      []FormatOption `json:"formats"`
}

// FormatOption represents a downloadable format
type FormatOption struct {
	FormatID     string `json:"format_id"`
	Format       string `json:"format"`
	Extension    string `json:"ext"`
	Resolution   string `json:"resolution"`
	VideoCodec   string `json:"video_codec"`
	AudioCodec   string `json:"audio_codec"`
	FileSize     int64  `json:"file_size"`
	Fps          int    `json:"fps"`
	Quality      string `json:"quality"` // FHD, HD, SD, Audio
	OfficialName string `json:"official_name"`
}

// DownloadRequest represents a user's download request
type DownloadRequest struct {
	URL      string `json:"url" binding:"required"`
	FormatID string `json:"format_id" binding:"required"`
	Quality  string `json:"quality"`   // FHD, HD, Audio, etc.
	FileSize int64  `json:"file_size"` // File size in bytes for backend validation
}

// DownloadResponse represents the response to a download request
type DownloadResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	DownloadLink string `json:"download_link"`
	ExpiresAt    int64  `json:"expires_at"`
}

// DownloadedFile tracks downloaded files for cleanup
type DownloadedFile struct {
	ID        string
	Filename  string
	FilePath  string
	Size      int64
	CreatedAt time.Time
	ExpiresAt time.Time
	URL       string
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// VideoMetadata contains parsed video metadata from yt-dlp
type VideoMetadata struct {
	ID        string                   `json:"id"`
	Title     string                   `json:"title"`
	Duration  float64                  `json:"duration"`
	Thumbnail string                   `json:"thumbnail"`
	Uploader  string                   `json:"uploader"`
	URL       string                   `json:"url"`
	Formats   []map[string]interface{} `json:"formats"`
}
