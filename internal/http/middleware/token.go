package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrBadToken = errors.New("bad token")

// token format (base64url): userId|unix|sig
// sig = HMAC_SHA256(secret, userId|unix)
func MakeToken(secret []byte, userID string, now time.Time) (string, error) {
	if len(secret) == 0 || userID == "" {
		return "", ErrBadToken
	}
	payload := fmt.Sprintf("%s|%d", userID, now.Unix())
	sig := sign(secret, payload)
	raw := payload + "|" + sig
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

func ParseToken(secret []byte, token string) (string, error) {
	if len(secret) == 0 {
		return "", ErrBadToken
	}
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return "", ErrBadToken
	}
	parts := strings.Split(string(b), "|")
	if len(parts) != 3 {
		return "", ErrBadToken
	}
	userID := parts[0]
	unix := parts[1]
	sig := parts[2]

	payload := userID + "|" + unix
	if !verify(secret, payload, sig) || userID == "" {
		return "", ErrBadToken
	}
	return userID, nil
}

func sign(secret []byte, payload string) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func verify(secret []byte, payload, sig string) bool {
	want := sign(secret, payload)
	return hmac.Equal([]byte(want), []byte(sig))
}
