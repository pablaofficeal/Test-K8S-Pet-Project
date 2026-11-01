package images

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func ListImages(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	prefix := "users/" + userUUID.(string) + "/"
	items := []gin.H{}
	for obj := range storage.Client.ListObjects(c.Request.Context(), configs.S3Bucket(), minio.ListObjectsOptions{Prefix: prefix}) {
		if obj.Err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "storage error"})
			return
		}
		items = append(items, gin.H{"key": strings.TrimPrefix(obj.Key, prefix), "size": obj.Size, "last_modified": obj.LastModified})
	}
	c.JSON(http.StatusOK, gin.H{"images": items})
}