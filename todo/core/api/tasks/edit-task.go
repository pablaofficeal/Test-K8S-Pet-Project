package tasks

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"encoding/json"
	"unicode/utf8"
	"net/http"
	"strings"
	"time"
	"log"
)

type EditTaskRequest struct {
	Title *string `json:"title"`
	Description *string `json:"description"`
	Status *string `json:"status"`
	Priority *string `json:"priority"`
	Tags *[]string `json:"tags"`
	Deadline *string `json:"deadline"`
}

func EditTask(c *gin.Context) {
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
	var body EditTaskRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	changes := map[string]any{}
	if body.Title != nil {
		title := strings.TrimSpace(*body.Title)
		if title == "" || utf8.RuneCountInString(title) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-500 characters"})
			return
		}
		changes["title"] = title
	}
	if body.Description != nil {
		if utf8.RuneCountInString(*body.Description) > 10000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "description is too long"})
			return
		}
		changes["description"] = *body.Description
	}
	if body.Status != nil {
		if *body.Status != "todo" && *body.Status != "in_progress" && *body.Status != "done" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be todo, in_progress or done"})
			return
		}
		changes["status"] = *body.Status
	}
	if body.Priority != nil {
		if *body.Priority != "low" && *body.Priority != "normal" && *body.Priority != "high" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "priority must be low, normal or high"})
			return
		}
		changes["priority"] = *body.Priority
	}
	if body.Tags != nil {
		tags := []string{}
		for _, tag := range *body.Tags {
			if tag = strings.TrimSpace(tag); tag != "" {
				if utf8.RuneCountInString(tag) > 32 || strings.ContainsAny(tag, ",|") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag"})
					return
				}
				tags = append(tags, tag)
			}
		}
		if len(tags) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "too many tags (max 20)"})
			return
		}
		raw, _ := json.Marshal(tags)
		changes["tags"] = string(raw)
	}
	if body.Deadline != nil {
		if *body.Deadline == "" {
			changes["deadline"] = nil
		} else {
			d, err := time.Parse("2006-01-02", *body.Deadline)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "deadline must look like 2026-09-23"})
				return
			}
			changes["deadline"] = d
		}
	}
	if len(changes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to change"})
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
	if err := db.DB.WithContext(ctx).Model(&task).Updates(changes).Error; err != nil {
		log.Printf("trace_id=%s edit task error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	db.DB.WithContext(ctx).First(&task, "id = ?", taskID)
	c.JSON(http.StatusOK, task)
}