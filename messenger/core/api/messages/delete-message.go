package messages

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func DeleteMessage(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}
	result := db.DB.WithContext(c.Request.Context()).Where("id = ? AND sender_uuid = ?", id, userUUID).Delete(&models.MessengerMessage{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	c.Status(http.StatusNoContent)
}