package nodes

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"Cadence-Cloud-Todo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"log"
)

type ReorderNodesRequest struct {
	Order []string `json:"order"`
}

func ReorderNodes(c *gin.Context) {
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
	var body ReorderNodesRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Order == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order is required"})
		return
	}
	ctx := c.Request.Context()
	var board models.TodoBoard
	if err := db.DB.WithContext(ctx).First(&board, "id = ? AND owner_uuid = ?", boardID, userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}
	existing := []string{}
	db.DB.WithContext(ctx).Model(&models.TodoNode{}).Where("board_id = ?", boardID).Pluck("id", &existing)
	known := make(map[string]bool, len(existing))
	for _, id := range existing {
		known[strings.ToLower(id)] = true
	}
	order := make([]string, 0, len(body.Order))
	seen := make(map[string]bool, len(body.Order))
	for _, id := range body.Order {
		id = strings.ToLower(id)
		if !known[id] || seen[id] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order must list every block of the board exactly once"})
			return
		}
		seen[id] = true
		order = append(order, id)
	}
	if len(order) != len(existing) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order must list every block of the board exactly once"})
		return
	}
	if err := cache.SetOrder(ctx, boardID, order); err != nil {
		log.Printf("trace_id=%s redis set order error: %v", traceID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "order storage unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order})
}