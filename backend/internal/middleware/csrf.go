package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const csrfCookieName = "tablebite_csrf"

func CSRF(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if secret == "" {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(csrfCookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "csrf token required", http.StatusForbidden)
				return
			}

			header := r.Header.Get("X-CSRF-Token")
			if header == "" || !ValidateCSRFToken(secret, header) || !hmac.Equal([]byte(header), []byte(cookie.Value)) {
				http.Error(w, "csrf token required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewCSRFToken(secret string) (string, error) {
	if secret == "" {
		return "", nil
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	payload := strconv.FormatInt(time.Now().Unix(), 10) + ":" + hex.EncodeToString(buf)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + ":" + sig, nil
}

func ValidateCSRFToken(secret, token string) bool {
	if token == "" || secret == "" {
		return false
	}
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return false
	}
	payload := strings.Join(parts[:2], ":")
	sig := parts[2]
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}
