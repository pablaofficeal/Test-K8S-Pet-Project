package boards

import (
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"Cadence-Cloud-Todo/db"
	"net/http"
	"log"
)

type BoardSummary struct {
	models.TodoBoard
	NodesCount int64 `json:"nodes_count"`
	TasksTotal int64 `json:"tasks_total"`
	TasksDone int64 `json:"tasks_done"`
}

type boardCount struct {
	BoardID string
	Total int64
	Done int64
}

func ListBoards(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	ctx := c.Request.Context()
	found := []models.TodoBoard{}
	if err := db.DB.WithContext(ctx).Where("owner_uuid = ?", userUUID).Order("updated_at DESC").Find(&found).Error; err != nil {
		log.Printf("trace_id=%s list boards error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	nodeCounts := []boardCount{}
	taskCounts := []boardCount{}
	owned := db.DB.WithContext(ctx).Table("todo_boards").Select("id").Where("owner_uuid = ?", userUUID)
	db.DB.WithContext(ctx).Table("todo_nodes").Select("board_id, count(*) AS total").Where("board_id IN (?)", owned).Group("board_id").Scan(&nodeCounts)
	db.DB.WithContext(ctx).Table("todo_tasks").Select("board_id, count(*) AS total, count(*) FILTER (WHERE status = 'done') AS done").Where("board_id IN (?)", owned).Group("board_id").Scan(&taskCounts)

	byBoard := make(map[string]*BoardSummary, len(found))
	items := make([]BoardSummary, len(found))
	for i, board := range found {
		items[i] = BoardSummary{TodoBoard: board}
		byBoard[board.ID] = &items[i]
	}
	for _, row := range nodeCounts {
		if item, ok := byBoard[row.BoardID]; ok {
			item.NodesCount = row.Total
		}
	}
	for _, row := range taskCounts {
		if item, ok := byBoard[row.BoardID]; ok {
			item.TasksTotal, item.TasksDone = row.Total, row.Done
		}
	}
	c.JSON(http.StatusOK, gin.H{"boards": items})
}