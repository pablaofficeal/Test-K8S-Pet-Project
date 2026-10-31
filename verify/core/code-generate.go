package core

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"crypto/rand"
	"math/big"
	"fmt"
)

func GenerateCode(cfg configs.VerifyConfig) (string, error) {
	if cfg.Provider == "test" {
		return cfg.TestCode, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}