package handler

import (
	"encoding/json"
	"testing"

	"github.com/qvora/api/internal/db"
)

func TestMuxWebhookPayload_HasEventIDForDedup(t *testing.T) {
	body := []byte(`{"type":"video.asset.ready","id":"evt_123","attemptnum":1,"created_at":"2026-04-15T00:00:00Z","data":{"id":"asset_1","passthrough":"variant_1","playback_ids":[{"id":"play_1"}]}}`)

	var payload MuxWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if payload.EventID != "evt_123" {
		t.Fatalf("expected EventID evt_123, got %s", payload.EventID)
	}
}

func TestIsDuplicateMuxWebhook(t *testing.T) {
	if !isDuplicateMuxWebhook(0) {
		t.Fatal("expected rowsAffected=0 to be treated as duplicate")
	}
	if isDuplicateMuxWebhook(1) {
		t.Fatal("expected rowsAffected>0 to be treated as first-seen event")
	}
}

func TestIsVariantPlaybackReady(t *testing.T) {
	asset := "asset_1"
	play := "play_1"

	if !isVariantPlaybackReady("complete", &asset, &play) {
		t.Fatal("expected complete + mux ids to be playback-ready")
	}
	if isVariantPlaybackReady("postprocessing", &asset, &play) {
		t.Fatal("expected non-complete status to be not-ready")
	}
	if isVariantPlaybackReady("complete", nil, &play) {
		t.Fatal("expected missing asset id to be not-ready")
	}
	if isVariantPlaybackReady("complete", &asset, nil) {
		t.Fatal("expected missing playback id to be not-ready")
	}
}

func TestAreAllVariantsComplete(t *testing.T) {
	if areAllVariantsComplete(nil) {
		t.Fatal("expected empty variants to be not complete")
	}

	complete := []db.Variant{{Status: "complete"}, {Status: "complete"}}
	if !areAllVariantsComplete(complete) {
		t.Fatal("expected all-complete variants to return true")
	}

	withFailed := []db.Variant{{Status: "complete"}, {Status: "failed"}}
	if areAllVariantsComplete(withFailed) {
		t.Fatal("expected failed variant to block completion")
	}

	withPending := []db.Variant{{Status: "postprocessing"}, {Status: "complete"}}
	if areAllVariantsComplete(withPending) {
		t.Fatal("expected pending variant to block completion")
	}
}
