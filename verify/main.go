package main

import (
	"Test-K8s-Pet-Project-Verify/utils/middleware"
	"Test-K8s-Pet-Project-Verify/providers"
	"Test-K8s-Pet-Project-Verify/routers"
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/models"
	"Test-K8s-Pet-Project-Verify/db"
	"Test-K8s-Pet-Project-Verify/mq"
	"github.com/gin-gonic/gin"
	"os/signal"
	"net/http"
	"context"
	"syscall"
	"time"
	"log"
	"os"
)

func main() {
	if err := configs.EnvLoad(); err != nil {
		log.Fatal(err)
	}
	cfg, err := configs.LoadVerifyConfig()
	if err != nil {
		log.Fatal(err)
	}
	provider, err := providers.New(cfg.Provider)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Provider == "test" {
		log.Printf("WARNING: the test provider is active, every account is verified with the fixed test code")
	}

	interval := configs.WatchInterval()
	db.WaitForDB(configs.GetDatabaseURL(), interval)
	db.WaitForTables(interval, &models.User{}, &models.EmailVerification{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if url := configs.GetRabbitURL(); url != "" {
		go mq.RunConsumer(ctx, url, interval, mq.NewVerificationHandler(db.DB, provider, cfg))
	} else {
		log.Printf("RABBITMQ_VERIFY_URL is empty, only the HTTP API is running")
	}

	gin.DisableConsoleColor()
	app := gin.New()
	app.Use(
		middleware.TraceMiddleware(),
		gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
			traceID := params.Request.Header.Get("X-Request-ID")
			return "[GIN] " + params.TimeStamp.Format("2006/01/02 - 15:04:05") + " | " +
				params.Method + " " + params.Path + " | " + http.StatusText(params.StatusCode) + " | trace_id=" + traceID + "\n"
		}),
		gin.Recovery(),
	)
	routers.VerifyRouters(app, db.DB, provider, cfg)

	server := &http.Server{Addr: configs.GetAppPort(), Handler: app}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("Verify Service is running on port %s", configs.GetAppPort())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}