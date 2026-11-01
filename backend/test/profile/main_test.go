package profiletest

import (
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/joho/godotenv"
	"net/url"
	"testing"
	"os"
)

var dbReady bool

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env")
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
	os.Exit(m.Run())
}