package routers

import (
	"Test-K8s-Pet-Project-Verify/utils/middleware"
	"Test-K8s-Pet-Project-Verify/providers"
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func VerifyRouters(app *gin.Engine, gdb *gorm.DB, provider providers.Provider, cfg configs.VerifyConfig) {
	public := app.Group("/api/v1/main/verify")
	{
		public.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "API is working"})
		})
	}
	protected := app.Group("/api/v1/main/verify")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/status", api.VerifyStatus(gdb, cfg))
		protected.POST("/resend", api.VerifyResend(gdb, provider, cfg))
		protected.POST("/confirm", api.VerifyConfirm(gdb, cfg))
	}
}