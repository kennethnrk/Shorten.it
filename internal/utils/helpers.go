package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateLongURLHash(longURL string) string {
	hash := sha256.New()
	hash.Write([]byte(longURL))
	return hex.EncodeToString(hash.Sum(nil))
}
