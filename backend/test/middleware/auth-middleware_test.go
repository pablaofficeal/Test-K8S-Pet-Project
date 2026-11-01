package middlewaretest

import (
	"Test-K8s-Pet-Project-Backend/utils/middleware"
	"Test-K8s-Pet-Project-Backend/utils/jwt"
	gojwt "github.com/golang-jwt/jwt/v5"
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
	der, _ := x509.MarshalPKIXPublicKey(pub)
	t.Setenv("JWT_PUBLIC_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))
	access, _ := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, jwt.Claims{UUID: "u1", Email: "a@b.com", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString(priv)
	refresh, _ := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, jwt.Claims{UUID: "u1", Purpose: "refresh", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString(priv)

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"no header", "", 401},
		{"no scheme", access, 401},
		{"wrong scheme", "Basic " + access, 401},
		{"lowercase scheme", "bearer " + access, 401},
		{"garbage token", "Bearer abc", 401},
		{"refresh token", "Bearer " + refresh, 401},
		{"valid token", "Bearer " + access, 200},
	}
	for _, tc := range cases {
		var gotUUID, gotEmail any
		r := gin.New()
		r.GET("/p", middleware.AuthMiddleware(), func(c *gin.Context) {
			gotUUID, _ = c.Get("userUUID")
			gotEmail, _ = c.Get("userEmail")
			c.Status(200)
		})
		req := httptest.NewRequest("GET", "/p", nil)
		if tc.header != "" {
			req.Header.Set("Authorization", tc.header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, w.Code, tc.want)
		}
		if tc.want == 200 && (gotUUID != "u1" || gotEmail != "a@b.com") {
			t.Errorf("%s: context not populated: %v %v", tc.name, gotUUID, gotEmail)
		}
		if tc.want == 401 && gotUUID != nil {
			t.Errorf("%s: handler must not run on 401", tc.name)
		}
	}
}