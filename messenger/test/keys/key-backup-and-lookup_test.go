package keystest

import (
	"github.com/google/uuid"
	"encoding/json"
	"strings"
	"testing"
)

func TestBackupRules(t *testing.T) {
	tx := beginTx(t)
	user := newUser(tx, "backup-user")
	r := router(&user)

	if w := send(r, "PUT", "/backup", map[string]any{"backup": "blob"}); w.Code != 404 {
		t.Errorf("backup before a key: expected 404, got %d", w.Code)
	}
	if w := send(r, "GET", "/key", nil); w.Code != 404 {
		t.Errorf("get my key before registering: expected 404, got %d", w.Code)
	}
	send(r, "PUT", "/key", keyDoc())
	for name, body := range map[string]any{"empty": map[string]any{"backup": ""}, "missing": map[string]any{}, "wrong type": map[string]any{"backup": 5}} {
		if w := send(r, "PUT", "/backup", body); w.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
	if w := send(r, "PUT", "/backup", map[string]any{"backup": strings.Repeat("x", 65537)}); w.Code != 413 {
		t.Errorf("oversized backup: expected 413, got %d", w.Code)
	}
	if w := send(r, "PUT", "/backup", map[string]any{"backup": strings.Repeat("x", 65536)}); w.Code != 204 {
		t.Errorf("max size backup: expected 204, got %d", w.Code)
	}
}

func TestUserKeyLookup(t *testing.T) {
	tx := beginTx(t)
	alice, bob := newUser(tx, "alice-l"), newUser(tx, "bob-l")
	current := alice
	r := router(&current)
	doc := keyDoc()
	send(r, "PUT", "/key", doc)
	send(r, "PUT", "/backup", map[string]any{"backup": "secret-blob"})

	current = bob
	w := send(r, "GET", "/user/"+alice, nil)
	if w.Code != 200 {
		t.Fatalf("lookup: %d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["key_id"] != doc["key_id"] || body["user_uuid"] != alice {
		t.Errorf("lookup body: %v", body)
	}
	if strings.Contains(w.Body.String(), "secret-blob") || strings.Contains(w.Body.String(), "backup") {
		t.Error("another user must never see the backup blob")
	}
	if w := send(r, "GET", "/user/"+uuid.NewString(), nil); w.Code != 404 {
		t.Errorf("unknown user: %d", w.Code)
	}
	if w := send(r, "GET", "/user/nope", nil); w.Code != 400 {
		t.Errorf("bad id: %d", w.Code)
	}
}