package img

import (
	"Test-K8s-Pet-Project-Backend/core/api/images"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestListImagesUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/list", images.ListImages)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/list", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}