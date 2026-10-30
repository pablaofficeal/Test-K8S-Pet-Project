package main

import (
	"test-k8s-test-app/core/routers"
	"test-k8s-test-app/configs"
	"github.com/gin-gonic/gin"
	"fmt"
	"io"
	"os"
)

func main() {
	gin.DisableConsoleColor()
	os.MkdirAll("logs", 0755)
	f, _ := os.OpenFile("logs/auth-service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	gin.DefaultWriter = io.MultiWriter(f)
	app := gin.Default()

	routers.AuthRouters(app)
	fmt.Println("Auth Service is running on port", "http://localhost"+configs.GetAppPort())
	app.Run(configs.GetAppPort())
}