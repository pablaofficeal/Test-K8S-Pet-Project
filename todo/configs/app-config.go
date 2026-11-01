package configs

import (
	"github.com/joho/godotenv"
	"strings"
	"time"
	"os"
)

func EnvLoad() error {
	godotenv.Load(".env")
	return nil
}

func GetAppPort() string {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "8992"
	}
	return ":" + port
}

func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

func GetRedisURL() string {
	return os.Getenv("REDIS_URL")
}

func GetJWTPublicKey() string {
	return strings.ReplaceAll(os.Getenv("JWT_PUBLIC_KEY"), "\\n", "\n")
}

func PositionFlushInterval() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("TODO_POSITION_FLUSH_INTERVAL")); err == nil && d > 0 {
		return d
	}
	return 3 * time.Second
}

func WatchInterval() time.Duration {
	return 10 * time.Second
}