package apitest

import (
	"Cadence-Cloud-Todo/core/api/boards"
	"Cadence-Cloud-Todo/core/api/nodes"
	"Cadence-Cloud-Todo/core/api/tasks"
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http/httptest"
	"encoding/json"
	"strings"
	"testing"
	"context"
)

func TestKanbanListsOrderLivesInRedisUntilFlushed(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	redisServer.FlushAll()
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	alice, bob := uuid.NewString(), uuid.NewString()
	current := alice
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", current); c.Next() })
	r.POST("/b", boards.CreateBoard)
	r.GET("/b/:id", boards.GetBoard)
	r.POST("/b/:id/n", nodes.CreateNode)
	r.PUT("/o/:id", nodes.ReorderNodes)
	r.PATCH("/n/:id", nodes.EditNode)
	r.DELETE("/n/:id", nodes.DeleteNode)

	send := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	titles := func(boardID string) []string {
		var view struct{ Nodes []struct{ Title string } }
		json.Unmarshal(send("GET", "/b/"+boardID, "").Body.Bytes(), &view)
		out := []string{}
		for _, n := range view.Nodes {
			out = append(out, n.Title)
		}
		return out
	}

	var board struct{ ID string }
	json.Unmarshal(send("POST", "/b", `{"title":"Kanban"}`).Body.Bytes(), &board)
	ids := map[string]string{}
	for _, name := range []string{"A", "B", "C"} {
		w := send("POST", "/b/"+board.ID+"/n", `{"title":"`+name+`"}`)
		if w.Code != 201 {
			t.Fatalf("create node %s: %d %s", name, w.Code, w.Body.String())
		}
		var node struct{ ID string }
		json.Unmarshal(w.Body.Bytes(), &node)
		ids[name] = node.ID
	}
	if got := strings.Join(titles(board.ID), ""); got != "ABC" {
		t.Fatalf("initial order = %s", got)
	}

	if w := send("POST", "/b/"+board.ID+"/n", `{"title":""}`); w.Code != 400 {
		t.Errorf("empty node title: %d", w.Code)
	}
	current = bob
	if w := send("PATCH", "/n/"+ids["A"], `{"title":"hijack"}`); w.Code != 404 {
		t.Errorf("edit of a foreign node: %d", w.Code)
	}
	if w := send("PUT", "/o/"+board.ID, `{"order":["`+ids["C"]+`","`+ids["A"]+`","`+ids["B"]+`"]}`); w.Code != 404 {
		t.Errorf("reorder of a foreign board: %d", w.Code)
	}
	current = alice

	for name, body := range map[string]string{
		"invalid json": `{`,
		"no order":     `{}`,
		"null order":   `{"order":null}`,
		"empty order":  `{"order":[]}`,
		"missing node": `{"order":["` + ids["A"] + `","` + ids["B"] + `"]}`,
		"duplicate":    `{"order":["` + ids["A"] + `","` + ids["A"] + `","` + ids["B"] + `"]}`,
		"unknown node": `{"order":["` + ids["A"] + `","` + ids["B"] + `","` + uuid.NewString() + `"]}`,
		"extra node":   `{"order":["` + ids["A"] + `","` + ids["B"] + `","` + ids["C"] + `","` + uuid.NewString() + `"]}`,
	} {
		if w := send("PUT", "/o/"+board.ID, body); w.Code != 400 {
			t.Errorf("reorder %s: expected 400, got %d", name, w.Code)
		}
	}
	if got := strings.Join(titles(board.ID), ""); got != "ABC" {
		t.Errorf("a rejected reorder changed the order: %s", got)
	}

	if w := send("PUT", "/o/"+board.ID, `{"order":["`+ids["C"]+`","`+ids["A"]+`","`+ids["B"]+`"]}`); w.Code != 200 {
		t.Fatalf("reorder: %d %s", w.Code, w.Body.String())
	}
	var stored models.TodoNode
	tx.First(&stored, "id = ?", ids["C"])
	if stored.SortOrder != 2 {
		t.Errorf("reordering wrote to postgres before the flush: %d", stored.SortOrder)
	}
	if got := strings.Join(titles(board.ID), ""); got != "CAB" {
		t.Errorf("get board must show the hot order, got %s", got)
	}

	cache.FlushOrders(context.Background())
	for name, want := range map[string]int{"C": 0, "A": 1, "B": 2} {
		var flushed models.TodoNode
		tx.First(&flushed, "id = ?", ids[name])
		if flushed.SortOrder != want {
			t.Errorf("after the flush %s has sort_order %d, want %d", name, flushed.SortOrder, want)
		}
	}

	w := send("POST", "/b/"+board.ID+"/n", `{"title":"D"}`)
	var added struct{ ID string }
	json.Unmarshal(w.Body.Bytes(), &added)
	if got := strings.Join(titles(board.ID), ""); got != "CABD" {
		t.Errorf("a new list must go to the end of the hot order, got %s", got)
	}
	if w := send("PATCH", "/n/"+ids["A"], `{"title":"A2"}`); w.Code != 200 {
		t.Errorf("edit node: %d", w.Code)
	}
	if w := send("DELETE", "/n/"+ids["A"], ""); w.Code != 204 {
		t.Fatalf("delete node: %d", w.Code)
	}
	if got := strings.Join(titles(board.ID), ""); got != "CBD" {
		t.Errorf("a deleted list must leave the hot order, got %s", got)
	}
	hot, _, _ := cache.GetOrder(context.Background(), board.ID)
	for _, id := range hot {
		if id == ids["A"] {
			t.Error("deleted node is still in redis")
		}
	}
}

