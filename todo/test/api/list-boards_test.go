package apitest

import (
	"Cadence-Cloud-Todo/core/api/boards"
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http/httptest"
	"encoding/json"
	"testing"
	"time"
)

func TestListBoardsReturnsStatisticsAndNewestFirst(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	owner, stranger := uuid.NewString(), uuid.NewString()
	older := models.TodoBoard{OwnerUUID: owner, Title: "older", UpdatedAt: time.Now().Add(-time.Hour)}
	newer := models.TodoBoard{OwnerUUID: owner, Title: "newer"}
	foreign := models.TodoBoard{OwnerUUID: stranger, Title: "foreign"}
	tx.Create(&older)
	tx.Create(&newer)
	tx.Create(&foreign)
	tx.Model(&older).UpdateColumn("updated_at", time.Now().Add(-time.Hour))

	nodeA := models.TodoNode{BoardID: older.ID, Title: "a"}
	nodeB := models.TodoNode{BoardID: older.ID, Title: "b"}
	foreignNode := models.TodoNode{BoardID: foreign.ID, Title: "x"}
	tx.Create(&nodeA)
	tx.Create(&nodeB)
	tx.Create(&foreignNode)
	for _, status := range []string{"done", "done", "todo", "in_progress"} {
		tx.Create(&models.TodoTask{BoardID: older.ID, NodeID: nodeA.ID, Title: "t", Status: status})
	}
	tx.Create(&models.TodoTask{BoardID: foreign.ID, NodeID: foreignNode.ID, Title: "x", Status: "done"})

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", owner); c.Next() })
	r.GET("/b", boards.ListBoards)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/b", nil))
	if w.Code != 200 {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		Boards []struct {
			ID, Title string
			Nodes int64 `json:"nodes_count"`
			Total int64 `json:"tasks_total"`
			Done int64 `json:"tasks_done"`
		}
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Boards) != 2 {
		t.Fatalf("expected the two own boards, got %s", w.Body.String())
	}
	if body.Boards[0].Title != "newer" || body.Boards[1].Title != "older" {
		t.Errorf("boards must be newest first: %s, %s", body.Boards[0].Title, body.Boards[1].Title)
	}
	empty, filled := body.Boards[0], body.Boards[1]
	if empty.Nodes != 0 || empty.Total != 0 || empty.Done != 0 {
		t.Errorf("an empty board must report zeros: %+v", empty)
	}
	if filled.Nodes != 2 || filled.Total != 4 || filled.Done != 2 {
		t.Errorf("statistics of the filled board: nodes=%d total=%d done=%d", filled.Nodes, filled.Total, filled.Done)
	}
}