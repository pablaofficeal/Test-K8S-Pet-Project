package users

import (
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"log"
)

type userHit struct {
	UUID string `json:"uuid"`
	Username string `json:"username"`
	KeyID *string `json:"key_id"`
}

func SearchUsers(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	if len([]rune(query)) < 2 || len(query) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query must be 2-64 characters"})
		return
	}
	pattern := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(strings.ToLower(query)) + "%"
	hits := []userHit{}
	err := db.DB.WithContext(c.Request.Context()).Raw(
		"SELECT u.uuid::text AS uuid, u.username AS username, k.key_id AS key_id FROM users u LEFT JOIN messenger_keys k ON k.user_uuid = u.uuid WHERE u.uuid <> ? AND u.is_active AND LOWER(u.username) LIKE ? ESCAPE '\\' ORDER BY u.username LIMIT 20",
		userUUID, pattern,
	).Scan(&hits).Error
	if err != nil {
		log.Printf("trace_id=%s search users error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": hits})
}