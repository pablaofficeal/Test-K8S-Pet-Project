package boards

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"Cadence-Cloud-Todo/db"
	"net/http"
	"log"
)

func ListBoards(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	items := []models.TodoBoard{}
	if err := db.DB.WithContext(c.Request.Context()).Where("owner_uuid = ?", userUUID).Order("created_at").Find(&items).Error; err != nil {
		log.Printf("trace_id=%s list boards error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"boards": items})
}