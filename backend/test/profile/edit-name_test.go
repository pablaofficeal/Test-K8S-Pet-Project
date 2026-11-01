package profiletest

import (
	"Test-K8s-Pet-Project-Backend/core/api/profile"
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEditMyNameUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/name", profile.EditMyName)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/name", strings.NewReader(`{"username":"x"}`)))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestEditMyNameValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/name", func(c *gin.Context) { c.Set("userUUID", "u1"); profile.EditMyName(c) })
	cases := map[string]string{
		"invalid json":   `{`,
		"empty":          `{"username":""}`,
		"only spaces":    `{"username":"   "}`,
		"missing field":  `{}`,
		"65 chars":       `{"username":"` + strings.Repeat("a", 65) + `"}`,
		"65 unicode":     `{"username":"` + strings.Repeat("ы", 65) + `"}`,
	}
	for name, body := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PATCH", "/name", strings.NewReader(body)))
		if w.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
}

func TestEditMyNameAgainstDB(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	a, b := uuid.NewString(), uuid.NewString()
	tx.Create(&models.User{UUID: a, Username: "a-" + a, Email: a + "@test.local", Password: "x"})
	tx.Create(&models.User{UUID: b, Username: "b-" + b, Email: b + "@test.local", Password: "x"})

	r := gin.New()
	var current string
	r.PATCH("/name", func(c *gin.Context) { c.Set("userUUID", current); profile.EditMyName(c) })

	current = a
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/name", strings.NewReader(`{"username":"  fresh-`+a+`  "}`)))
	if w.Code != 200 {
		t.Fatalf("rename: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var stored models.User
	tx.First(&stored, "uuid = ?", a)
	if stored.Username != "fresh-"+a {
		t.Errorf("stored username = %q, want trimmed value", stored.Username)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/name", strings.NewReader(`{"username":"fresh-`+a+`"}`)))
	if w.Code != 200 {
		t.Errorf("keeping own name: expected 200, got %d", w.Code)
	}

	current = b
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/name", strings.NewReader(`{"username":"fresh-`+a+`"}`)))
	if w.Code != 409 {
		t.Errorf("taken name: expected 409, got %d", w.Code)
	}

	current = uuid.NewString()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/name", strings.NewReader(`{"username":"nobody-`+a+`"}`)))
	if w.Code != 404 {
		t.Errorf("unknown user: expected 404, got %d", w.Code)
	}
}