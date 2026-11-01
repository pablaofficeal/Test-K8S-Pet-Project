package keys

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"net/http"
)

func GetMyKey(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	var key models.MessengerKey
	if err := db.DB.WithContext(c.Request.Context()).First(&key, "user_uuid = ?", userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key is not registered"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key_id": key.KeyID, "public_key": json.RawMessage(key.PublicKey), "backup": key.Backup})
}