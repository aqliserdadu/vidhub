package service

import (
	"sync"
	"time"

	"videodownload/internal/model"
	"videodownload/pkg/logger"

	"go.uber.org/zap"
)

// QuotaEntry tracks quota usage per IP
type QuotaEntry struct {
	IP         string
	UsedMB     int64
	ResetTime  time.Time
	LastUpdate time.Time
}

// QuotaService manages user download quotas
type QuotaService struct {
	cfg      *model.QuotaConfig
	quotas   map[string]*QuotaEntry
	mu       sync.RWMutex
	quitChan chan bool
}

// NewQuotaService creates a new quota service
func NewQuotaService(cfg *model.QuotaConfig) *QuotaService {
	service := &QuotaService{
		cfg:      cfg,
		quotas:   make(map[string]*QuotaEntry),
		quitChan: make(chan bool),
	}

	if cfg.Enabled {
		go service.resetRoutine()
	}

	return service
}

// CheckQuota checks if IP has remaining quota
func (qs *QuotaService) CheckQuota(ip string, requestedSizeMB int64) (bool, int64) {
	if !qs.cfg.Enabled {
		return true, qs.cfg.DailyLimitMB
	}

	qs.mu.RLock()
	entry, exists := qs.quotas[ip]
	qs.mu.RUnlock()

	// Create new quota entry if not exists
	if !exists {
		qs.mu.Lock()
		resetTime := qs.calculateResetTime()
		qs.quotas[ip] = &QuotaEntry{
			IP:         ip,
			UsedMB:     0,
			ResetTime:  resetTime,
			LastUpdate: time.Now(),
		}
		entry = qs.quotas[ip]
		qs.mu.Unlock()

		logger.Logger.Info("New quota entry created", zap.String("ip", ip), zap.Time("reset_time", resetTime))
	}

	// Check if quota reset time has passed
	now := time.Now()
	if now.After(entry.ResetTime) {
		qs.mu.Lock()
		entry.UsedMB = 0
		entry.ResetTime = qs.calculateResetTime()
		entry.LastUpdate = now
		qs.mu.Unlock()

		logger.Logger.Info("Quota reset for IP", zap.String("ip", ip), zap.Time("new_reset_time", entry.ResetTime))
	}

	// Check if quota is available
	remaining := qs.cfg.DailyLimitMB - entry.UsedMB
	if remaining <= 0 {
		logger.Logger.Warn("Quota exhausted", zap.String("ip", ip), zap.Int64("limit_mb", qs.cfg.DailyLimitMB))
		return false, 0
	}

	if requestedSizeMB > remaining {
		logger.Logger.Warn("Quota insufficient", zap.String("ip", ip), zap.Int64("requested_mb", requestedSizeMB), zap.Int64("remaining_mb", remaining))
		return false, remaining
	}

	return true, remaining
}

// AddUsage adds to quota usage for an IP
func (qs *QuotaService) AddUsage(ip string, sizeMB int64) error {
	if !qs.cfg.Enabled {
		return nil
	}

	qs.mu.Lock()
	defer qs.mu.Unlock()

	entry, exists := qs.quotas[ip]
	if !exists {
		resetTime := qs.calculateResetTime()
		qs.quotas[ip] = &QuotaEntry{
			IP:         ip,
			UsedMB:     sizeMB,
			ResetTime:  resetTime,
			LastUpdate: time.Now(),
		}
		logger.Logger.Info("Quota usage added for new IP", zap.String("ip", ip), zap.Int64("used_mb", sizeMB))
		return nil
	}

	entry.UsedMB += sizeMB
	entry.LastUpdate = time.Now()

	logger.Logger.Debug("Quota usage updated", zap.String("ip", ip), zap.Int64("used_mb", entry.UsedMB), zap.Int64("limit_mb", qs.cfg.DailyLimitMB))

	return nil
}

// GetQuotaInfo returns current quota info for IP
func (qs *QuotaService) GetQuotaInfo(ip string) map[string]interface{} {
	if !qs.cfg.Enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	qs.mu.RLock()
	defer qs.mu.RUnlock()

	entry, exists := qs.quotas[ip]
	if !exists {
		resetTime := qs.calculateResetTime()
		return map[string]interface{}{
			"enabled":      true,
			"used_mb":      0,
			"limit_mb":     qs.cfg.DailyLimitMB,
			"remaining_mb": qs.cfg.DailyLimitMB,
			"reset_time":   resetTime,
		}
	}

	remaining := qs.cfg.DailyLimitMB - entry.UsedMB
	if remaining < 0 {
		remaining = 0
	}

	return map[string]interface{}{
		"enabled":      true,
		"used_mb":      entry.UsedMB,
		"limit_mb":     qs.cfg.DailyLimitMB,
		"remaining_mb": remaining,
		"reset_time":   entry.ResetTime,
	}
}

// calculateResetTime calculates next reset time based on config
func (qs *QuotaService) calculateResetTime() time.Time {
	now := time.Now()
	resetTime := time.Date(now.Year(), now.Month(), now.Day(), qs.cfg.ResetHour, qs.cfg.ResetMinute, 0, 0, now.Location())

	// If reset time has already passed today, set for tomorrow
	if resetTime.Before(now) {
		resetTime = resetTime.AddDate(0, 0, 1)
	}

	return resetTime
}

// resetRoutine periodically checks and resets quotas
func (qs *QuotaService) resetRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-qs.quitChan:
			logger.Logger.Info("Quota service stopped")
			return
		case <-ticker.C:
			qs.checkAndResetQuotas()
		}
	}
}

// checkAndResetQuotas checks for expired quotas and resets them
func (qs *QuotaService) checkAndResetQuotas() {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	now := time.Now()
	resetCount := 0

	for _, entry := range qs.quotas {
		if now.After(entry.ResetTime) {
			entry.UsedMB = 0
			entry.ResetTime = qs.calculateResetTime()
			entry.LastUpdate = now
			resetCount++
		}
	}

	if resetCount > 0 {
		logger.Logger.Info("Quota reset completed", zap.Int("entries_reset", resetCount))
	}
}

// Stop stops the quota service
func (qs *QuotaService) Stop() {
	if qs.cfg.Enabled {
		qs.quitChan <- true
	}
}
