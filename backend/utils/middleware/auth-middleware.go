package middleware


import (
	"Test-K8s-Pet-Project-Backend/utils/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"log"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("trace_id=%s authorization header is missing", traceID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("trace_id=%s authorization header has an invalid format", traceID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			log.Printf("trace_id=%s error validating token: %v", traceID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Authorized"})
			c.Abort()
			return
		}

		c.Set("userUUID", claims.UUID)
		c.Set("userEmail", claims.Email)

		c.Next()
	}
}