package messages

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"log"
)

type messageView struct {
	ID int64 `json:"id"`
	SenderUUID string `json:"sender_uuid"`
	RecipientUUID string `json:"recipient_uuid"`
	Envelope json.RawMessage `json:"envelope"`
	ReadAt *time.Time `json:"read_at"`
	CreatedAt time.Time `json:"created_at"`
}

func ListMessages(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	peer := c.Query("peer")
	if _, err := uuid.Parse(peer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer id"})
		return
	}
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be 1-200"})
			return
		}
		limit = parsed
	}

	query := db.DB.WithContext(c.Request.Context()).Model(&models.MessengerMessage{}).
		Where("(sender_uuid = ? AND recipient_uuid = ?) OR (sender_uuid = ? AND recipient_uuid = ?)", userUUID, peer, peer, userUUID)
	rows := []models.MessengerMessage{}
	if raw := c.Query("after"); raw != "" {
		after, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || after < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid after cursor"})
			return
		}
		err = query.Where("id > ?", after).Order("id ASC").Limit(limit).Find(&rows).Error
		if err != nil {
			log.Printf("trace_id=%s list messages error: %v", traceID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	} else {
		if err := query.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
			log.Printf("trace_id=%s list messages error: %v", traceID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}
	views := make([]messageView, len(rows))
	for i, row := range rows {
		views[i] = messageView{ID: row.ID, SenderUUID: row.SenderUUID, RecipientUUID: row.RecipientUUID, Envelope: json.RawMessage(row.Envelope), ReadAt: row.ReadAt, CreatedAt: row.CreatedAt}
	}
	c.JSON(http.StatusOK, gin.H{"messages": views})
}