package messages

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
	"log"
)

type MarkReadRequest struct {
	Peer string `json:"peer"`
}

func MarkRead(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	var body MarkReadRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if _, err := uuid.Parse(body.Peer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer id"})
		return
	}
	result := db.DB.WithContext(c.Request.Context()).Model(&models.MessengerMessage{}).
		Where("recipient_uuid = ? AND sender_uuid = ? AND read_at IS NULL", userUUID, body.Peer).Update("read_at", time.Now())
	if result.Error != nil {
		log.Printf("trace_id=%s mark read error: %v", traceID, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"marked": result.RowsAffected})
}