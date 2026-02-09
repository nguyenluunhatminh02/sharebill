package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger returns a Gin middleware that logs each HTTP request
// with structured logging using zap
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// Get user ID if authenticated
		userID := ""
		if uid, exists := c.Get("firebase_uid"); exists {
			userID = uid.(string)
		}

		// Build log fields
		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Log any errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				fields = append(fields, zap.String("error", e.Error()))
			}
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			logger.Error("Server error", fields...)
		case statusCode >= 400:
			logger.Warn("Client error", fields...)
		case statusCode >= 300:
			logger.Info("Redirect", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}

// RecoveryWithLogger returns a Gin middleware that recovers from panics
// and logs the error with zap
func RecoveryWithLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
