package boards

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"Cadence-Cloud-Todo/db"
	"unicode/utf8"
	"net/http"
	"strings"
	"log"
)

type CreateBoardRequest struct {
	Title string `json:"title"`
}

func CreateBoard(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	var body CreateBoardRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" || utf8.RuneCountInString(title) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-200 characters"})
		return
	}
	var count int64
	db.DB.WithContext(c.Request.Context()).Model(&models.TodoBoard{}).Where("owner_uuid = ?", userUUID).Count(&count)
	if count >= 100 {
		c.JSON(http.StatusConflict, gin.H{"error": "board limit reached"})
		return
	}
	board := models.TodoBoard{OwnerUUID: userUUID.(string), Title: title}
	if err := db.DB.WithContext(c.Request.Context()).Create(&board).Error; err != nil {
		log.Printf("trace_id=%s create board error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, board)
}