package main

import (
	"Low-Code-Bot-Platform-Backend-Service/utils/middleware"
	"Low-Code-Bot-Platform-Backend-Service/core/routers"
	"Low-Code-Bot-Platform-Backend-Service/configs"
	"Low-Code-Bot-Platform-Backend-Service/models"
	"Low-Code-Bot-Platform-Backend-Service/db"
	"github.com/gin-gonic/gin"
	"fmt"
	"log"
	"io"
	"os"
)

func main() {
	if err := configs.EnvLoad(); err != nil {
		log.Fatal(err)
	}
	dsn := configs.GetDatabaseURL()
	interval := configs.WatchInterval()
	db.WaitForDB(dsn, interval)
	db.WaitForTables(interval, &models.User{})

	gin.DisableConsoleColor()
	os.MkdirAll("logs", 0755)
	f, _ := os.OpenFile("logs/backend-service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(f)
	gin.DefaultWriter = io.MultiWriter(f)
	app := gin.New()
	app.Use(
		middleware.TraceMiddleware(),
		gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
			traceID := params.Request.Header.Get("X-Request-ID")
			return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s | trace_id=%s\n",
				params.TimeStamp.Format("2006/01/02 - 15:04:05"),
				params.StatusCode,
				params.Latency,
				params.ClientIP,
				params.Method,
				params.Path,
				traceID,
			)
		}),
		gin.Recovery(),
	)

	routers.AuthRouters(app)
	fmt.Println("Backend Service is running on port", "http://localhost"+configs.GetAppPort())
	app.Run(configs.GetAppPort())
}