package messagestest

import (
	"github.com/google/uuid"
	"strings"
	"testing"
)

func register(w *world, user string) string {
	doc := keyDoc()
	w.call(user, "PUT", "/key", doc)
	return doc["key_id"].(string)
}

func TestSendRequiresAuthentication(t *testing.T) {
	w := newWorld()
	if rec := w.call("", "POST", "/send", map[string]any{}); rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestSendValidation(t *testing.T) {
	tx := beginTx(t)
	w := newWorld()
	alice, bob := newUser(tx, "alice-s"), newUser(tx, "bob-s")
	aliceKey, bobKey := register(w, alice), register(w, bob)

	mutate := func(field string, value any) map[string]any {
		e := envelope(aliceKey, bobKey)
		e[field] = value
		return e
	}
	extra := envelope(aliceKey, bobKey)
	extra["hmm"] = "x"
	missing := envelope(aliceKey, bobKey)
	delete(missing, "salt")
	cases := map[string]map[string]any{
		"extra field":      extra,
		"missing field":    missing,
		"old version":      mutate("version", 1),
		"other algorithm":  mutate("algorithm", "AES"),
		"other type":       mutate("type", "x"),
		"empty ciphertext": mutate("ciphertext", ""),
		"number tag":       mutate("auth_tag", 5),
		"spaces in salt":   mutate("salt", "a b"),
		"short key id":     mutate("sender_key_id", "abc"),
	}
	for name, e := range cases {
		if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": e}); rec.Code != 400 {
			t.Errorf("%s: expected 400, got %d %s", name, rec.Code, rec.Body.String())
		}
	}
	for name, body := range map[string]map[string]any{
		"no envelope":  {"recipient_uuid": bob},
		"bad uuid":     {"recipient_uuid": "x", "envelope": envelope(aliceKey, bobKey)},
		"empty uuid":   {"envelope": envelope(aliceKey, bobKey)},
	} {
		if rec := w.call(alice, "POST", "/send", body); rec.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, rec.Code)
		}
	}
}

func TestSendRules(t *testing.T) {
	tx := beginTx(t)
	w := newWorld()
	alice, bob, carol := newUser(tx, "alice-r"), newUser(tx, "bob-r"), newUser(tx, "carol-r")
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": envelope(strings.Repeat("a", 32), strings.Repeat("b", 32))}); rec.Code != 409 {
		t.Errorf("sender without a key: expected 409, got %d", rec.Code)
	}
	aliceKey := register(w, alice)
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": envelope(aliceKey, strings.Repeat("b", 32))}); rec.Code != 404 {
		t.Errorf("recipient without a key: expected 404, got %d", rec.Code)
	}
	bobKey := register(w, bob)
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": uuid.NewString(), "envelope": envelope(aliceKey, bobKey)}); rec.Code != 404 {
		t.Errorf("unknown recipient: expected 404, got %d", rec.Code)
	}
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": envelope(strings.Repeat("c", 32), bobKey)}); rec.Code != 409 {
		t.Errorf("foreign sender key id: expected 409, got %d", rec.Code)
	}
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": envelope(aliceKey, strings.Repeat("d", 32))}); rec.Code != 409 || !strings.Contains(rec.Body.String(), "changed") {
		t.Errorf("stale recipient key: expected 409, got %d %s", rec.Code, rec.Body.String())
	}
	register(w, carol)
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": envelope(aliceKey, bobKey)}); rec.Code != 201 {
		t.Errorf("valid message: expected 201, got %d %s", rec.Code, rec.Body.String())
	}
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": alice, "envelope": envelope(aliceKey, aliceKey)}); rec.Code != 201 {
		t.Errorf("a note to yourself: expected 201, got %d", rec.Code)
	}
}

func TestSendLimitsSizeAndRate(t *testing.T) {
	tx := beginTx(t)
	w := newWorld()
	alice, bob := newUser(tx, "alice-z"), newUser(tx, "bob-z")
	aliceKey, bobKey := register(w, alice), register(w, bob)

	big := envelope(aliceKey, bobKey)
	big["ciphertext"] = strings.Repeat("x", 1<<20)
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": big}); rec.Code != 413 {
		t.Errorf("oversized message: expected 413, got %d", rec.Code)
	}
	for i := 0; i < 120; i++ {
		tx.Exec("INSERT INTO messenger_messages (sender_uuid, recipient_uuid, envelope, created_at) VALUES (?, ?, '{}', now())", alice, bob)
	}
	if rec := w.call(alice, "POST", "/send", map[string]any{"recipient_uuid": bob, "envelope": envelope(aliceKey, bobKey)}); rec.Code != 429 {
		t.Errorf("rate limit: expected 429, got %d", rec.Code)
	}
	if rec := w.call(bob, "POST", "/send", map[string]any{"recipient_uuid": alice, "envelope": envelope(bobKey, aliceKey)}); rec.Code != 201 {
		t.Errorf("another sender must not be limited: got %d", rec.Code)
	}
}