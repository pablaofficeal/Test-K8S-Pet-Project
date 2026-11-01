package nodes

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http"
	"strings"
	"math"
	"time"
	"log"
)

type MoveNodeRequest struct {
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
}

func MoveNode(c *gin.Context) {
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
	var body MoveNodeRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.X == nil || body.Y == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "x and y are required"})
		return
	}
	if math.IsNaN(*body.X) || math.IsNaN(*body.Y) || math.Abs(*body.X) > 1e6 || math.Abs(*body.Y) > 1e6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "position is out of range"})
		return
	}

	ctx := c.Request.Context()
	owner, boardID := "", ""
	if cached, err := cache.Client.Get(ctx, "todo:node:"+nodeID).Result(); err == nil {
		owner, boardID, _ = strings.Cut(cached, ":")
	} else {
		var node models.TodoNode
		err := db.DB.WithContext(ctx).Joins("JOIN todo_boards b ON b.id = todo_nodes.board_id").
			Where("todo_nodes.id = ? AND b.owner_uuid = ?", nodeID, userUUID).First(&node).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
			return
		}
		owner, boardID = userUUID.(string), node.BoardID
		cache.Client.Set(ctx, "todo:node:"+nodeID, owner+":"+boardID, time.Hour)
	}
	if owner != userUUID.(string) || boardID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if err := cache.SetPosition(ctx, boardID, nodeID, *body.X, *body.Y); err != nil {
		log.Printf("trace_id=%s redis set position error: %v", traceID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "position storage unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": nodeID, "x": *body.X, "y": *body.Y})
}