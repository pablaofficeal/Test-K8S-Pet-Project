package keystest

import (
	"Cadence-Cloud-Messenger/core/api/keys"
	"Cadence-Cloud-Messenger/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"encoding/json"
	"strings"
	"testing"
)

func router(user *string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if *user != "" {
			c.Set("userUUID", *user)
		}
		c.Next()
	})
	r.PUT("/key", keys.PutMyKey)
	r.GET("/key", keys.GetMyKey)
	r.PUT("/backup", keys.PutKeyBackup)
	r.GET("/user/:uuid", keys.GetUserKey)
	return r
}

func send(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
	return w
}

func TestKeyEndpointsRequireAuthentication(t *testing.T) {
	nobody := ""
	r := router(&nobody)
	for _, tc := range [][2]string{{"PUT", "/key"}, {"GET", "/key"}, {"PUT", "/backup"}, {"GET", "/user/" + uuid.NewString()}} {
		if w := send(r, tc[0], tc[1], map[string]any{}); w.Code != 401 {
			t.Errorf("%s %s: expected 401, got %d", tc[0], tc[1], w.Code)
		}
	}
}

func TestPutMyKeyValidation(t *testing.T) {
	user := uuid.NewString()
	r := router(&user)
	good := keyDoc()
	mutate := func(field string, value any) map[string]any {
		doc := keyDoc()
		doc[field] = value
		return doc
	}
	extra := keyDoc()
	extra["admin"] = true
	missing := keyDoc()
	delete(missing, "algorithm")
	for name, doc := range map[string]any{
		"empty":          map[string]any{},
		"extra field":    extra,
		"missing field":  missing,
		"version 2":      mutate("version", 2),
		"algorithm":      mutate("algorithm", "RSA"),
		"short key":      mutate("public_key", "AAAA"),
		"not base64":     mutate("public_key", "!!!!"),
		"wrong key id":   mutate("key_id", strings.Repeat("a", 32)),
		"numeric id":     mutate("key_id", 5),
		"string version": mutate("version", "1"),
	} {
		if w := send(r, "PUT", "/key", doc); w.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
	_ = good
}

func TestPutMyKeyAgainstPostgres(t *testing.T) {
	tx := beginTx(t)
	alice, bob := newUser(tx, "alice-k"), newUser(tx, "bob-k")
	current := alice
	r := router(&current)

	first := keyDoc()
	if w := send(r, "PUT", "/key", first); w.Code != 201 {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var mine struct {
		KeyID string `json:"key_id"`
		PublicKey map[string]any `json:"public_key"`
		Backup *string
	}
	w := send(r, "GET", "/key", nil)
	json.Unmarshal(w.Body.Bytes(), &mine)
	if w.Code != 200 || mine.KeyID != first["key_id"] || mine.PublicKey["public_key"] != first["public_key"] || mine.Backup != nil {
		t.Errorf("get my key: %d %s", w.Code, w.Body.String())
	}

	if w := send(r, "PUT", "/backup", map[string]any{"backup": "opaque-blob"}); w.Code != 204 {
		t.Errorf("backup: %d", w.Code)
	}
	second := keyDoc()
	if w := send(r, "PUT", "/key", second); w.Code != 200 {
		t.Errorf("rotation: expected 200, got %d", w.Code)
	}
	w = send(r, "GET", "/key", nil)
	json.Unmarshal(w.Body.Bytes(), &mine)
	if mine.KeyID != second["key_id"] || mine.Backup == nil || *mine.Backup != "opaque-blob" {
		t.Errorf("rotation must replace the key and keep the backup: %s", w.Body.String())
	}

	current = bob
	if w := send(r, "PUT", "/key", second); w.Code != 409 {
		t.Errorf("registering someone else's key: expected 409, got %d", w.Code)
	}

	var stored models.MessengerKey
	tx.First(&stored, "user_uuid = ?", alice)
	if stored.KeyID != second["key_id"] {
		t.Error("stored key must be the rotated one")
	}
}