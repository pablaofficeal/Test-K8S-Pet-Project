package images

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"log"
)

func UploadImage(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	const maxImageSize = 10 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if fh.Size > maxImageSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 10MB)"})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read file"})
		return
	}
	defer f.Close()

	head := make([]byte, 512)
	n, _ := f.Read(head)
	contentType := http.DetectContentType(head[:n])
	allowedTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}
	ext, allowed := allowedTypes[contentType]
	if !allowed {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "only jpeg, png, webp, gif are allowed"})
		return
	}
	if _, err := f.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	name := uuid.NewString() + ext
	key := "users/" + userUUID.(string) + "/" + name
	_, err = storage.Client.PutObject(c.Request.Context(), configs.S3Bucket(), key, f, fh.Size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("trace_id=%s s3 put error: %v", traceID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "storage error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"key": name, "content_type": contentType, "size": fh.Size})
}