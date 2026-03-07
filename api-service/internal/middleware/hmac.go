package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func HMACAuth(secret string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sigHeader := r.Header.Get("X-Signature")
			if sigHeader == "" {
				http.Error(w, "missing signature", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(sigHeader, "=", 2)
			if len(parts) != 2 || parts[0] != "sha256" {
				http.Error(w, "invalid signature format", http.StatusUnauthorized)
				return
			}
			expectedHex := parts[1]

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("middleware.hmac: read body", zap.Error(err))
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}
			_ = r.Body.Close()

			mac := hmac.New(sha256.New, []byte(secret))
			if _, err := mac.Write(bodyBytes); err != nil {
				logger.Error("middleware.hmac: compute hmac", zap.Error(err))
				http.Error(w, "failed to verify signature", http.StatusInternalServerError)
				return
			}
			sum := mac.Sum(nil)
			expected, err := hex.DecodeString(expectedHex)
			if err != nil {
				http.Error(w, "invalid signature encoding", http.StatusUnauthorized)
				return
			}

			if !hmac.Equal(sum, expected) {
				http.Error(w, "signature mismatch", http.StatusUnauthorized)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			next.ServeHTTP(w, r)
		})
	}
}

