package handler

import (
	"strings"

	"github.com/qvora/api/internal/db"
)

// MuxWebhookPayload is the envelope for all Mux webhook events.
type MuxWebhookPayload struct {
	Type       string `json:"type"`
	EventID    string `json:"id"`
	Attemptnum int    `json:"attemptnum"`
	CreatedAt  string `json:"created_at"`
	Data       struct {
		ID          string `json:"id"`
		Passthrough string `json:"passthrough"`
		PlaybackIDs []struct {
			ID string `json:"id"`
		} `json:"playback_ids"`
	} `json:"data"`
}

// isDuplicateMuxWebhook returns true when rowsAffected == 0, meaning this
// event ID was already recorded (duplicate delivery).
func isDuplicateMuxWebhook(rowsAffected int64) bool {
	return rowsAffected == 0
}

// isVariantPlaybackReady returns true when the variant has reached the
// "complete" status and has both a Mux asset ID and a Mux playback ID.
func isVariantPlaybackReady(status string, muxAssetID, muxPlaybackID *string) bool {
	if !strings.EqualFold(strings.TrimSpace(status), "complete") {
		return false
	}
	return muxAssetID != nil && muxPlaybackID != nil
}

// areAllVariantsComplete returns true when the slice is non-empty and every
// variant has reached the "complete" status.
func areAllVariantsComplete(variants []db.Variant) bool {
	if len(variants) == 0 {
		return false
	}
	for _, v := range variants {
		if !strings.EqualFold(strings.TrimSpace(v.Status), "complete") {
			return false
		}
	}
	return true
}
