package userstest

import (
	"Cadence-Cloud-Messenger/core/api/users"
	"Cadence-Cloud-Messenger/core/api/keys"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"encoding/json"
	"net/url"
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
	r.GET("/search", users.SearchUsers)
	r.PUT("/key", keys.PutMyKey)
	return r
}

func search(r *gin.Engine, q string) (int, []map[string]any) {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/search?q="+url.QueryEscape(q), nil))
	var body struct{ Users []map[string]any }
	json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body.Users
}

func TestSearchRequiresAuthentication(t *testing.T) {
	nobody := ""
	if code, _ := search(router(&nobody), "alice"); code != 401 {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestSearchQueryValidation(t *testing.T) {
	user := "11111111-1111-1111-1111-111111111111"
	r := router(&user)
	for _, q := range []string{"", "a", " a ", strings.Repeat("x", 65)} {
		if code, _ := search(r, q); code != 400 {
			t.Errorf("query %q: expected 400, got %d", q, code)
		}
	}
}

func TestSearchAgainstPostgres(t *testing.T) {
	tx := beginTx(t)
	me := newUser(tx, "zed-me")
	alice := newUser(tx, "Zed-Alice")
	bob := newUser(tx, "zed-bob")
	inactive := newUser(tx, "zed-inactive")
	tx.Exec("UPDATE users SET is_active = false WHERE uuid = ?", inactive)
	newUser(tx, "zed_underscore")
	newUser(tx, "other-person")

	current := alice
	r := router(&current)
	doc := keyDoc()
	if w := httptest.NewRecorder(); true {
		raw, _ := json.Marshal(doc)
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/key", strings.NewReader(string(raw))))
	}

	current = me
	code, found := search(r, "ZED")
	if code != 200 {
		t.Fatalf("search: %d", code)
	}
	names := map[string]map[string]any{}
	for _, hit := range found {
		names[hit["username"].(string)] = hit
	}
	if len(names) != 4 || names["zed-me"] != nil || names["zed-inactive"] != nil {
		t.Errorf("must exclude me and inactive users, got %v", names)
	}
	if names["Zed-Alice"]["key_id"] != doc["key_id"] || names["zed-bob"]["key_id"] != nil {
		t.Errorf("key_id must be set only for users with a key: %v", names)
	}
	if names["Zed-Alice"]["uuid"] != alice || names["zed-bob"]["uuid"] != bob {
		t.Error("uuid mismatch")
	}
	if _, leaked := found[0]["email"]; leaked {
		t.Error("email must not be exposed")
	}

	if _, found := search(r, "%%"); len(found) != 0 {
		t.Errorf("a wildcard must not match everything: %v", found)
	}
	if _, found := search(r, "zed_"); len(found) != 1 || found[0]["username"] != "zed_underscore" {
		t.Errorf("underscore must be literal: %v", found)
	}
	if _, found := search(r, "person"); len(found) != 0 {
		t.Errorf("search is by prefix, got %v", found)
	}
}

func TestSearchLimitsResults(t *testing.T) {
	tx := beginTx(t)
	me := newUser(tx, "lim-me")
	for i := 0; i < 25; i++ {
		newUser(tx, "limit-user-"+string(rune('a'+i)))
	}
	r := router(&me)
	if _, found := search(r, "limit-user"); len(found) != 20 {
		t.Errorf("expected 20 results, got %d", len(found))
	}
}