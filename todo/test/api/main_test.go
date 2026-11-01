package apitest

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
	"Cadence-Cloud-Todo/cache"
	"Cadence-Cloud-Todo/db"
	"net/url"
	"testing"
	"os"
)

var dbReady bool
var redisServer *miniredis.Miniredis

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env")
	redisServer, _ = miniredis.Run()
	cache.Client = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if u, err := url.Parse(os.Getenv("DATABASE_URL")); err == nil && u.Host != "" {
			u.Host = "localhost:56799"
			dsn = u.String()
		}
	}
	if dsn != "" && db.InitDB(dsn) == nil {
		if sqlDB, err := db.DB.DB(); err == nil && sqlDB.Ping() == nil {
			dbReady = true
		}
	}
	code := m.Run()
	redisServer.Close()
	os.Exit(code)
}