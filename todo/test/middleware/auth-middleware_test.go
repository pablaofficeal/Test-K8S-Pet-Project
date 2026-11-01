package middlewaretest

import (
	"Cadence-Cloud-Todo/utils/middleware"
	gojwt "github.com/golang-jwt/jwt/v5"
	"Cadence-Cloud-Todo/utils/jwt"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"crypto/ed25519"
	"encoding/pem"
	"crypto/x509"
	"crypto/rand"
	"testing"
	"time"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	_, other, _ := ed25519.GenerateKey(rand.Reader)
	der, _ := x509.MarshalPKIXPublicKey(pub)
	t.Setenv("JWT_PUBLIC_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))
	exp := gojwt.NewNumericDate(time.Now().Add(time.Hour))
	access, _ := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, jwt.Claims{UUID: "u1", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: exp}}).SignedString(priv)
	refresh, _ := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, jwt.Claims{UUID: "u1", Purpose: "refresh", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: exp}}).SignedString(priv)
	forged, _ := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, jwt.Claims{UUID: "u1", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: exp}}).SignedString(other)

	for header, want := range map[string]int{
		"":                 401,
		"Basic " + access:  401,
		"Bearer garbage":   401,
		"Bearer " + refresh: 401,
		"Bearer " + forged: 401,
		"Bearer " + access: 200,
	} {
		var seen any
		r := gin.New()
		r.GET("/p", middleware.AuthMiddleware(), func(c *gin.Context) { seen, _ = c.Get("userUUID"); c.Status(200) })
		req := httptest.NewRequest("GET", "/p", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != want {
			t.Errorf("%.20q: got %d, want %d", header, w.Code, want)
		}
		if want == 200 && seen != "u1" {
			t.Errorf("userUUID = %v", seen)
		}
	}
}

func TestTraceMiddlewareKeepsValidAndReplacesInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for in, same := range map[string]bool{"req-1_A.b": true, "bad id": false, "": false} {
		r := gin.New()
		r.Use(middleware.TraceMiddleware())
		r.GET("/t", func(c *gin.Context) { c.Status(200) })
		req := httptest.NewRequest("GET", "/t", nil)
		if in != "" {
			req.Header.Set("X-Request-ID", in)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		got := w.Header().Get("X-Request-ID")
		if got == "" || (same && got != in) || (!same && got == in) {
			t.Errorf("%q -> %q", in, got)
		}
	}
}