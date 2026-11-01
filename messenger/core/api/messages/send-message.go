package messages

import (
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"log"
)

type SendMessageRequest struct {
	RecipientUUID string `json:"recipient_uuid"`
	Envelope map[string]any `json:"envelope"`
}

func SendMessage(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	var body SendMessageRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Envelope == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_uuid and envelope are required"})
		return
	}
	if _, err := uuid.Parse(body.RecipientUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipient id"})
		return
	}

	envelope := body.Envelope
	required := []string{"type", "version", "algorithm", "alphabet_hash", "alphabet_hash2", "sender_key_id", "recipient_key_id", "salt", "nonce", "ciphertext", "auth_tag"}
	if len(envelope) != len(required) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "envelope has missing or extra fields"})
		return
	}
	for _, name := range required {
		if _, ok := envelope[name]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "envelope has missing or extra fields"})
			return
		}
	}
	if envelope["type"] != "my-masengers/message" || envelope["version"] != float64(2) || envelope["algorithm"] != "X25519-HKDF-SHA256+DualAlphabetNoise-FullRandom-v2" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported envelope version or algorithm"})
		return
	}
	for _, name := range []string{"alphabet_hash", "alphabet_hash2", "salt", "nonce", "ciphertext", "auth_tag", "sender_key_id", "recipient_key_id"} {
		value, ok := envelope[name].(string)
		if !ok || value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "envelope field " + name + " must be a non-empty string"})
			return
		}
		if name != "ciphertext" && strings.ContainsAny(value, " \t\r\n") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "envelope field " + name + " must not contain spaces"})
			return
		}
	}
	senderKeyID, recipientKeyID := envelope["sender_key_id"].(string), envelope["recipient_key_id"].(string)
	if len(senderKeyID) != 32 || len(recipientKeyID) != 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}
	stored, _ := json.Marshal(envelope)
	if len(stored) > 1<<20 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "message is larger than 1MB"})
		return
	}

	ctx := c.Request.Context()
	var senderKey, recipientKey models.MessengerKey
	if err := db.DB.WithContext(ctx).First(&senderKey, "user_uuid = ?", userUUID).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "register your key first"})
		return
	}
	if err := db.DB.WithContext(ctx).First(&recipientKey, "user_uuid = ?", body.RecipientUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipient has no key"})
		return
	}
	if senderKey.KeyID != senderKeyID {
		c.JSON(http.StatusConflict, gin.H{"error": "sender key does not match your registered key"})
		return
	}
	if recipientKey.KeyID != recipientKeyID {
		c.JSON(http.StatusConflict, gin.H{"error": "recipient key has changed, refresh it and encrypt again"})
		return
	}

	var recent int64
	db.DB.WithContext(ctx).Model(&models.MessengerMessage{}).Where("sender_uuid = ? AND created_at > ?", userUUID, time.Now().Add(-time.Minute)).Count(&recent)
	if recent >= 120 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many messages, slow down"})
		return
	}

	message := models.MessengerMessage{SenderUUID: userUUID.(string), RecipientUUID: body.RecipientUUID, Envelope: string(stored)}
	if err := db.DB.WithContext(ctx).Create(&message).Error; err != nil {
		log.Printf("trace_id=%s store message error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": message.ID, "created_at": message.CreatedAt})
}