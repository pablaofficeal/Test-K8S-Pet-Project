package routers

import (
	"Test-K8s-Pet-Project-Backend/utils/middleware"
	"Test-K8s-Pet-Project-Backend/core/api/profile"
	"github.com/gin-gonic/gin"
)

func testrouters(app *gin.RouterGroup) {
	app.GET("/api/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API is working"})
	})
	app.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API is working"})
	})
}

func AuthRouters(app *gin.Engine) {
	v1 := app.Group("/api/v1/main/backend")
	{
		testrouters(v1)

	}
	protected := app.Group("/api/v1/main/backend")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/get/my/profile/info", profile.GetProfileInfo)
		protected.PATCH("/edit/my/name", profile.EditMyName)
		protected.PATCH("/edit/my/email", profile.EditMyEmail)
	}
}