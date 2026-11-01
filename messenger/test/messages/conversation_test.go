package messagestest

import (
	"github.com/google/uuid"
	"encoding/json"
	"strconv"
	"testing"
)

func send(t *testing.T, w *world, from, to, fromKey, toKey string) int64 {
	rec := w.call(from, "POST", "/send", map[string]any{"recipient_uuid": to, "envelope": envelope(fromKey, toKey)})
	if rec.Code != 201 {
		t.Fatalf("send: %d %s", rec.Code, rec.Body.String())
	}
	var body struct{ ID int64 }
	json.Unmarshal(rec.Body.Bytes(), &body)
	return body.ID
}

func ids(rec interface{ Bytes() []byte }) []int64 {
	var body struct{ Messages []struct{ ID int64 } }
	json.Unmarshal(rec.Bytes(), &body)
	out := []int64{}
	for _, m := range body.Messages {
		out = append(out, m.ID)
	}
	return out
}

func TestListMessages(t *testing.T) {
	tx := beginTx(t)
	w := newWorld()
	alice, bob, carol := newUser(tx, "alice-m"), newUser(tx, "bob-m"), newUser(tx, "carol-m")
	ak, bk, ck := register(w, alice), register(w, bob), register(w, carol)
	m1 := send(t, w, alice, bob, ak, bk)
	m2 := send(t, w, bob, alice, bk, ak)
	m3 := send(t, w, alice, bob, ak, bk)
	send(t, w, alice, carol, ak, ck)
	send(t, w, carol, bob, ck, bk)

	if rec := w.call("", "GET", "/messages?peer="+bob, nil); rec.Code != 401 {
		t.Errorf("unauthenticated: %d", rec.Code)
	}
	for _, q := range []string{"", "?peer=x", "?peer=" + bob + "&limit=0", "?peer=" + bob + "&limit=201", "?peer=" + bob + "&after=-1", "?peer=" + bob + "&after=x"} {
		if rec := w.call(alice, "GET", "/messages"+q, nil); rec.Code != 400 {
			t.Errorf("query %q: expected 400, got %d", q, rec.Code)
		}
	}

	rec := w.call(alice, "GET", "/messages?peer="+bob, nil)
	got := ids(rec.Body)
	if len(got) != 3 || got[0] != m1 || got[1] != m2 || got[2] != m3 {
		t.Errorf("thread must hold only the pair, oldest first: %v", got)
	}
	var body struct{ Messages []struct{ Envelope map[string]any } }
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Messages[0].Envelope["ciphertext"] != "tok1 tok2 tok3" {
		t.Error("envelope must come back unchanged")
	}

	if got := ids(w.call(alice, "GET", "/messages?peer="+bob+"&limit=2", nil).Body); len(got) != 2 || got[0] != m2 || got[1] != m3 {
		t.Errorf("limit returns the newest messages: %v", got)
	}
	if got := ids(w.call(bob, "GET", "/messages?peer="+alice+"&after="+strconv.FormatInt(m1, 10), nil).Body); len(got) != 2 || got[0] != m2 || got[1] != m3 {
		t.Errorf("after cursor: %v", got)
	}
	if got := ids(w.call(alice, "GET", "/messages?peer="+bob+"&after="+strconv.FormatInt(m3, 10), nil).Body); len(got) != 0 {
		t.Errorf("nothing after the last message: %v", got)
	}
	if got := ids(w.call(carol, "GET", "/messages?peer="+alice, nil).Body); len(got) != 1 {
		t.Errorf("each side sees the same pair thread: %v", got)
	}
	if got := ids(w.call(carol, "GET", "/messages?peer="+uuid.NewString(), nil).Body); len(got) != 0 {
		t.Errorf("unknown peer: %v", got)
	}
}

func TestConversationsReadAndDelete(t *testing.T) {
	tx := beginTx(t)
	w := newWorld()
	alice, bob, carol := newUser(tx, "alice-c"), newUser(tx, "bob-c"), newUser(tx, "carol-c")
	ak, bk, ck := register(w, alice), register(w, bob), register(w, carol)
	send(t, w, bob, alice, bk, ak)
	send(t, w, bob, alice, bk, ak)
	last := send(t, w, carol, alice, ck, ak)
	send(t, w, alice, bob, ak, bk)

	var body struct {
		Conversations []struct {
			PeerUUID string `json:"peer_uuid"`
			Username string `json:"username"`
		}
	}
	rec := w.call(alice, "GET", "/conversations", nil)
	json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Conversations) != 2 {
		t.Fatalf("expected two conversations: %s", rec.Body.String())
	}
	first, second := body.Conversations[0], body.Conversations[1]
	if first.PeerUUID != bob || second.PeerUUID != carol {
		t.Errorf("newest conversation first: %s, %s", first.Username, second.Username)
	}
	if first.Username != "bob-c" {
		t.Errorf("username must be resolved: %q", first.Username)
	}

	var raw struct{ Conversations []map[string]any }
	json.Unmarshal(rec.Body.Bytes(), &raw)
	unread := map[string]float64{}
	for _, c := range raw.Conversations {
		unread[c["peer_uuid"].(string)] = c["unread"].(float64)
	}
	if unread[bob] != 2 || unread[carol] != 1 {
		t.Errorf("unread counts must only count incoming messages: %v", unread)
	}

	if rec := w.call(alice, "PATCH", "/read", map[string]any{"peer": "x"}); rec.Code != 400 {
		t.Errorf("bad peer: %d", rec.Code)
	}
	rec = w.call(alice, "PATCH", "/read", map[string]any{"peer": bob})
	var marked struct{ Marked int64 }
	json.Unmarshal(rec.Body.Bytes(), &marked)
	if marked.Marked != 2 {
		t.Errorf("marked = %d, want 2", marked.Marked)
	}
	if rec := w.call(alice, "PATCH", "/read", map[string]any{"peer": bob}); !jsonHas(rec.Body.String(), `"marked":0`) {
		t.Errorf("marking twice is a no-op: %s", rec.Body.String())
	}
	rec = w.call(alice, "GET", "/conversations", nil)
	json.Unmarshal(rec.Body.Bytes(), &raw)
	for _, c := range raw.Conversations {
		if c["peer_uuid"] == bob && c["unread"].(float64) != 0 {
			t.Error("read messages must not count as unread")
		}
	}

	if rec := w.call(alice, "DELETE", "/message/"+strconv.FormatInt(last, 10), nil); rec.Code != 404 {
		t.Errorf("the recipient cannot delete: %d", rec.Code)
	}
	if rec := w.call(carol, "DELETE", "/message/"+strconv.FormatInt(last, 10), nil); rec.Code != 204 {
		t.Errorf("the sender can delete: %d", rec.Code)
	}
	if rec := w.call(carol, "DELETE", "/message/"+strconv.FormatInt(last, 10), nil); rec.Code != 404 {
		t.Errorf("deleting twice: %d", rec.Code)
	}
	if rec := w.call(carol, "DELETE", "/message/x", nil); rec.Code != 400 {
		t.Errorf("bad id: %d", rec.Code)
	}
	if rec := w.call("", "DELETE", "/message/1", nil); rec.Code != 401 {
		t.Errorf("unauthenticated: %d", rec.Code)
	}
}

func jsonHas(body, part string) bool {
	return len(body) >= len(part) && (func() bool {
		for i := 0; i+len(part) <= len(body); i++ {
			if body[i:i+len(part)] == part {
				return true
			}
		}
		return false
	})()
}