package messagestest

import (
	"Cadence-Cloud-Messenger/core/api/messages"
	"Cadence-Cloud-Messenger/core/api/keys"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"encoding/json"
	"strings"
)

type world struct {
	r *gin.Engine
	as *string
}

func newWorld() *world {
	gin.SetMode(gin.TestMode)
	user := ""
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if user != "" {
			c.Set("userUUID", user)
		}
		c.Next()
	})
	r.PUT("/key", keys.PutMyKey)
	r.POST("/send", messages.SendMessage)
	r.GET("/messages", messages.ListMessages)
	r.GET("/conversations", messages.ListConversations)
	r.PATCH("/read", messages.MarkRead)
	r.DELETE("/message/:id", messages.DeleteMessage)
	w := &world{r: r, as: &user}
	return w
}

func (w *world) call(user, method, path string, body any) *httptest.ResponseRecorder {
	*w.as = user
	raw, _ := json.Marshal(body)
	rec := httptest.NewRecorder()
	w.r.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
	return rec
}

func envelope(senderKey, recipientKey string) map[string]any {
	return map[string]any{
		"type": "my-masengers/message", "version": 2,
		"algorithm": "X25519-HKDF-SHA256+DualAlphabetNoise-FullRandom-v2",
		"alphabet_hash": "aaaa", "alphabet_hash2": "bbbb",
		"sender_key_id": senderKey, "recipient_key_id": recipientKey,
		"salt": "c2FsdA", "nonce": "bm9uY2U", "ciphertext": "tok1 tok2 tok3", "auth_tag": "dGFn",
	}
}