package configs

import (
	"github.com/joho/godotenv"
	"strings"
	"time"
	"fmt"
	"os"
)

func EnvLoad() error {
	if err := godotenv.Load(".env"); err == nil {
		return nil
	}
	return nil
}

func GetAppPort() string {
	APP_PORT := "8989"
	return ":" + APP_PORT
}

func GetDatabaseURL() string {
	DATABASE_URL := os.Getenv("DATABASE_URL")
	return DATABASE_URL
}

func GetJWTPublicKey() string {
	JWT_PUBLIC_KEY := os.Getenv("JWT_PUBLIC_KEY")
	return strings.ReplaceAll(JWT_PUBLIC_KEY, "\\n", "\n")
}

func JWTAlgorithm() string {
	JWT_ALGORITHM := os.Getenv("JWT_ALGORITHM")
	return JWT_ALGORITHM
}

func JWTAccessTTL() time.Duration {
	JWT_ACCESS_TTL := os.Getenv("JWT_ACCESS_TTL")
	return MustParseTTL(JWT_ACCESS_TTL)
}

func JWTRefreshTTL() time.Duration {
	JWT_REFRESH_TTL := os.Getenv("JWT_REFRESH_TTL")
	return MustParseTTL(JWT_REFRESH_TTL)
}

func WatchInterval() time.Duration {
	return 10 * time.Second
}

func MustParseTTL(value string) time.Duration {
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	var n int
	var unit byte
	if _, err := fmt.Sscanf(value, "%d%c", &n, &unit); err == nil {
		switch unit {
		case 's', 'S':
			return time.Duration(n) * time.Second
		case 'm', 'M':
			return time.Duration(n) * time.Minute
		case 'h', 'H':
			return time.Duration(n) * time.Hour
		case 'd', 'D':
			return time.Duration(n) * 24 * time.Hour
		}
	}
	panic("invalid TTL value: " + value)
}

func S3Endpoint() string {
	return os.Getenv("S3_ENDPOINT") 
}
func S3AccessKey() string { 
	return os.Getenv("S3_ACCESS_KEY") 
}
func S3SecretKey() string { 
	return os.Getenv("S3_SECRET_KEY") 
}
func S3Bucket() string { 
	return os.Getenv("S3_BUCKET") 
}
func S3Region() string {
	if r := os.Getenv("S3_REGION"); r != "" {
		return r
	}
	return "us-east-1"
}
func S3UseSSL() bool { 
	return os.Getenv("S3_USE_SSL") == "true" 
}
