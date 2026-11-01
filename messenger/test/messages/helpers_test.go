package messagestest

import (
	"Cadence-Cloud-Messenger/db"
	"github.com/google/uuid"
	"encoding/base64"
	"crypto/sha256"
	"encoding/hex"
	"gorm.io/gorm"
	"crypto/rand"
	"testing"
)

func beginTx(t *testing.T) *gorm.DB {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })
	return tx
}

func keyDoc() map[string]any {
	raw := make([]byte, 32)
	rand.Read(raw)
	digest := sha256.Sum256(raw)
	return map[string]any{"version": 1, "algorithm": "X25519", "key_id": hex.EncodeToString(digest[:])[:32], "public_key": base64.RawURLEncoding.EncodeToString(raw)}
}

func newUser(tx *gorm.DB, name string) string {
	id := uuid.NewString()
	tx.Exec("INSERT INTO users (uuid, username, email, password, is_active) VALUES (?, ?, ?, 'x', true)", id, name, id+"@test.local")
	return id
}