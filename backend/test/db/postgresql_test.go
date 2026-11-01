package dbtest

import (
	"Test-K8s-Pet-Project-Backend/models"
	"Test-K8s-Pet-Project-Backend/db"
	"github.com/joho/godotenv"
	"net/url"
	"testing"
	"time"
	"os"
)

func TestInitDBFailsOnUnreachableHost(t *testing.T) {
	if err := db.InitDB("postgresql://u:p@127.0.0.1:1/x?connect_timeout=1"); err == nil {
		t.Fatal("expected connection error")
	}
}

func TestDBLifecycleAgainstPostgres(t *testing.T) {
	godotenv.Load("../../../.env")
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		u, err := url.Parse(os.Getenv("DATABASE_URL"))
		if err != nil || u.Host == "" {
			t.Skip("DATABASE_URL is not configured")
		}
		u.Host = "localhost:56799"
		dsn = u.String()
	}
	if err := db.InitDB(dsn); err != nil {
		t.Skipf("postgres is not available: %v", err)
	}
	if db.DB == nil {
		t.Fatal("db.DB is nil after successful InitDB")
	}

	done := make(chan struct{})
	go func() {
		db.WaitForDB(dsn, 100*time.Millisecond)
		db.WaitForTables(100*time.Millisecond, &models.User{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("WaitForDB/WaitForTables did not return for a ready database")
	}
}