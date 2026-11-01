package nodes

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"unicode/utf8"
	"net/http"
	"strings"
	"log"
)

type EditNodeRequest struct {
	Title string `json:"title"`
}

func EditNode(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	nodeID := c.Param("id")
	if _, err := uuid.Parse(nodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}
	var body EditNodeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" || utf8.RuneCountInString(title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-200 characters"})
		return
	}
	ctx := c.Request.Context()
	var node models.TodoNode
	err := db.DB.WithContext(ctx).Joins("JOIN todo_boards b ON b.id = todo_nodes.board_id").
		Where("todo_nodes.id = ? AND b.owner_uuid = ?", nodeID, userUUID).First(&node).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if err := db.DB.WithContext(ctx).Model(&node).Update("title", title).Error; err != nil {
		log.Printf("trace_id=%s edit node error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": node.ID, "title": title})
}