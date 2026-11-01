package profile

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"log"
)

func UploadAvatar(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	const maxAvatarSize = 5 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if fh.Size > maxAvatarSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 5MB)"})
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

	ctx := c.Request.Context()
	var user models.User
	if err := db.DB.WithContext(ctx).First(&user, "uuid = ?", userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	name := uuid.NewString() + ext
	prefix := "users/" + userUUID.(string) + "/"
	_, err = storage.Client.PutObject(ctx, configs.S3Bucket(), prefix+name, f, fh.Size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("trace_id=%s s3 put error: %v", traceID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "storage error"})
		return
	}

	if err := db.DB.WithContext(ctx).Model(&models.User{}).Where("uuid = ?", userUUID).Update("avatar_key", name).Error; err != nil {
		log.Printf("trace_id=%s error saving avatar key: %v", traceID, err)
		storage.Client.RemoveObject(ctx, configs.S3Bucket(), prefix+name, minio.RemoveObjectOptions{})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if user.AvatarKey != nil {
		storage.Client.RemoveObject(ctx, configs.S3Bucket(), prefix+*user.AvatarKey, minio.RemoveObjectOptions{})
	}
	c.JSON(http.StatusOK, gin.H{"avatar_key": name})
}