package configstest

import (
	"Test-K8s-Pet-Project-Backend/configs"
	"testing"
	"time"
)

func TestParseTTL(t *testing.T) {
	cases := map[string]time.Duration{
		"15m":   15 * time.Minute,
		"30s":   30 * time.Second,
		"2h":    2 * time.Hour,
		"1h30m": 90 * time.Minute,
		"7d":    7 * 24 * time.Hour,
		"10M":   10 * time.Minute,
		"5S":    5 * time.Second,
		"3H":    3 * time.Hour,
		"2D":    48 * time.Hour,
	}
	for in, want := range cases {
		if got := configs.MustParseTTL(in); got != want {
			t.Errorf("MustParseTTL(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseTTLPanicsOnInvalid(t *testing.T) {
	for _, in := range []string{"", "abc", "10x", "d7"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("MustParseTTL(%q) did not panic", in)
				}
			}()
			configs.MustParseTTL(in)
		}()
	}
}

func TestStaticValues(t *testing.T) {
	if got := configs.GetAppPort(); got != ":8989" {
		t.Errorf("GetAppPort() = %q", got)
	}
	if got := configs.WatchInterval(); got != 10*time.Second {
		t.Errorf("WatchInterval() = %v", got)
	}
	if err := configs.EnvLoad(); err != nil {
		t.Errorf("EnvLoad() returned %v", err)
	}
}

func TestEnvGetters(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://u:p@h:5432/d")
	t.Setenv("JWT_ALGORITHM", "EdDSA")
	t.Setenv("JWT_ACCESS_TTL", "15m")
	t.Setenv("JWT_REFRESH_TTL", "7d")
	t.Setenv("JWT_PUBLIC_KEY", `-----BEGIN PUBLIC KEY-----\nabc\n-----END PUBLIC KEY-----`)
	if got := configs.GetDatabaseURL(); got != "postgresql://u:p@h:5432/d" {
		t.Errorf("GetDatabaseURL() = %q", got)
	}
	if got := configs.JWTAlgorithm(); got != "EdDSA" {
		t.Errorf("JWTAlgorithm() = %q", got)
	}
	if got := configs.JWTAccessTTL(); got != 15*time.Minute {
		t.Errorf("JWTAccessTTL() = %v", got)
	}
	if got := configs.JWTRefreshTTL(); got != 7*24*time.Hour {
		t.Errorf("JWTRefreshTTL() = %v", got)
	}
	if got := configs.GetJWTPublicKey(); got != "-----BEGIN PUBLIC KEY-----\nabc\n-----END PUBLIC KEY-----" {
		t.Errorf("GetJWTPublicKey() did not expand newlines: %q", got)
	}
}

func TestS3Getters(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "localhost:9002")
	t.Setenv("S3_ACCESS_KEY", "ak")
	t.Setenv("S3_SECRET_KEY", "sk")
	t.Setenv("S3_BUCKET", "b")
	t.Setenv("S3_REGION", "eu-west-1")
	t.Setenv("S3_USE_SSL", "true")
	if configs.S3Endpoint() != "localhost:9002" || configs.S3AccessKey() != "ak" || configs.S3SecretKey() != "sk" || configs.S3Bucket() != "b" {
		t.Error("S3 getters returned unexpected values")
	}
	if configs.S3Region() != "eu-west-1" {
		t.Errorf("S3Region() = %q", configs.S3Region())
	}
	if !configs.S3UseSSL() {
		t.Error("S3UseSSL() = false, want true")
	}
}

func TestS3Defaults(t *testing.T) {
	t.Setenv("S3_REGION", "")
	t.Setenv("S3_USE_SSL", "")
	if configs.S3Region() != "us-east-1" {
		t.Errorf("S3Region() default = %q", configs.S3Region())
	}
	if configs.S3UseSSL() {
		t.Error("S3UseSSL() default = true, want false")
	}
	t.Setenv("S3_USE_SSL", "yes")
	if configs.S3UseSSL() {
		t.Error("S3UseSSL() must be true only for the literal \"true\"")
	}
}