package boards

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http"
	"log"
)

func GetBoard(c *gin.Context) {
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
	ctx := c.Request.Context()
	var board models.TodoBoard
	if err := db.DB.WithContext(ctx).First(&board, "id = ? AND owner_uuid = ?", boardID, userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}
	nodes := []models.TodoNode{}
	tasks := []models.TodoTask{}
	if err := db.DB.WithContext(ctx).Where("board_id = ?", boardID).Order("sort_order, created_at").Find(&nodes).Error; err != nil {
		log.Printf("trace_id=%s load nodes error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if err := db.DB.WithContext(ctx).Where("board_id = ?", boardID).Order("sort_order, created_at").Find(&tasks).Error; err != nil {
		log.Printf("trace_id=%s load tasks error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if hot, err := cache.GetPositions(ctx, boardID); err == nil {
		for i := range nodes {
			if p, ok := hot[nodes[i].ID]; ok {
				nodes[i].PosX, nodes[i].PosY = p[0], p[1]
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"board": board, "nodes": nodes, "tasks": tasks})
}