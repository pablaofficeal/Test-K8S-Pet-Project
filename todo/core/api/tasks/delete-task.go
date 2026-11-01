package tasks

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"net/http"
	"log"
)

func DeleteTask(c *gin.Context) {
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
	ctx := c.Request.Context()
	var task models.TodoTask
	err := db.DB.WithContext(ctx).Joins("JOIN todo_boards b ON b.id = todo_tasks.board_id").
		Where("todo_tasks.id = ? AND b.owner_uuid = ?", taskID, userUUID).First(&task).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	const query = "WITH RECURSIVE sub AS (SELECT id FROM todo_tasks WHERE id = ? UNION ALL SELECT t.id FROM todo_tasks t JOIN sub ON t.parent_id = sub.id) DELETE FROM todo_tasks WHERE id IN (SELECT id FROM sub)"
	if err := db.DB.WithContext(ctx).Exec(query, taskID).Error; err != nil {
		log.Printf("trace_id=%s delete task error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}