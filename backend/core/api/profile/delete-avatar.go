package profile

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/minio/minio-go/v7"
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

func DeleteAvatar(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}

	ctx := c.Request.Context()
	var user models.User
	if err := db.DB.WithContext(ctx).First(&user, "uuid = ?", userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user.AvatarKey == nil {
		c.Status(http.StatusNoContent)
		return
	}

	if err := db.DB.WithContext(ctx).Model(&models.User{}).Where("uuid = ?", userUUID).Update("avatar_key", nil).Error; err != nil {
		log.Printf("trace_id=%s error clearing avatar key: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	storage.Client.RemoveObject(ctx, configs.S3Bucket(), "users/"+userUUID.(string)+"/"+*user.AvatarKey, minio.RemoveObjectOptions{})
	c.Status(http.StatusNoContent)
}