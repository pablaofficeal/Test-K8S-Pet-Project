package profiletest

import (
	"Test-K8s-Pet-Project-Backend/core/api/profile"
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"encoding/json"
	"testing"
	"strings"
)

func TestGetProfileInfoUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/info", profile.GetProfileInfo)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/info", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetProfileInfoAgainstDB(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	id := uuid.NewString()
	tx.Create(&models.User{UUID: id, Username: "n-" + id, Email: id + "@test.local", Password: "secret-hash"})

	r := gin.New()
	var current string
	r.GET("/info", func(c *gin.Context) { c.Set("userUUID", current); profile.GetProfileInfo(c) })

	current = id
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/info", nil))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["uuid"] != id || body["email"] != id+"@test.local" || body["email_verified"] != false {
		t.Errorf("unexpected body: %v", body)
	}
	if strings.Contains(w.Body.String(), "secret-hash") {
		t.Error("password hash leaked in response")
	}

	current = uuid.NewString()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/info", nil))
	if w.Code != 404 {
		t.Fatalf("unknown user: expected 404, got %d", w.Code)
	}
}