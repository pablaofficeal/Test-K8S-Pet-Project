package profile

import (
	"Low-Code-Bot-Platform-Backend-Service/models"
	"Low-Code-Bot-Platform-Backend-Service/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

func GetProfileInfo(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		log.Printf("trace_id=%s user UUID not found in context", traceID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}

	var user models.User
	if result := db.DB.WithContext(c.Request.Context()).First(&user, "uuid = ?", userUUID); result.Error != nil {
		log.Printf("trace_id=%s error fetching user profile: %v", traceID, result.Error)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uuid": user.UUID,
		"username": user.Username,
		"email": user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}