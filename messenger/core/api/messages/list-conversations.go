package messages

import (
	"Cadence-Cloud-Messenger/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"log"
)

type conversationView struct {
	PeerUUID string `json:"peer_uuid"`
	Username string `json:"username"`
	KeyID *string `json:"key_id"`
	Unread int64 `json:"unread"`
	LastID int64 `json:"last_id"`
	LastAt time.Time `json:"last_at"`
}

func ListConversations(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	items := []conversationView{}
	err := db.DB.WithContext(c.Request.Context()).Raw(`
		SELECT t.peer::text AS peer_uuid, COALESCE(u.username, '') AS username, k.key_id AS key_id,
			COUNT(*) FILTER (WHERE t.recipient_uuid = ? AND t.read_at IS NULL) AS unread,
			MAX(t.id) AS last_id, MAX(t.created_at) AS last_at
		FROM (
			SELECT CASE WHEN sender_uuid = ? THEN recipient_uuid ELSE sender_uuid END AS peer, recipient_uuid, read_at, id, created_at
			FROM messenger_messages WHERE sender_uuid = ? OR recipient_uuid = ?
		) t
		LEFT JOIN users u ON u.uuid = t.peer
		LEFT JOIN messenger_keys k ON k.user_uuid = t.peer
		GROUP BY t.peer, u.username, k.key_id
		ORDER BY last_id DESC LIMIT 100`,
		userUUID, userUUID, userUUID, userUUID,
	).Scan(&items).Error
	if err != nil {
		log.Printf("trace_id=%s list conversations error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": items})
}