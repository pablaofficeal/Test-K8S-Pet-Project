package middleware

import (
	"Test-K8s-Pet-Project-Verify/utils/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"log"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("trace_id=%s authorization header is missing or malformed", traceID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(parts[1])
		if err != nil {
			log.Printf("trace_id=%s error validating token: %v", traceID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
			c.Abort()
			return
		}

		c.Set("userUUID", claims.UUID)
		c.Next()
	}
}