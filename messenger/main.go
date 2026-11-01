package main

import (
	"Cadence-Cloud-Messenger/utils/middleware"
	"Cadence-Cloud-Messenger/core/routers"
	"Cadence-Cloud-Messenger/configs"
	"Cadence-Cloud-Messenger/models"
	"Cadence-Cloud-Messenger/db"
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
	interval := configs.WatchInterval()
	db.WaitForDB(configs.GetDatabaseURL(), interval)
	db.WaitForTables(interval, &models.MessengerKey{}, &models.MessengerMessage{})

	gin.DisableConsoleColor()
	os.MkdirAll("logs", 0755)
	f, _ := os.OpenFile("logs/messenger-service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(f)
	gin.DefaultWriter = io.MultiWriter(f)
	app := gin.New()
	app.Use(
		middleware.TraceMiddleware(),
		gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s | trace_id=%s\n",
				params.TimeStamp.Format("2006/01/02 - 15:04:05"),
				params.StatusCode,
				params.Latency,
				params.ClientIP,
				params.Method,
				params.Path,
				params.Request.Header.Get("X-Request-ID"),
			)
		}),
		gin.Recovery(),
	)
	routers.AppRouters(app)
	fmt.Println("Messenger Service is running on port", "http://localhost"+configs.GetAppPort())
	app.Run(configs.GetAppPort())
}