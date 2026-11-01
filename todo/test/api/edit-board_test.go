package apitest

import (
	"Cadence-Cloud-Todo/core/api/boards"
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEditBoardValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/b/:id", func(c *gin.Context) { c.Set("userUUID", "u1"); boards.EditBoard(c) })
	for name, tc := range map[string][2]string{
		"bad id":    {"/b/x", `{"title":"t"}`},
		"bad json":  {"/b/" + uuid.NewString(), `{`},
		"empty":     {"/b/" + uuid.NewString(), `{"title":" "}`},
		"too long":  {"/b/" + uuid.NewString(), `{"title":"` + strings.Repeat("a", 201) + `"}`},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PATCH", tc[0], strings.NewReader(tc[1])))
		if w.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
}

func TestEditBoardAgainstPostgres(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	owner := uuid.NewString()
	board := models.TodoBoard{OwnerUUID: owner, Title: "old"}
	tx.Create(&board)
	current := owner
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", current); c.Next() })
	r.PATCH("/b/:id", boards.EditBoard)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/b/"+board.ID, strings.NewReader(`{"title":"  new  "}`)))
	var stored models.TodoBoard
	tx.First(&stored, "id = ?", board.ID)
	if w.Code != 200 || stored.Title != "new" {
		t.Errorf("rename: %d, stored %q", w.Code, stored.Title)
	}

	current = uuid.NewString()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/b/"+board.ID, strings.NewReader(`{"title":"hijack"}`)))
	tx.First(&stored, "id = ?", board.ID)
	if w.Code != 404 || stored.Title != "new" {
		t.Errorf("foreign rename: %d, stored %q", w.Code, stored.Title)
	}
}