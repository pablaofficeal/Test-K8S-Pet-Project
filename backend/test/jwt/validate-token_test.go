package jwttest

import (
	"Test-K8s-Pet-Project-Backend/utils/jwt"
	gojwt "github.com/golang-jwt/jwt/v5"
	"crypto/ed25519"
	"encoding/pem"
	"crypto/x509"
	"crypto/rand"
	"testing"
	"time"
)

func TestValidateToken(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	_, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	der, _ := x509.MarshalPKIXPublicKey(pub)
	t.Setenv("JWT_PUBLIC_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))

	future := gojwt.NewNumericDate(time.Now().Add(time.Hour))
	past := gojwt.NewNumericDate(time.Now().Add(-time.Hour))
	cases := []struct {
		name    string
		claims  jwt.Claims
		method  gojwt.SigningMethod
		key     any
		wantErr bool
	}{
		{"valid", jwt.Claims{UUID: "u1", Email: "a@b.com", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: future}}, gojwt.SigningMethodEdDSA, priv, false},
		{"expired", jwt.Claims{UUID: "u1", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: past}}, gojwt.SigningMethodEdDSA, priv, true},
		{"refresh purpose", jwt.Claims{UUID: "u1", Purpose: "refresh", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: future}}, gojwt.SigningMethodEdDSA, priv, true},
		{"empty purpose", jwt.Claims{UUID: "u1", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: future}}, gojwt.SigningMethodEdDSA, priv, true},
		{"empty uuid", jwt.Claims{Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: future}}, gojwt.SigningMethodEdDSA, priv, true},
		{"foreign signature", jwt.Claims{UUID: "u1", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: future}}, gojwt.SigningMethodEdDSA, otherPriv, true},
		{"hmac method", jwt.Claims{UUID: "u1", Purpose: "access", RegisteredClaims: gojwt.RegisteredClaims{ExpiresAt: future}}, gojwt.SigningMethodHS256, []byte("secret"), true},
	}
	for _, tc := range cases {
		token, err := gojwt.NewWithClaims(tc.method, tc.claims).SignedString(tc.key)
		if err != nil {
			t.Fatalf("%s: cannot sign: %v", tc.name, err)
		}
		claims, err := jwt.ValidateToken(token)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
		if err == nil && (claims.UUID != "u1" || claims.Email != "a@b.com") {
			t.Errorf("%s: unexpected claims %+v", tc.name, claims)
		}
	}
}

func TestValidateTokenGarbage(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	der, _ := x509.MarshalPKIXPublicKey(pub)
	t.Setenv("JWT_PUBLIC_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))
	for _, in := range []string{"", "abc", "a.b.c", "Bearer x"} {
		if _, err := jwt.ValidateToken(in); err == nil {
			t.Errorf("ValidateToken(%q) returned nil error", in)
		}
	}
}

func TestValidateTokenWithoutPublicKey(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY", "")
	if _, err := jwt.ValidateToken("a.b.c"); err == nil {
		t.Error("expected error when public key is not configured")
	}
}