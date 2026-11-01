package images

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetImage(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	name := c.Param("key")
	if name == "" || strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}
	key := "users/" + userUUID.(string) + "/" + name
	obj, err := storage.Client.GetObject(c.Request.Context(), configs.S3Bucket(), key, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "storage error"})
		return
	}
	defer obj.Close()
	st, err := obj.Stat()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}
	c.DataFromReader(http.StatusOK, st.Size, st.ContentType, obj, map[string]string{
		"Cache-Control": "private, max-age=31536000, immutable",
		"ETag":          "\"" + st.ETag + "\"",
	})
}