package profile

import (
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/mail"
	"strings"
	"log"
)

type EditEmailRequest struct {
	Email string `json:"email"`
}

func EditMyEmail(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		log.Printf("trace_id=%s user UUID not found in context", traceID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}

	var body EditEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	email := strings.TrimSpace(body.Email)
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	ctx := c.Request.Context()
	var taken int64
	if err := db.DB.WithContext(ctx).Model(&models.User{}).Where("LOWER(email) = LOWER(?) AND uuid <> ?", email, userUUID).Count(&taken).Error; err != nil {
		log.Printf("trace_id=%s error checking email: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if taken > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		return
	}

	result := db.DB.WithContext(ctx).Model(&models.User{}).Where("uuid = ?", userUUID).Update("email", email)
	if result.Error != nil {
		log.Printf("trace_id=%s error updating email: %v", traceID, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"email": email})
}