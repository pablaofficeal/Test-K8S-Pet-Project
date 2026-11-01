package profiletest

import (
	"Test-K8s-Pet-Project-Backend/core/api/profile"
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"encoding/base64"
	"mime/multipart"
	"encoding/json"
	"context"
	"testing"
	"bytes"
)

func TestUploadAvatarUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/avatar", profile.UploadAvatar)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/avatar", nil))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUploadAvatarValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/avatar", func(c *gin.Context) { c.Set("userUUID", "u1"); profile.UploadAvatar(c) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/avatar", nil))
	if w.Code != 400 {
		t.Errorf("no file: expected 400, got %d", w.Code)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "note.txt")
	part.Write([]byte("plain text, not an image"))
	mw.Close()
	req := httptest.NewRequest("POST", "/avatar", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 415 {
		t.Errorf("non image: expected 415, got %d", w.Code)
	}
}

func TestAvatarFlowAgainstDBAndS3(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	if err := storage.InitS3(context.Background()); err != nil {
		t.Skipf("s3 is not available: %v", err)
	}
	gin.SetMode(gin.TestMode)
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	id := uuid.NewString()
	tx.Create(&models.User{UUID: id, Username: "av-" + id, Email: id + "@test.local", Password: "x"})

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userUUID", id); c.Next() })
	r.POST("/avatar", profile.UploadAvatar)
	r.DELETE("/avatar", profile.DeleteAvatar)
	r.GET("/info", profile.GetProfileInfo)

	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")
	keys := []string{}
	for i := 0; i < 2; i++ {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		part, _ := mw.CreateFormFile("file", "pixel.png")
		part.Write(png)
		mw.Close()
		req := httptest.NewRequest("POST", "/avatar", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("upload %d: expected 200, got %d: %s", i, w.Code, w.Body.String())
		}
		var res struct{ AvatarKey string `json:"avatar_key"` }
		json.Unmarshal(w.Body.Bytes(), &res)
		keys = append(keys, res.AvatarKey)
	}
	if keys[0] == "" || keys[0] == keys[1] {
		t.Fatalf("each upload must produce a new key: %v", keys)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/info", nil))
	var info map[string]any
	json.Unmarshal(w.Body.Bytes(), &info)
	if info["avatar_key"] != keys[1] {
		t.Errorf("profile avatar_key = %v, want %s", info["avatar_key"], keys[1])
	}

	prefix := "users/" + id + "/"
	if _, err := storage.Client.StatObject(context.Background(), configs.S3Bucket(), prefix+keys[0], minio.StatObjectOptions{}); err == nil {
		t.Error("previous avatar object must be removed after replacing")
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/avatar", nil))
	if w.Code != 204 {
		t.Fatalf("delete: expected 204, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/info", nil))
	info = map[string]any{}
	json.Unmarshal(w.Body.Bytes(), &info)
	if info["avatar_key"] != nil {
		t.Errorf("avatar_key after delete = %v, want null", info["avatar_key"])
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/avatar", nil))
	if w.Code != 204 {
		t.Errorf("deleting without avatar: expected 204, got %d", w.Code)
	}
}