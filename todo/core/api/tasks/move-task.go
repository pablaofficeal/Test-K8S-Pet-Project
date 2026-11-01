package tasks

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"log"
)

type MoveTaskRequest struct {
	NodeID string `json:"node_id"`
	Index *int `json:"index"`
}

func MoveTask(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	taskID := c.Param("id")
	if _, err := uuid.Parse(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var body MoveTaskRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Index == nil || *body.Index < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and index are required"})
		return
	}
	if _, err := uuid.Parse(body.NodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}

	ctx := c.Request.Context()
	var task models.TodoTask
	err := db.DB.WithContext(ctx).Joins("JOIN todo_boards b ON b.id = todo_tasks.board_id").
		Where("todo_tasks.id = ? AND b.owner_uuid = ?", taskID, userUUID).First(&task).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if task.ParentID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only top-level tasks can be moved"})
		return
	}
	var target models.TodoNode
	if err := db.DB.WithContext(ctx).First(&target, "id = ? AND board_id = ?", body.NodeID, task.BoardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	index := *body.Index
	err = db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if target.ID != task.NodeID {
			const query = "WITH RECURSIVE sub AS (SELECT id FROM todo_tasks WHERE id = ? UNION ALL SELECT t.id FROM todo_tasks t JOIN sub ON t.parent_id = sub.id) UPDATE todo_tasks SET node_id = ? WHERE id IN (SELECT id FROM sub)"
			if err := tx.Exec(query, taskID, target.ID).Error; err != nil {
				return err
			}
		}
		siblings := []models.TodoTask{}
		if err := tx.Where("node_id = ? AND parent_id IS NULL AND id <> ?", target.ID, taskID).Order("sort_order, created_at").Find(&siblings).Error; err != nil {
			return err
		}
		if index > len(siblings) {
			index = len(siblings)
		}
		ordered := make([]string, 0, len(siblings)+1)
		for _, sibling := range siblings[:index] {
			ordered = append(ordered, sibling.ID)
		}
		ordered = append(ordered, taskID)
		for _, sibling := range siblings[index:] {
			ordered = append(ordered, sibling.ID)
		}
		for position, id := range ordered {
			if err := tx.Model(&models.TodoTask{}).Where("id = ?", id).Update("sort_order", position).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("trace_id=%s move task error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": taskID, "node_id": target.ID, "index": index})
}