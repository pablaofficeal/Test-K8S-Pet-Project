package configs

import (
	"github.com/joho/godotenv"
	"strings"
	"time"
	"os"
)

func EnvLoad() error {
	if err := godotenv.Load(".env"); err == nil {
		return nil
	}
	_ = godotenv.Load("../.env")
	return nil
}

func GetAppPort() string {
	port := os.Getenv("VERIFY_PORT")
	if port == "" {
		port = "8991"
	}
	return ":" + port
}

func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

func GetJWTPublicKey() string {
	return strings.ReplaceAll(os.Getenv("JWT_PUBLIC_KEY"), "\\n", "\n")
}

func GetRabbitURL() string {
	return os.Getenv("RABBITMQ_VERIFY_URL")
}

func WatchInterval() time.Duration {
	return 10 * time.Second
}