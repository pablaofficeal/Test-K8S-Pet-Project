package img

import (
	"Test-K8s-Pet-Project-Backend/core/api/images"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestGetImageUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/image/:key", images.GetImage)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/image/a.png", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGetImageRejectsTraversalKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/image/:key", func(c *gin.Context) { c.Set("userUUID", "u1"); images.GetImage(c) })
	for _, key := range []string{"..", "a..b", "a%5Cb"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/image/"+key, nil))
		if w.Code != 400 {
			t.Fatalf("key %q: expected 400, got %d", key, w.Code)
		}
	}
}