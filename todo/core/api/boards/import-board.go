package boards

import (
	"Cadence-Cloud-Todo/markdown"
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"Cadence-Cloud-Todo/db"
	"encoding/json"
	"unicode/utf8"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"log"
	"io"
)

type ImportBoardRequest struct {
	Markdown string `json:"markdown"`
	Title string `json:"title"`
}

func ImportBoard(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	const maxSize = 1 << 20
	raw, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxSize))
	if err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "markdown is larger than 1MB"})
		return
	}
	source, title := string(raw), ""
	if strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		var body ImportBoardRequest
		if err := json.Unmarshal(raw, &body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		source, title = body.Markdown, strings.TrimSpace(body.Title)
	}
	doc, err := markdown.Parse(source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if title == "" {
		title = doc.Title
	}
	if title == "" || utf8.RuneCountInString(title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-200 characters"})
		return
	}
	if len(doc.Nodes) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many sections (max 200)"})
		return
	}

	board := models.TodoBoard{OwnerUUID: userUUID.(string), Title: title}
	total := 0
	err = db.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&board).Error; err != nil {
			return err
		}
		var save func(nodeID string, parentID *string, tasks []*markdown.Task) error
		save = func(nodeID string, parentID *string, tasks []*markdown.Task) error {
			for i, t := range tasks {
				total++
				if total > 5000 {
					return gorm.ErrInvalidData
				}
				tags, _ := json.Marshal(t.Tags)
				row := models.TodoTask{BoardID: board.ID, NodeID: nodeID, ParentID: parentID, Title: t.Title, Description: t.Description, Status: t.Status, Priority: t.Priority, Tags: tags, Deadline: t.Deadline, SortOrder: i}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
				if err := save(nodeID, &row.ID, t.Children); err != nil {
					return err
				}
			}
			return nil
		}
		for i, n := range doc.Nodes {
			node := models.TodoNode{BoardID: board.ID, Title: n.Title, SortOrder: i}
			if err := tx.Create(&node).Error; err != nil {
				return err
			}
			if err := save(node.ID, nil, n.Tasks); err != nil {
				return err
			}
		}
		return nil
	})
	if err == gorm.ErrInvalidData {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many tasks (max 5000)"})
		return
	}
	if err != nil {
		log.Printf("trace_id=%s import board error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"board": board, "nodes": len(doc.Nodes), "tasks": total})
}