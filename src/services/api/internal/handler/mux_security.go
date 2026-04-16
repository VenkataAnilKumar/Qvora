package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

// generateMuxPlaybackToken produces an HS256-signed JWT for Mux signed playback.
// The token carries sub=workspaceID, aud="v", exp=15 minutes from now.
func generateMuxPlaybackToken(workspaceID string) (string, time.Time, error) {
	secret := os.Getenv("MUX_SECRET_TOKEN")
	if secret == "" {
		return "", time.Time{}, errors.New("MUX_SECRET_TOKEN is required")
	}

	expiresAt := time.Now().UTC().Add(15 * time.Minute)

	headerJSON, _ := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	payloadJSON, _ := json.Marshal(map[string]any{
		"sub": workspaceID,
		"aud": "v",
		"exp": expiresAt.Unix(),
	})

	enc := base64.RawURLEncoding
	h := enc.EncodeToString(headerJSON)
	p := enc.EncodeToString(payloadJSON)
	unsigned := h + "." + p

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	sig := enc.EncodeToString(mac.Sum(nil))

	return unsigned + "." + sig, expiresAt, nil
}

// verifyMuxSignature validates a Mux webhook signature of the form "mux_v2 <hex>".
// Returns false if MUX_WEBHOOK_SECRET is not set or the signature does not match.
func verifyMuxSignature(body, signature string) bool {
	secret := os.Getenv("MUX_WEBHOOK_SECRET")
	if secret == "" {
		return false
	}

	const prefix = "mux_v2 "
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	provided := strings.TrimPrefix(signature, prefix)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(provided), []byte(expected))
}
