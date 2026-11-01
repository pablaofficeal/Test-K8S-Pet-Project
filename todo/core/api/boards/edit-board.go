package boards

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"unicode/utf8"
	"net/http"
	"strings"
	"log"
)

type EditBoardRequest struct {
	Title string `json:"title"`
}

func EditBoard(c *gin.Context) {
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
	var body EditBoardRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" || utf8.RuneCountInString(title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-200 characters"})
		return
	}
	ctx := c.Request.Context()
	var board models.TodoBoard
	if err := db.DB.WithContext(ctx).First(&board, "id = ? AND owner_uuid = ?", boardID, userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}
	if err := db.DB.WithContext(ctx).Model(&board).Update("title", title).Error; err != nil {
		log.Printf("trace_id=%s edit board error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": board.ID, "title": title})
}