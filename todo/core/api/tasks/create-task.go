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

type CreateTaskRequest struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Status string `json:"status"`
	Priority string `json:"priority"`
	Tags []string `json:"tags"`
	Deadline string `json:"deadline"`
	ParentID *string `json:"parent_id"`
}

func CreateTask(c *gin.Context) {
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
	var body CreateTaskRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" || utf8.RuneCountInString(title) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-500 characters"})
		return
	}
	if utf8.RuneCountInString(body.Description) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is too long"})
		return
	}
	status, priority := body.Status, body.Priority
	if status == "" {
		status = "todo"
	}
	if priority == "" {
		priority = "normal"
	}
	if status != "todo" && status != "in_progress" && status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be todo, in_progress or done"})
		return
	}
	if priority != "low" && priority != "normal" && priority != "high" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "priority must be low, normal or high"})
		return
	}
	tags := []string{}
	for _, tag := range body.Tags {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	if len(tags) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many tags (max 20)"})
		return
	}
	for _, tag := range tags {
		if utf8.RuneCountInString(tag) > 32 || strings.ContainsAny(tag, ",|") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag"})
			return
		}
	}
	var deadline *time.Time
	if body.Deadline != "" {
		d, err := time.Parse("2006-01-02", body.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "deadline must look like 2026-09-23"})
			return
		}
		deadline = &d
	}

	ctx := c.Request.Context()
	var node models.TodoNode
	err := db.DB.WithContext(ctx).Joins("JOIN todo_boards b ON b.id = todo_nodes.board_id").
		Where("todo_nodes.id = ? AND b.owner_uuid = ?", nodeID, userUUID).First(&node).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if body.ParentID != nil {
		if _, err := uuid.Parse(*body.ParentID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent id"})
			return
		}
		depth, cursor := 1, *body.ParentID
		for cursor != "" && depth <= 8 {
			var parent models.TodoTask
			if err := db.DB.WithContext(ctx).First(&parent, "id = ? AND node_id = ?", cursor, nodeID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "parent task not found"})
				return
			}
			cursor = ""
			if parent.ParentID != nil {
				cursor = *parent.ParentID
			}
			depth++
		}
		if depth > 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tasks are nested too deep (max 8)"})
			return
		}
	}
	var count int64
	siblings := db.DB.WithContext(ctx).Model(&models.TodoTask{}).Where("node_id = ?", nodeID)
	if body.ParentID != nil {
		siblings = siblings.Where("parent_id = ?", *body.ParentID)
	} else {
		siblings = siblings.Where("parent_id IS NULL")
	}
	siblings.Count(&count)
	rawTags, _ := json.Marshal(tags)
	task := models.TodoTask{BoardID: node.BoardID, NodeID: nodeID, ParentID: body.ParentID, Title: title, Description: body.Description, Status: status, Priority: priority, Tags: rawTags, Deadline: deadline, SortOrder: int(count)}
	if err := db.DB.WithContext(ctx).Create(&task).Error; err != nil {
		log.Printf("trace_id=%s create task error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, task)
}