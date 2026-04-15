package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGenerateMuxPlaybackToken_MissingSecret(t *testing.T) {
	t.Setenv("MUX_SECRET_TOKEN", "")

	_, _, err := generateMuxPlaybackToken("org_test")
	if err == nil {
		t.Fatal("expected error when MUX_SECRET_TOKEN is missing")
	}
}

func TestGenerateMuxPlaybackToken_ValidJWT(t *testing.T) {
	secret := "test_mux_secret"
	workspaceID := "org_test"
	t.Setenv("MUX_SECRET_TOKEN", secret)

	token, expiresAt, err := generateMuxPlaybackToken(workspaceID)
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}

	enc := base64.RawURLEncoding
	payloadBytes, err := enc.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode JWT payload: %v", err)
	}

	var payload struct {
		Sub string `json:"sub"`
		Aud string `json:"aud"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if payload.Sub != workspaceID {
		t.Fatalf("expected sub=%s, got %s", workspaceID, payload.Sub)
	}
	if payload.Aud != "v" {
		t.Fatalf("expected aud=v, got %s", payload.Aud)
	}

	if payload.Exp <= time.Now().UTC().Unix() {
		t.Fatalf("expected exp in future, got %d", payload.Exp)
	}

	// Verify signature integrity
	unsigned := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	expectedSig := enc.EncodeToString(mac.Sum(nil))
	if parts[2] != expectedSig {
		t.Fatalf("invalid signature; expected %s got %s", expectedSig, parts[2])
	}

	if expiresAt.Unix() != payload.Exp {
		t.Fatalf("expiresAt mismatch; expected %d got %d", expiresAt.Unix(), payload.Exp)
	}
}

func TestVerifyMuxSignature(t *testing.T) {
	body := `{"type":"video.asset.ready","id":"evt_test"}`
	secret := "test_webhook_secret"
	t.Setenv("MUX_WEBHOOK_SECRET", secret)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(body))
	validSig := "mux_v2 " + hex.EncodeToString(h.Sum(nil))

	if !verifyMuxSignature(body, validSig) {
		t.Fatal("expected valid signature to pass")
	}

	if verifyMuxSignature(body, "mux_v2 deadbeef") {
		t.Fatal("expected invalid signature to fail")
	}

	t.Setenv("MUX_WEBHOOK_SECRET", "")
	if verifyMuxSignature(body, validSig) {
		t.Fatal("expected verification to fail when secret is missing")
	}
}
