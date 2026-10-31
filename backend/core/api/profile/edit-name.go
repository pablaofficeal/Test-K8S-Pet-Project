package profile

import (
	"Low-Code-Bot-Platform-Backend-Service/models"
	"Low-Code-Bot-Platform-Backend-Service/db"
	"github.com/gin-gonic/gin"
	"unicode/utf8"
	"net/http"
	"strings"
	"log"
)

type EditNameRequest struct {
	Username string `json:"username"`
}

func EditMyName(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		log.Printf("trace_id=%s user UUID not found in context", traceID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}

	var body EditNameRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" || utf8.RuneCountInString(username) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be 1-64 characters"})
		return
	}

	ctx := c.Request.Context()
	var taken int64
	if err := db.DB.WithContext(ctx).Model(&models.User{}).Where("username = ? AND uuid <> ?", username, userUUID).Count(&taken).Error; err != nil {
		log.Printf("trace_id=%s error checking username: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if taken > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	result := db.DB.WithContext(ctx).Model(&models.User{}).Where("uuid = ?", userUUID).Update("username", username)
	if result.Error != nil {
		log.Printf("trace_id=%s error updating username: %v", traceID, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": username})
}