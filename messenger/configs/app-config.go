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
	port := os.Getenv("MESSENGER_PORT")
	if port == "" {
		port = "8993"
	}
	return ":" + port
}

func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

func GetJWTPublicKey() string {
	return strings.ReplaceAll(os.Getenv("JWT_PUBLIC_KEY"), "\\n", "\n")
}

func WatchInterval() time.Duration {
	return 10 * time.Second
}
