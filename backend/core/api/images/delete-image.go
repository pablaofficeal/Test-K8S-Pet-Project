package images

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func DeleteImage(c *gin.Context) {
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
	if err := storage.Client.RemoveObject(c.Request.Context(), configs.S3Bucket(), key, minio.RemoveObjectOptions{}); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "storage error"})
		return
	}
	c.Status(http.StatusNoContent)
}