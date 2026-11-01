package profiletest

import (
	"Test-K8s-Pet-Project-Backend/core/api/profile"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestDeleteAvatarUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/avatar", profile.DeleteAvatar)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/avatar", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}