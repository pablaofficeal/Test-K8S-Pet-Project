package routers

import (
	"Cadence-Cloud-Messenger/core/api/messages"
	"Cadence-Cloud-Messenger/utils/middleware"
	"Cadence-Cloud-Messenger/core/api/users"
	"Cadence-Cloud-Messenger/core/api/keys"
	"github.com/gin-gonic/gin"
)

func AppRouters(app *gin.Engine) {
	open := app.Group("/api/v1/main/messenger")
	{
		open.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Messenger service is working"})
		})
	}
	protected := app.Group("/api/v1/main/messenger")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/put/my/key", keys.PutMyKey)
		protected.GET("/get/my/key", keys.GetMyKey)
		protected.PUT("/put/my/key/backup", keys.PutKeyBackup)
		protected.GET("/get/user/key/:uuid", keys.GetUserKey)
		protected.GET("/search/users", users.SearchUsers)
		protected.POST("/send/message", messages.SendMessage)
		protected.GET("/get/my/messages", messages.ListMessages)
		protected.GET("/get/my/conversations", messages.ListConversations)
		protected.PATCH("/mark/my/messages/read", messages.MarkRead)
		protected.DELETE("/delete/my/message/:id", messages.DeleteMessage)
	}
}