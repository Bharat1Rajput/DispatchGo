package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// creates a SHA-256 HMAC signature for the given payload and secret.
func GenerateSignature(payload []byte, secret string) string {
	if secret == "" {
		return ""
	}

	//Create a new HMAC hasher
	h := hmac.New(sha256.New, []byte(secret))

	//Write the body content
	h.Write(payload)

	//Get the hex string (e.g., "a1b2c3...")
	return hex.EncodeToString(h.Sum(nil))
}
