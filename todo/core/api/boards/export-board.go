package boards

import (
	"Cadence-Cloud-Todo/markdown"
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"encoding/json"
	"net/http"
	"log"
)

func ExportBoard(c *gin.Context) {
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
		log.Printf("trace_id=%s export nodes error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if err := db.DB.WithContext(ctx).Where("board_id = ?", boardID).Order("sort_order, created_at").Find(&tasks).Error; err != nil {
		log.Printf("trace_id=%s export tasks error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	byID := map[string]*markdown.Task{}
	for _, t := range tasks {
		tags := []string{}
		json.Unmarshal(t.Tags, &tags)
		byID[t.ID] = &markdown.Task{Title: t.Title, Description: t.Description, Status: t.Status, Priority: t.Priority, Tags: tags, Deadline: t.Deadline}
	}
	doc := &markdown.Document{Title: board.Title}
	nodeIndex := map[string]*markdown.Node{}
	for _, n := range nodes {
		node := &markdown.Node{Title: n.Title}
		nodeIndex[n.ID] = node
		doc.Nodes = append(doc.Nodes, node)
	}
	for _, t := range tasks {
		if t.ParentID != nil {
			if parent, ok := byID[*t.ParentID]; ok {
				parent.Children = append(parent.Children, byID[t.ID])
				continue
			}
		}
		if node, ok := nodeIndex[t.NodeID]; ok {
			node.Tasks = append(node.Tasks, byID[t.ID])
		}
	}
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(markdown.Render(doc)))
}