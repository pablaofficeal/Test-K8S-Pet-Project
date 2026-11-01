package img

import (
	"Test-K8s-Pet-Project-Backend/core/api/images"
	"Test-K8s-Pet-Project-Backend/storage"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"encoding/base64"
	"mime/multipart"
	"encoding/json"
	"context"
	"testing"
	"bytes"
)

func TestImageFlowAgainstS3(t *testing.T) {
	godotenv.Load("../../.env")
	if err := storage.InitS3(context.Background()); err != nil {
		t.Skipf("s3 is not available: %v", err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", "test-flow-user"); c.Next() })
	r.POST("/upload", images.UploadImage)
	r.GET("/list", images.ListImages)
	r.GET("/image/:key", images.GetImage)
	r.DELETE("/image/:key", images.DeleteImage)

	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "pixel.png")
	part.Write(png)
	mw.Close()
	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("upload: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var uploaded struct{ Key string `json:"key"` }
	json.Unmarshal(w.Body.Bytes(), &uploaded)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/list", nil))
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(uploaded.Key)) {
		t.Fatalf("list: got %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/image/"+uploaded.Key, nil))
	if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), png) {
		t.Fatalf("get: got %d, body mismatch", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/image/"+uploaded.Key, nil))
	if w.Code != 204 {
		t.Fatalf("delete: expected 204, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/image/"+uploaded.Key, nil))
	if w.Code != 404 {
		t.Fatalf("get after delete: expected 404, got %d", w.Code)
	}
}