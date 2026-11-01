package nodes

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"unicode/utf8"
	"net/http"
	"strings"
	"log"
)

type CreateNodeRequest struct {
	Title string `json:"title"`
}

func CreateNode(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	boardID := c.Param("id")
	if _, err := uuid.Parse(boardID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board id"})
		return
	}
	var body CreateNodeRequest
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
	var board models.TodoBoard
	if err := db.DB.WithContext(ctx).First(&board, "id = ? AND owner_uuid = ?", boardID, userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}
	var count int64
	db.DB.WithContext(ctx).Model(&models.TodoNode{}).Where("board_id = ?", boardID).Count(&count)
	if count >= 200 {
		c.JSON(http.StatusConflict, gin.H{"error": "node limit reached"})
		return
	}
	node := models.TodoNode{BoardID: boardID, Title: title, SortOrder: int(count)}
	if err := db.DB.WithContext(ctx).Create(&node).Error; err != nil {
		log.Printf("trace_id=%s create node error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	cache.AppendToOrder(ctx, boardID, node.ID)
	c.JSON(http.StatusCreated, node)
}