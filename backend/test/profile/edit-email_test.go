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
	"time"
)

func TestEditMyEmailUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/email", profile.EditMyEmail)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/email", strings.NewReader(`{"email":"a@b.com"}`)))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestEditMyEmailValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/email", func(c *gin.Context) { c.Set("userUUID", "u1"); profile.EditMyEmail(c) })
	cases := map[string]string{
		"invalid json":   `{`,
		"empty":          `{"email":""}`,
		"missing field":  `{}`,
		"no at sign":     `{"email":"plain"}`,
		"display name":   `{"email":"Name <a@b.com>"}`,
		"angle brackets": `{"email":"<a@b.com>"}`,
		"two addresses":  `{"email":"a@b.com, c@d.com"}`,
		"too long":       `{"email":"` + strings.Repeat("a", 250) + `@b.com"}`,
	}
	for name, body := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PATCH", "/email", strings.NewReader(body)))
		if w.Code != 400 {
			t.Errorf("%s: expected 400, got %d", name, w.Code)
		}
	}
}

func TestEditMyEmailAgainstDB(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	a, b := uuid.NewString(), uuid.NewString()
	verifiedAt := time.Now()
	tx.Create(&models.User{UUID: a, Username: "a-" + a, Email: a + "@test.local", Password: "x", EmailVerified: true, EmailVerifiedAt: &verifiedAt})
	tx.Create(&models.User{UUID: b, Username: "b-" + b, Email: b + "@test.local", Password: "x"})

	r := gin.New()
	var current string
	r.PATCH("/email", func(c *gin.Context) { c.Set("userUUID", current); profile.EditMyEmail(c) })

	current = a
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/email", strings.NewReader(`{"email":"new-`+a+`@test.local"}`)))
	if w.Code != 200 {
		t.Fatalf("change: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var stored models.User
	tx.First(&stored, "uuid = ?", a)
	if stored.Email != "new-"+a+"@test.local" || stored.EmailVerified || stored.EmailVerifiedAt != nil {
		t.Errorf("email change must reset verification, got %+v", stored)
	}

	current = b
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/email", strings.NewReader(`{"email":"NEW-`+strings.ToUpper(a)+`@TEST.LOCAL"}`)))
	if w.Code != 409 {
		t.Errorf("case-insensitive duplicate: expected 409, got %d", w.Code)
	}

	current = uuid.NewString()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/email", strings.NewReader(`{"email":"ghost-`+a+`@test.local"}`)))
	if w.Code != 404 {
		t.Errorf("unknown user: expected 404, got %d", w.Code)
	}
}