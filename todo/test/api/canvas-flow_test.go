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

func TestCanvasWithNodesAndTasksAgainstPostgres(t *testing.T) {
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
	r.PATCH("/n/:id", nodes.EditNode)
	r.PATCH("/m/:id", nodes.MoveNode)
	r.DELETE("/n/:id", nodes.DeleteNode)
	r.POST("/n/:id/t", tasks.CreateTask)
	r.PATCH("/t/:id", tasks.EditTask)
	r.DELETE("/t/:id", tasks.DeleteTask)

	send := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}

	var board struct{ ID string }
	json.Unmarshal(send("POST", "/b", `{"title":"Canvas"}`).Body.Bytes(), &board)

	if w := send("POST", "/b/"+board.ID+"/n", `{"title":""}`); w.Code != 400 {
		t.Errorf("empty node title: %d", w.Code)
	}
	if w := send("POST", "/b/"+board.ID+"/n", `{"title":"x","pos_x":2000000}`); w.Code != 400 {
		t.Errorf("out of range position: %d", w.Code)
	}
	w := send("POST", "/b/"+board.ID+"/n", `{"title":"UserSpace","pos_x":10,"pos_y":20}`)
	if w.Code != 201 {
		t.Fatalf("create node: %d %s", w.Code, w.Body.String())
	}
	var node struct{ ID string }
	json.Unmarshal(w.Body.Bytes(), &node)

	current = bob
	if w := send("PATCH", "/n/"+node.ID, `{"title":"hijack"}`); w.Code != 404 {
		t.Errorf("edit of a foreign node: %d", w.Code)
	}
	if w := send("PATCH", "/m/"+node.ID, `{"x":1,"y":1}`); w.Code != 404 {
		t.Errorf("move of a foreign node: %d", w.Code)
	}
	if w := send("POST", "/n/"+node.ID+"/t", `{"title":"intruder"}`); w.Code != 404 {
		t.Errorf("task in a foreign node: %d", w.Code)
	}
	if hot, _ := cache.GetPositions(context.Background(), board.ID); len(hot) != 0 {
		t.Errorf("a foreign move reached redis: %v", hot)
	}
	current = alice

	if w := send("PATCH", "/n/"+node.ID, `{"title":"Kernel"}`); w.Code != 200 {
		t.Errorf("edit node: %d", w.Code)
	}

	for name, body := range map[string]string{
		"no coordinates": `{}`,
		"only x":         `{"x":1}`,
		"invalid json":   `{`,
		"out of range":   `{"x":1e9,"y":0}`,
	} {
		if w := send("PATCH", "/m/"+node.ID, body); w.Code != 400 {
			t.Errorf("move %s: expected 400, got %d", name, w.Code)
		}
	}
	for _, p := range [][2]float64{{100, 200}, {110, 210}, {120, 220}} {
		b, _ := json.Marshal(map[string]float64{"x": p[0], "y": p[1]})
		if w := send("PATCH", "/m/"+node.ID, string(b)); w.Code != 200 {
			t.Fatalf("move: %d %s", w.Code, w.Body.String())
		}
	}
	var stored models.TodoNode
	tx.First(&stored, "id = ?", node.ID)
	if stored.PosX != 10 || stored.PosY != 20 {
		t.Errorf("dragging wrote to postgres before the flush: %v,%v", stored.PosX, stored.PosY)
	}
	w = send("GET", "/b/"+board.ID, "")
	var view struct{ Nodes []struct {
		ID string
		PosX float64 `json:"pos_x"`
		PosY float64 `json:"pos_y"`
	} }
	json.Unmarshal(w.Body.Bytes(), &view)
	if len(view.Nodes) != 1 || view.Nodes[0].PosX != 120 {
		t.Errorf("get board must overlay the hot position, got %s", w.Body.String())
	}
	cache.FlushPositions(context.Background())
	tx.First(&stored, "id = ?", node.ID)
	if stored.PosX != 120 || stored.PosY != 220 {
		t.Errorf("after the flush postgres has %v,%v, want 120,220", stored.PosX, stored.PosY)
	}

	for name, body := range map[string]string{
		"empty title":  `{"title":""}`,
		"bad status":   `{"title":"a","status":"finished"}`,
		"bad priority": `{"title":"a","priority":"urgent"}`,
		"bad deadline": `{"title":"a","deadline":"23.09.2026"}`,
		"bad tag":      `{"title":"a","tags":["a,b"]}`,
		"bad parent":   `{"title":"a","parent_id":"nope"}`,
		"foreign parent": `{"title":"a","parent_id":"` + uuid.NewString() + `"}`,
	} {
		wantCode := 400
		if name == "foreign parent" {
			wantCode = 404
		}
		if w := send("POST", "/n/"+node.ID+"/t", body); w.Code != wantCode {
			t.Errorf("create task %s: expected %d, got %d", name, wantCode, w.Code)
		}
	}

	w = send("POST", "/n/"+node.ID+"/t", `{"title":"User space ведёт себя не так","status":"in_progress","priority":"high","tags":["bug"," kernel "],"deadline":"2026-09-23","description":"line1\nline2"}`)
	if w.Code != 201 {
		t.Fatalf("create task: %d %s", w.Code, w.Body.String())
	}
	var parent struct{ ID, Status, Priority, Deadline string; Tags []string }
	json.Unmarshal(w.Body.Bytes(), &parent)
	if parent.Status != "in_progress" || parent.Priority != "high" || len(parent.Tags) != 2 || parent.Tags[1] != "kernel" || !strings.HasPrefix(parent.Deadline, "2026-09-23") {
		t.Errorf("task stored wrong: %s", w.Body.String())
	}

	chain := parent.ID
	for depth := 2; depth <= 9; depth++ {
		w = send("POST", "/n/"+node.ID+"/t", `{"title":"level","parent_id":"`+chain+`"}`)
		if depth <= 8 && w.Code != 201 {
			t.Fatalf("level %d: %d %s", depth, w.Code, w.Body.String())
		}
		if depth == 9 && w.Code != 400 {
			t.Errorf("level 9 must be rejected, got %d", w.Code)
		}
		var child struct{ ID string }
		json.Unmarshal(w.Body.Bytes(), &child)
		chain = child.ID
	}

	if w := send("PATCH", "/t/"+parent.ID, `{}`); w.Code != 400 {
		t.Errorf("empty edit: %d", w.Code)
	}
	if w := send("PATCH", "/t/"+parent.ID, `{"status":"done","priority":"low","tags":[],"deadline":"","title":"Fixed"}`); w.Code != 200 {
		t.Fatalf("edit task: %d %s", w.Code, w.Body.String())
	}
	var edited models.TodoTask
	tx.First(&edited, "id = ?", parent.ID)
	if edited.Status != "done" || edited.Priority != "low" || edited.Title != "Fixed" || edited.Deadline != nil || string(edited.Tags) != "[]" {
		t.Errorf("edit not applied: %+v", edited)
	}

	if w := send("DELETE", "/t/"+parent.ID, ""); w.Code != 204 {
		t.Fatalf("delete task: %d", w.Code)
	}
	var left int64
	tx.Model(&models.TodoTask{}).Where("node_id = ?", node.ID).Count(&left)
	if left != 0 {
		t.Errorf("deleting a task must remove its %d descendants", left)
	}

	send("POST", "/n/"+node.ID+"/t", `{"title":"to be removed with the node"}`)
	if w := send("DELETE", "/n/"+node.ID, ""); w.Code != 204 {
		t.Fatalf("delete node: %d", w.Code)
	}
	tx.Model(&models.TodoTask{}).Where("node_id = ?", node.ID).Count(&left)
	if left != 0 {
		t.Error("deleting a node must remove its tasks")
	}
	if hot, _ := cache.GetPositions(context.Background(), board.ID); len(hot) != 0 {
		t.Errorf("hot position of a deleted node must be dropped: %v", hot)
	}
}