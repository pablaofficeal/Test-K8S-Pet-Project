package img

import (
	"Test-K8s-Pet-Project-Backend/core/api/images"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestDeleteImageUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/image/:key", images.DeleteImage)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/image/a.png", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDeleteImageRejectsTraversalKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/image/:key", func(c *gin.Context) { c.Set("userUUID", "u1"); images.DeleteImage(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/image/..", nil))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}