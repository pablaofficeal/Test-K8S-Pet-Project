package main

import (
	"Cadence-Cloud-Todo/utils/middleware"
	"Cadence-Cloud-Todo/core/routers"
	"Cadence-Cloud-Todo/configs"
	"Cadence-Cloud-Todo/models"
	"github.com/gin-gonic/gin"
	"Cadence-Cloud-Todo/cache"
	"Cadence-Cloud-Todo/db"
	"os/signal"
	"net/http"
	"context"
	"syscall"
	"time"
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
	db.WaitForTables(interval, &models.TodoBoard{}, &models.TodoNode{}, &models.TodoTask{})
	if err := cache.InitRedis(context.Background(), configs.GetRedisURL()); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}

	gin.DisableConsoleColor()
	os.MkdirAll("logs", 0755)
	f, _ := os.OpenFile("logs/todo-service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	flusherDone := make(chan struct{})
	cache.StartFlusher(ctx, configs.PositionFlushInterval(), flusherDone)

	server := &http.Server{Addr: configs.GetAppPort(), Handler: app}
	go func() {
		fmt.Println("Todo Service is running on port", "http://localhost"+configs.GetAppPort())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdown)
	<-flusherDone
}