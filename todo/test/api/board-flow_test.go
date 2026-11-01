package apitest

import (
	"Cadence-Cloud-Todo/core/api/boards"
	"Cadence-Cloud-Todo/core/api/nodes"
	"Cadence-Cloud-Todo/core/api/tasks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http/httptest"
	"encoding/json"
	"strings"
	"testing"
)

func TestBoardHandlersUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/b", boards.CreateBoard)
	r.GET("/b", boards.ListBoards)
	r.GET("/b/:id", boards.GetBoard)
	r.DELETE("/b/:id", boards.DeleteBoard)
	r.POST("/i", boards.ImportBoard)
	r.GET("/e/:id", boards.ExportBoard)
	r.POST("/b/:id/n", nodes.CreateNode)
	r.PATCH("/n/:id", nodes.EditNode)
	r.PUT("/o/:id", nodes.ReorderNodes)
	r.PATCH("/mt/:id", tasks.MoveTask)
	r.DELETE("/n/:id", nodes.DeleteNode)
	r.POST("/n/:id/t", tasks.CreateTask)
	r.PATCH("/t/:id", tasks.EditTask)
	r.DELETE("/t/:id", tasks.DeleteTask)
	id := uuid.NewString()
	for _, tc := range [][2]string{
		{"POST", "/b"}, {"GET", "/b"}, {"GET", "/b/" + id}, {"DELETE", "/b/" + id}, {"POST", "/i"}, {"GET", "/e/" + id},
		{"POST", "/b/" + id + "/n"}, {"PATCH", "/n/" + id}, {"PUT", "/o/" + id}, {"PATCH", "/mt/" + id}, {"DELETE", "/n/" + id},
		{"POST", "/n/" + id + "/t"}, {"PATCH", "/t/" + id}, {"DELETE", "/t/" + id},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(tc[0], tc[1], strings.NewReader("{}")))
		if w.Code != 401 {
			t.Errorf("%s %s: expected 401, got %d", tc[0], tc[1], w.Code)
		}
	}
}

func TestCreateBoardValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/b", func(c *gin.Context) { c.Set("userUUID", "u1"); boards.CreateBoard(c) })
	for name, body := range map[string]string{
		"invalid json": `{`,
		"empty":        `{"title":""}`,
		"spaces":       `{"title":"   "}`,
		"too long":     `{"title":"` + strings.Repeat("a", 201) + `"}`,
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", "/b", strings.NewReader(body)))
		if w.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
}

func TestInvalidIDsAreRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", "u1"); c.Next() })
	r.GET("/b/:id", boards.GetBoard)
	r.DELETE("/b/:id", boards.DeleteBoard)
	r.GET("/e/:id", boards.ExportBoard)
	r.PUT("/o/:id", nodes.ReorderNodes)
	r.PATCH("/mt/:id", tasks.MoveTask)
	r.PATCH("/t/:id", tasks.EditTask)
	for _, tc := range [][2]string{{"GET", "/b/x"}, {"DELETE", "/b/x"}, {"GET", "/e/x"}, {"PUT", "/o/x"}, {"PATCH", "/mt/x"}, {"PATCH", "/t/x"}} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(tc[0], tc[1], strings.NewReader("{}")))
		if w.Code != 400 {
			t.Errorf("%s %s: expected 400, got %d", tc[0], tc[1], w.Code)
		}
	}
}

func TestBoardLifecycleAgainstPostgres(t *testing.T) {
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
	r.GET("/b", boards.ListBoards)
	r.GET("/b/:id", boards.GetBoard)
	r.DELETE("/b/:id", boards.DeleteBoard)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/b", strings.NewReader(`{"title":"  Roadmap  "}`)))
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var board struct{ ID, Title string }
	json.Unmarshal(w.Body.Bytes(), &board)
	if board.Title != "Roadmap" {
		t.Errorf("title was not trimmed: %q", board.Title)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/b", nil))
	var list struct{ Boards []struct{ ID string } }
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Boards) != 1 || list.Boards[0].ID != board.ID {
		t.Errorf("list = %s", w.Body.String())
	}

	current = bob
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/b", nil))
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Boards) != 0 {
		t.Errorf("another user sees foreign boards: %s", w.Body.String())
	}
	for _, method := range []string{"GET", "DELETE"} {
		w = httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, "/b/"+board.ID, nil))
		if w.Code != 404 {
			t.Errorf("%s of a foreign board: expected 404, got %d", method, w.Code)
		}
	}

	current = alice
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/b/"+board.ID, nil))
	if w.Code != 200 {
		t.Errorf("get own board: %d", w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/b/"+board.ID, nil))
	if w.Code != 204 {
		t.Errorf("delete own board: %d", w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/b/"+board.ID, nil))
	if w.Code != 404 {
		t.Errorf("get after delete: %d", w.Code)
	}
}