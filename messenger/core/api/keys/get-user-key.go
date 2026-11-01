package keys

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"encoding/json"
	"net/http"
)

func GetUserKey(c *gin.Context) {
	if _, exists := c.Get("userUUID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	target := c.Param("uuid")
	if _, err := uuid.Parse(target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var key models.MessengerKey
	if err := db.DB.WithContext(c.Request.Context()).First(&key, "user_uuid = ?", target).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user has no key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_uuid": key.UserUUID, "key_id": key.KeyID, "public_key": json.RawMessage(key.PublicKey)})
}