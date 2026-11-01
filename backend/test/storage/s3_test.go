package storagetest

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"Test-K8s-Pet-Project-Backend/storage"
	"github.com/joho/godotenv"
	"context"
	"testing"
	"time"
)

func TestInitS3RejectsInvalidEndpoint(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "http://not-an-endpoint")
	if err := storage.InitS3(context.Background()); err == nil {
		t.Fatal("expected error for endpoint with scheme")
	}
}

func TestInitS3FailsOnUnreachableServer(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "127.0.0.1:1")
	t.Setenv("S3_ACCESS_KEY", "k")
	t.Setenv("S3_SECRET_KEY", "s")
	t.Setenv("S3_BUCKET", "b")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := storage.InitS3(ctx); err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestInitS3CreatesBucketAgainstRustFS(t *testing.T) {
	godotenv.Load("../../../.env")
	ctx := context.Background()
	if err := storage.InitS3(ctx); err != nil {
		t.Skipf("s3 is not available: %v", err)
	}
	exists, err := storage.Client.BucketExists(ctx, configs.S3Bucket())
	if err != nil || !exists {
		t.Fatalf("bucket %q must exist after InitS3: exists=%v err=%v", configs.S3Bucket(), exists, err)
	}
	if err := storage.InitS3(ctx); err != nil {
		t.Fatalf("InitS3 must be idempotent: %v", err)
	}
}