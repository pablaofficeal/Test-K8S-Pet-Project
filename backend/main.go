package main

import (
	"Test-K8s-Pet-Project-Backend/utils/middleware"
	"Test-K8s-Pet-Project-Backend/core/routers"
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/gin-gonic/gin"
	"context"
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

	if err := storage.InitS3(context.Background()); err != nil {
		log.Fatalf("s3 init failed: %v", err)
	}

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