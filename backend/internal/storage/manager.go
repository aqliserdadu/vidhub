package storage

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"videodownload/internal/model"
	"videodownload/pkg/logger"

	"go.uber.org/zap"
)

// Manager handles file storage and cleanup
type Manager struct {
	cfg      *model.StorageConfig
	files    map[string]*model.DownloadedFile
	mu       sync.RWMutex
	quitChan chan bool
}

// NewManager creates a new storage manager
func NewManager(cfg *model.StorageConfig) *Manager {
	return &Manager{
		cfg:      cfg,
		files:    make(map[string]*model.DownloadedFile),
		quitChan: make(chan bool),
	}
}

// Start starts the cleanup routine
func (m *Manager) Start() {
	go m.cleanupRoutine()
}

// Stop stops the cleanup routine
func (m *Manager) Stop() {
	// Use a non-blocking send with recovery for cleanup
	select {
	case m.quitChan <- true:
		// Successfully sent stop signal
	default:
		// Channel is full or closed, log if possible
		if logger.Logger != nil {
			logger.Logger.Warn("Could not send stop signal to cleanup routine")
		}
	}
}

// SaveFile saves file info for tracking
func (m *Manager) SaveFile(id string, file *model.DownloadedFile) error {
	file.ID = id
	file.CreatedAt = time.Now()
	file.ExpiresAt = time.Now().Add(time.Duration(m.cfg.FileTTLSeconds) * time.Second)

	m.mu.Lock()
	m.files[id] = file
	m.mu.Unlock()

	logger.Logger.Info("File saved", zap.String("id", id), zap.String("filename", file.Filename))
	return nil
}

// cleanupRoutine periodically removes expired files
func (m *Manager) cleanupRoutine() {
	ticker := time.NewTicker(time.Duration(m.cfg.CleanupInterval) * time.Second)
	defer ticker.Stop()

	// Log initial start with safe logging
	if logger.Logger != nil {
		logger.Logger.Info("Storage cleanup routine started",
			zap.Int("cleanup_interval_seconds", m.cfg.CleanupInterval),
			zap.Int("file_ttl_seconds", m.cfg.FileTTLSeconds))
	}

	for {
		select {
		case <-m.quitChan:
			if logger.Logger != nil {
				logger.Logger.Info("Storage cleanup routine stopped")
			}
			return
		case <-ticker.C:
			m.cleanupExpiredFiles()
		}
	}
}

// cleanupExpiredFiles removes files that have expired
func (m *Manager) cleanupExpiredFiles() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	deletedCount := 0
	errorCount := 0
	var deletedIds []string

	for id, file := range m.files {
		if now.After(file.ExpiresAt) {
			// Try to remove the actual file
			if err := os.Remove(file.FilePath); err != nil {
				if !os.IsNotExist(err) {
					// Log error only if file exists but can't be deleted
					if logger.Logger != nil {
						logger.Logger.Error("Failed to remove file",
							zap.String("id", id),
							zap.String("path", file.FilePath),
							zap.Error(err))
					}
					errorCount++
				} else {
					// File doesn't exist - that's okay, just remove from tracking
					if logger.Logger != nil {
						logger.Logger.Debug("File already deleted",
							zap.String("id", id),
							zap.String("path", file.FilePath))
					}
				}
			} else {
				if logger.Logger != nil {
					logger.Logger.Info("File removed by cleanup",
						zap.String("id", id),
						zap.String("path", file.FilePath))
				}
				deletedCount++
			}
			// Remove from tracking map regardless of deletion success
			deletedIds = append(deletedIds, id)
		}
	}

	// Delete from map
	for _, id := range deletedIds {
		delete(m.files, id)
	}

	// Log summary if anything happened and logger is available
	if logger.Logger != nil && (deletedCount > 0 || errorCount > 0) {
		logger.Logger.Info("Storage cleanup completed",
			zap.Int("deleted_count", deletedCount),
			zap.Int("error_count", errorCount),
			zap.Int("remaining_tracked_files", len(m.files)))
	}
}

// GetFile gets file info by ID
func (m *Manager) GetFile(id string) *model.DownloadedFile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.files[id]
}

// ValidateFileSize checks if file size is within limits
func (m *Manager) ValidateFileSize(sizeBytes int64) bool {
	maxSizeBytes := int64(m.cfg.MaxVideoSizeMB) * 1024 * 1024
	return sizeBytes <= maxSizeBytes
}

// EnsureDownloadDir ensures download directory exists
func (m *Manager) EnsureDownloadDir() error {
	return os.MkdirAll(m.cfg.DownloadDir, 0755)
}

// GetDownloadPath returns the path where file should be downloaded
func (m *Manager) GetDownloadPath(filename string) string {
	return filepath.Join(m.cfg.DownloadDir, filename)
}

// GetFileTTL returns the file time to live in seconds
func (m *Manager) GetFileTTL() int {
	return m.cfg.FileTTLSeconds
}

// GetCleanupInterval returns the cleanup interval in seconds
func (m *Manager) GetCleanupInterval() int {
	return m.cfg.CleanupInterval
}

// GetTrackedFilesCount returns the number of files currently being tracked
func (m *Manager) GetTrackedFilesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.files)
}

// GetTrackedFilesInfo returns information about all tracked files
func (m *Manager) GetTrackedFilesInfo() map[string]*model.DownloadedFile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make(map[string]*model.DownloadedFile)
	for k, v := range m.files {
		result[k] = v
	}
	return result
}

// GetExpiredFilesCount returns the number of files that have expired but not yet deleted
func (m *Manager) GetExpiredFilesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := time.Now()
	count := 0
	for _, file := range m.files {
		if now.After(file.ExpiresAt) {
			count++
		}
	}
	return count
}

// ManualCleanup manually triggers a cleanup run (useful for testing)
func (m *Manager) ManualCleanup() {
	m.cleanupExpiredFiles()
}
