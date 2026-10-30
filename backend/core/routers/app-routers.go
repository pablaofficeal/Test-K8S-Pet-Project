package routers

import (
	"test-k8s-test-app/core/api"
	"github.com/gin-gonic/gin"
)

func testRouters(app *gin.Engine) {
    app.GET("/api/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "API is working"})
    })

    app.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "API is working"})
    })

    app.GET("/healthz", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
}

func AuthRouters(app *gin.Engine) {

	v1 := app.Group("/api/v1/auth")
	{
		testRouters(app)
		v1.GET("/hello/k8s", api.TestRouters)
	}
}