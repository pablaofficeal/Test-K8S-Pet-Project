package core

import (
	"crypto/sha256"
	"encoding/hex"
	"crypto/hmac"
)

func HashCode(pepper string, userUUID string, code string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(userUUID + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}