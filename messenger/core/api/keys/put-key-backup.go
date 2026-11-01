package keys

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

type PutKeyBackupRequest struct {
	Backup string `json:"backup"`
}

func PutKeyBackup(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	var body PutKeyBackupRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Backup == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup is required"})
		return
	}
	if len(body.Backup) > 65536 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "backup is larger than 64KB"})
		return
	}
	result := db.DB.WithContext(c.Request.Context()).Model(&models.MessengerKey{}).Where("user_uuid = ?", userUUID).Update("backup", body.Backup)
	if result.Error != nil {
		log.Printf("trace_id=%s save backup error: %v", traceID, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "register a public key first"})
		return
	}
	c.Status(http.StatusNoContent)
}