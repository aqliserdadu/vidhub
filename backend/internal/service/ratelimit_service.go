package service

import (
	"sync"
	"time"

	"videodownload/internal/model"
	"videodownload/pkg/logger"

	"go.uber.org/zap"
)

// RateLimitEntry tracks request rate for an IP
type RateLimitEntry struct {
	IP       string
	Requests int
	ResetAt  time.Time
	Blocked  bool
}

// RateLimitService manages rate limiting for DDoS protection
type RateLimitService struct {
	cfg      *model.RateLimitConfig
	limits   map[string]*RateLimitEntry
	mu       sync.RWMutex
	quitChan chan bool
}

// NewRateLimitService creates a new rate limit service
func NewRateLimitService(cfg *model.RateLimitConfig) *RateLimitService {
	service := &RateLimitService{
		cfg:      cfg,
		limits:   make(map[string]*RateLimitEntry),
		quitChan: make(chan bool),
	}

	if cfg.Enabled {
		go service.cleanupRoutine()
	}

	return service
}

// IsAllowed checks if an IP is allowed to make a request
func (rls *RateLimitService) IsAllowed(ip string) bool {
	if !rls.cfg.Enabled {
		return true
	}

	rls.mu.Lock()
	defer rls.mu.Unlock()

	now := time.Now()
	entry, exists := rls.limits[ip]

	// Create new entry if not exists
	if !exists {
		rls.limits[ip] = &RateLimitEntry{
			IP:       ip,
			Requests: 1,
			ResetAt:  now.Add(time.Minute),
			Blocked:  false,
		}
		logger.Logger.Debug("New rate limit entry created", zap.String("ip", ip))
		return true
	}

	// Check if reset time has passed
	if now.After(entry.ResetAt) {
		entry.Requests = 1
		entry.ResetAt = now.Add(time.Minute)
		entry.Blocked = false
		return true
	}

	// Check if already blocked
	if entry.Blocked {
		logger.Logger.Warn("Request blocked - IP is rate limited", zap.String("ip", ip))
		return false
	}

	// Increment request counter
	entry.Requests++

	// Check if limit exceeded
	if entry.Requests > rls.cfg.RequestsPerMinute {
		entry.Blocked = true
		logger.Logger.Warn("Rate limit exceeded", zap.String("ip", ip), zap.Int("requests", entry.Requests), zap.Int("limit", rls.cfg.RequestsPerMinute))
		return false
	}

	logger.Logger.Debug("Request allowed", zap.String("ip", ip), zap.Int("requests", entry.Requests), zap.Int("limit", rls.cfg.RequestsPerMinute))
	return true
}

// AllowBurst checks if an IP is allowed with burst capacity
func (rls *RateLimitService) AllowBurst(ip string) bool {
	if !rls.cfg.Enabled {
		return true
	}

	rls.mu.Lock()
	defer rls.mu.Unlock()

	now := time.Now()
	entry, exists := rls.limits[ip]

	// Create new entry if not exists
	if !exists {
		rls.limits[ip] = &RateLimitEntry{
			IP:       ip,
			Requests: 1,
			ResetAt:  now.Add(time.Minute),
			Blocked:  false,
		}
		return true
	}

	// Check if reset time has passed
	if now.After(entry.ResetAt) {
		entry.Requests = 1
		entry.ResetAt = now.Add(time.Minute)
		entry.Blocked = false
		return true
	}

	// Check if blocked
	if entry.Blocked {
		return false
	}

	// Increment counter
	entry.Requests++

	// Use burst size as temporary limit
	limit := rls.cfg.RequestsPerMinute + rls.cfg.BurstSize
	if entry.Requests > limit {
		entry.Blocked = true
		logger.Logger.Warn("Burst limit exceeded, blocking IP", zap.String("ip", ip), zap.Int("requests", entry.Requests))
		return false
	}

	return true
}

// GetRemaining returns remaining requests for IP in current window
func (rls *RateLimitService) GetRemaining(ip string) int {
	if !rls.cfg.Enabled {
		return -1 // Unlimited
	}

	rls.mu.RLock()
	defer rls.mu.RUnlock()

	entry, exists := rls.limits[ip]
	if !exists {
		return rls.cfg.RequestsPerMinute
	}

	now := time.Now()
	if now.After(entry.ResetAt) {
		return rls.cfg.RequestsPerMinute
	}

	remaining := rls.cfg.RequestsPerMinute - entry.Requests
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// cleanupRoutine periodically cleans up old entries
func (rls *RateLimitService) cleanupRoutine() {
	ticker := time.NewTicker(time.Duration(rls.cfg.CleanupInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rls.quitChan:
			logger.Logger.Info("Rate limit service stopped")
			return
		case <-ticker.C:
			rls.cleanup()
		}
	}
}

// cleanup removes old entries
func (rls *RateLimitService) cleanup() {
	rls.mu.Lock()
	defer rls.mu.Unlock()

	now := time.Now()
	removed := 0

	for ip, entry := range rls.limits {
		// Remove entries with reset time older than 2 hours
		if now.Sub(entry.ResetAt) > 2*time.Hour {
			delete(rls.limits, ip)
			removed++
		}
	}

	if removed > 0 {
		logger.Logger.Debug("Rate limit entries cleaned up", zap.Int("removed", removed), zap.Int("remaining", len(rls.limits)))
	}
}

// Reset resets rate limit for a specific IP (admin operation)
func (rls *RateLimitService) Reset(ip string) {
	if !rls.cfg.Enabled {
		return
	}

	rls.mu.Lock()
	defer rls.mu.Unlock()

	delete(rls.limits, ip)
	logger.Logger.Info("Rate limit reset for IP", zap.String("ip", ip))
}

// Stop stops the rate limit service
func (rls *RateLimitService) Stop() {
	if rls.cfg.Enabled {
		rls.quitChan <- true
	}
}
