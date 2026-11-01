package nodes

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"gorm.io/gorm"
	"net/http"
	"log"
)

func DeleteNode(c *gin.Context) {
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
	ctx := c.Request.Context()
	var node models.TodoNode
	err := db.DB.WithContext(ctx).Joins("JOIN todo_boards b ON b.id = todo_nodes.board_id").
		Where("todo_nodes.id = ? AND b.owner_uuid = ?", nodeID, userUUID).First(&node).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	err = db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("node_id = ?", nodeID).Delete(&models.TodoTask{}).Error; err != nil {
			return err
		}
		return tx.Delete(&node).Error
	})
	if err != nil {
		log.Printf("trace_id=%s delete node error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	cache.DeletePosition(ctx, node.BoardID, nodeID)
	c.Status(http.StatusNoContent)
}