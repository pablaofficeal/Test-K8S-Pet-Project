package img

import (
	"Test-K8s-Pet-Project-Backend/core/api/images"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"mime/multipart"
	"testing"
	"bytes"
)

func TestUploadImageUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", images.UploadImage)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/upload", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUploadImageWithoutFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", func(c *gin.Context) { c.Set("userUUID", "u1"); images.UploadImage(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/upload", nil))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUploadImageRejectsNonImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/upload", func(c *gin.Context) { c.Set("userUUID", "u1"); images.UploadImage(c) })
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "note.txt")
	part.Write([]byte("just plain text, not an image"))
	mw.Close()
	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 415 {
		t.Fatalf("expected 415, got %d", w.Code)
	}
}