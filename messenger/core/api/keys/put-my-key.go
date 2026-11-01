package keys

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"encoding/base64"
	"encoding/json"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"log"
)

func PutMyKey(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	var doc map[string]any
	if err := c.ShouldBindJSON(&doc); err != nil || len(doc) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public key document"})
		return
	}
	version, okVersion := doc["version"].(float64)
	algorithm, okAlgorithm := doc["algorithm"].(string)
	keyID, okID := doc["key_id"].(string)
	encoded, okKey := doc["public_key"].(string)
	if !okVersion || !okAlgorithm || !okID || !okKey || version != 1 || algorithm != "X25519" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported public key format"})
		return
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(raw) != 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "public key must be 32 bytes in base64url"})
		return
	}
	digest := sha256.Sum256(raw)
	if keyID != hex.EncodeToString(digest[:])[:32] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key_id does not match the public key"})
		return
	}
	stored, _ := json.Marshal(doc)

	ctx := c.Request.Context()
	var taken models.MessengerKey
	if err := db.DB.WithContext(ctx).First(&taken, "key_id = ? AND user_uuid <> ?", keyID, userUUID).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "this key belongs to another user"})
		return
	}
	var current models.MessengerKey
	if err := db.DB.WithContext(ctx).First(&current, "user_uuid = ?", userUUID).Error; err != nil {
		current = models.MessengerKey{UserUUID: userUUID.(string), KeyID: keyID, PublicKey: string(stored)}
		if err := db.DB.WithContext(ctx).Create(&current).Error; err != nil {
			log.Printf("trace_id=%s create key error: %v", traceID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"key_id": keyID})
		return
	}
	update := map[string]any{"key_id": keyID, "public_key": string(stored)}
	if err := db.DB.WithContext(ctx).Model(&current).Updates(update).Error; err != nil {
		log.Printf("trace_id=%s update key error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key_id": keyID})
}