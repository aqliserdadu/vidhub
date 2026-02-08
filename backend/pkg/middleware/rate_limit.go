package middleware

import (
	"fmt"
	"net/http"

	"videodownload/internal/service"
	"videodownload/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(rateLimitService *service.RateLimitService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Check rate limit
		if !rateLimitService.IsAllowed(ip) {
			logger.Logger.Warn("Rate limit exceeded", zap.String("ip", ip))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
				"code":    http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		// Set remaining requests header
		remaining := rateLimitService.GetRemaining(ip)
		if remaining >= 0 {
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		}

		c.Next()
	}
}

// QuotaCheckMiddleware creates a middleware to check download quota
func QuotaCheckMiddleware(quotaService *service.QuotaService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check quota on download endpoints
		if c.Request.URL.Path == "/api/download" && c.Request.Method == "POST" {
			ip := c.ClientIP()

			// Check quota info for logging
			quotaInfo := quotaService.GetQuotaInfo(ip)
			c.Set("quota_info", quotaInfo)

			logger.Logger.Debug("Quota check", zap.String("ip", ip), zap.Any("quota_info", quotaInfo))
		}

		c.Next()
	}
}
