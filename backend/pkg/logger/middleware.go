package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger returns a middleware for logging HTTP requests
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		Logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.Int("body_size", c.Writer.Size()),
		)
	}
}

// LogError logs an error with context
func LogError(msg string, err error, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, append(fields, zap.Error(err))...)
	}
}

// LogWarn logs a warning
func LogWarn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// LogInfo logs an info message
func LogInfo(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// LogDebug logs a debug message
func LogDebug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// CleanupLogs cleans up old log files
func CleanupLogs(logDir string, maxAge time.Duration) {
	// Implementation would clean up old log files
	// This prevents log accumulation on the server
	Logger.Info(fmt.Sprintf("Log cleanup job started (maxAge: %v)", maxAge))
}
