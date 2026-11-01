package apitest

import (
	"Cadence-Cloud-Todo/core/api/boards"
	"Cadence-Cloud-Todo/markdown"
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http/httptest"
	"encoding/json"
	"strings"
	"testing"
	"reflect"
	"os"
)

func TestImportRejectsBadInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/i", func(c *gin.Context) { c.Set("userUUID", "u1"); boards.ImportBoard(c) })
	cases := map[string]struct{ contentType, body string }{
		"empty":           {"text/markdown", ""},
		"no tasks":        {"text/markdown", "# Title only"},
		"no title":        {"text/markdown", "## A\n- [ ] x"},
		"invalid json":    {"application/json", "{"},
		"json without md": {"application/json", `{"title":"t"}`},
	}
	for name, tc := range cases {
		req := httptest.NewRequest("POST", "/i", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", tc.contentType)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Errorf("%s: expected 400, got %d: %s", name, w.Code, w.Body.String())
		}
	}
	req := httptest.NewRequest("POST", "/i", strings.NewReader("# T\n## A\n"+strings.Repeat("- [ ] x\n", 5001)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 && w.Code != 500 {
		t.Errorf("5001 tasks must be rejected, got %d", w.Code)
	}
}

func TestImportAndExportSampleAgainstPostgres(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	owner := uuid.NewString()
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", owner); c.Next() })
	r.POST("/i", boards.ImportBoard)
	r.GET("/e/:id", boards.ExportBoard)
	r.GET("/b/:id", boards.GetBoard)

	sample, err := os.ReadFile("../markdown/testdata/sample.md")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/i", strings.NewReader(string(sample)))
	req.Header.Set("Content-Type", "text/markdown")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Board models.TodoBoard
		Nodes, Tasks int
	}
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.Board.Title != "PureC-OS todo" || created.Nodes != 7 || created.Tasks != 25 {
		t.Errorf("import result = %+v", created)
	}

	var positions int64
	tx.Model(&models.TodoNode{}).Where("board_id = ? AND (pos_x <> 0 OR pos_y <> 0)", created.Board.ID).Count(&positions)
	if positions != 6 {
		t.Errorf("sections must get an initial grid layout, %d have a position", positions)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/e/"+created.Board.ID, nil))
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("export: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	before, _ := markdown.Parse(string(sample))
	after, err := markdown.Parse(w.Body.String())
	if err != nil {
		t.Fatalf("exported markdown does not parse: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Error("export of an imported board differs from the source document")
	}

	req = httptest.NewRequest("POST", "/i", strings.NewReader(`{"title":"Own title","markdown":"## A\n- [ ] one"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"title":"Own title"`) {
		t.Errorf("json import: %d %s", w.Code, w.Body.String())
	}
}