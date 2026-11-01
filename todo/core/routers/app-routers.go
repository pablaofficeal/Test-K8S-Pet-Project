package routers

import (
	"Cadence-Cloud-Todo/utils/middleware"
	"Cadence-Cloud-Todo/core/api/boards"
	"Cadence-Cloud-Todo/core/api/tasks"
	"Cadence-Cloud-Todo/core/api/nodes"
	"github.com/gin-gonic/gin"
)

func AppRouters(app *gin.Engine) {
	open := app.Group("/api/v1/main/todo")
	{
		open.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Todo service is working"})
		})
	}
	protected := app.Group("/api/v1/main/todo")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/create/my/board", boards.CreateBoard)
		protected.GET("/get/my/boards", boards.ListBoards)
		protected.GET("/get/my/board/:id", boards.GetBoard)
		protected.PATCH("/edit/my/board/:id", boards.EditBoard)
		protected.DELETE("/delete/my/board/:id", boards.DeleteBoard)
		protected.POST("/import/my/board", boards.ImportBoard)
		protected.GET("/export/my/board/:id", boards.ExportBoard)
		protected.POST("/create/my/board/:id/node", nodes.CreateNode)
		protected.PATCH("/edit/my/node/:id", nodes.EditNode)
		protected.PUT("/reorder/my/board/:id", nodes.ReorderNodes)
		protected.DELETE("/delete/my/node/:id", nodes.DeleteNode)
		protected.POST("/create/my/node/:id/task", tasks.CreateTask)
		protected.PATCH("/edit/my/task/:id", tasks.EditTask)
		protected.PATCH("/move/my/task/:id", tasks.MoveTask)
		protected.DELETE("/delete/my/task/:id", tasks.DeleteTask)
	}
}