package jwt

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"github.com/golang-jwt/jwt/v5"
	"errors"
)

type Claims struct {
	UUID    string `json:"uuid"`
	Email   string `json:"email"`
	SID     string `json:"sid"`
	Ver     int    `json:"ver"`
	Purpose string `json:"purpose,omitempty"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	publicKey, err := jwt.ParseEdPublicKeyFromPEM([]byte(configs.GetJWTPublicKey()))
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UUID == "" {
		return nil, errors.New("empty uuid")
	}

	if claims.Purpose != "access" {
		return nil, errors.New("invalid token purpose")
	}

	return claims, nil
}