package task

import (
	"strings"
	"testing"
)

func TestPostprocessRequestID_UsesProvidedValue(t *testing.T) {
	payload := PostprocessPayload{PostprocessRequestID: "req_fixed_123"}
	got := postprocessRequestID(payload)
	if got != "req_fixed_123" {
		t.Fatalf("expected provided request id to be used, got %s", got)
	}
}

func TestPostprocessRequestID_GeneratesWhenMissing(t *testing.T) {
	payload := PostprocessPayload{}
	got := postprocessRequestID(payload)
	if strings.TrimSpace(got) == "" {
		t.Fatal("expected generated request id to be non-empty")
	}
}