func TestKanbanMoveTaskBetweenLists(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	redisServer.FlushAll()
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	alice, bob := uuid.NewString(), uuid.NewString()
	current := alice
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", current); c.Next() })
	r.POST("/b", boards.CreateBoard)
	r.POST("/b/:id/n", nodes.CreateNode)
	r.POST("/n/:id/t", tasks.CreateTask)
	r.PATCH("/mt/:id", tasks.MoveTask)

	send := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	idOf := func(w *httptest.ResponseRecorder) string {
		var v struct{ ID string }
		json.Unmarshal(w.Body.Bytes(), &v)
		return v.ID
	}
	order := func(nodeID string) string {
		rows := []models.TodoTask{}
		tx.Where("node_id = ? AND parent_id IS NULL", nodeID).Order("sort_order").Find(&rows)
		out := ""
		for _, row := range rows {
			out += row.Title
		}
		return out
	}

	boardID := idOf(send("POST", "/b", `{"title":"Kanban"}`))
	nodeA := idOf(send("POST", "/b/"+boardID+"/n", `{"title":"A"}`))
	nodeB := idOf(send("POST", "/b/"+boardID+"/n", `{"title":"B"}`))
	t1 := idOf(send("POST", "/n/"+nodeA+"/t", `{"title":"1"}`))
	t2 := idOf(send("POST", "/n/"+nodeA+"/t", `{"title":"2"}`))
	send("POST", "/n/"+nodeB+"/t", `{"title":"3"}`)
	child := idOf(send("POST", "/n/"+nodeA+"/t", `{"title":"c","parent_id":"`+t1+`"}`))
	if order(nodeA) != "12" || order(nodeB) != "3" {
		t.Fatalf("initial order A=%s B=%s", order(nodeA), order(nodeB))
	}

	for name, tc := range map[string][2]string{
		"bad id":         {"x", `{"node_id":"` + nodeB + `","index":0}`},
		"bad json":       {t1, `{`},
		"no index":       {t1, `{"node_id":"` + nodeB + `"}`},
		"negative index": {t1, `{"node_id":"` + nodeB + `","index":-1}`},
		"bad node id":    {t1, `{"node_id":"nope","index":0}`},
		"subtask":        {child, `{"node_id":"` + nodeB + `","index":0}`},
	} {
		if w := send("PATCH", "/mt/"+tc[0], tc[1]); w.Code != 400 {
			t.Errorf("move %s: expected 400, got %d", name, w.Code)
		}
	}
	if w := send("PATCH", "/mt/"+t1, `{"node_id":"`+uuid.NewString()+`","index":0}`); w.Code != 404 {
		t.Errorf("unknown target list: %d", w.Code)
	}
	otherBoard := idOf(send("POST", "/b", `{"title":"Other"}`))
	foreignList := idOf(send("POST", "/b/"+otherBoard+"/n", `{"title":"X"}`))
	if w := send("PATCH", "/mt/"+t1, `{"node_id":"`+foreignList+`","index":0}`); w.Code != 404 {
		t.Errorf("a list of another board: %d", w.Code)
	}
	current = bob
	if w := send("PATCH", "/mt/"+t1, `{"node_id":"`+nodeB+`","index":0}`); w.Code != 404 {
		t.Errorf("moving a foreign task: %d", w.Code)
	}
	current = alice
	if order(nodeA) != "12" {
		t.Errorf("rejected moves changed list A: %s", order(nodeA))
	}

	if w := send("PATCH", "/mt/"+t1, `{"node_id":"`+nodeB+`","index":0}`); w.Code != 200 {
		t.Fatalf("move to another list: %d %s", w.Code, w.Body.String())
	}
	if order(nodeA) != "2" || order(nodeB) != "13" {
		t.Errorf("after the move A=%s B=%s", order(nodeA), order(nodeB))
	}
	var moved models.TodoTask
	tx.First(&moved, "id = ?", child)
	if moved.NodeID != nodeB {
		t.Error("subtasks must move together with their card")
	}

	send("PATCH", "/mt/"+t1, `{"node_id":"`+nodeB+`","index":99}`)
	if order(nodeB) != "31" {
		t.Errorf("an index past the end must clamp to the end: %s", order(nodeB))
	}
	send("PATCH", "/mt/"+t2, `{"node_id":"`+nodeB+`","index":1}`)
	if order(nodeB) != "321" || order(nodeA) != "" {
		t.Errorf("insert in the middle: B=%s A=%s", order(nodeB), order(nodeA))
	}
	send("PATCH", "/mt/"+t2, `{"node_id":"`+nodeB+`","index":0}`)
	if order(nodeB) != "231" {
		t.Errorf("reorder inside one list: %s", order(nodeB))
	}
}