package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"context"
)

func TraceMiddleware() gin.HandlerFunc {
	type traceIDContextKey struct{}
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Request-ID")
		valid := len(traceID) > 0 && len(traceID) <= 128
		for _, char := range traceID {
			if (char < 'a' || char > 'z') &&
				(char < 'A' || char > 'Z') &&
				(char < '0' || char > '9') &&
				char != '-' && char != '_' && char != '.' {
				valid = false
				break
			}
		}
		if !valid {
			traceID = uuid.NewString()
		}

		ctx := context.WithValue(c.Request.Context(), traceIDContextKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Request.Header.Set("X-Request-ID", traceID)
		c.Set("trace_id", traceID)
		c.Header("X-Request-ID", traceID)

		c.Next()
	}
}