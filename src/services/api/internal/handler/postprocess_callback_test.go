package handler

import "testing"

func TestIsIdempotentCallback(t *testing.T) {
	if !isIdempotentCallback(0) {
		t.Fatal("expected rowsAffected=0 to be idempotent")
	}
	if isIdempotentCallback(1) {
		t.Fatal("expected rowsAffected>0 to be non-idempotent")
	}
}

func TestTargetVariantStatus_CompleteIsTerminal(t *testing.T) {
	status := targetVariantStatus("failed", "complete")
	if status != "complete" {
		t.Fatalf("expected terminal complete status, got %s", status)
	}
}

func TestTargetVariantStatus_SuccessPath(t *testing.T) {
	status := targetVariantStatus("success", "queued")
	if status != "postprocessing" {
		t.Fatalf("expected postprocessing for success callback, got %s", status)
	}
}

func TestTargetVariantStatus_FailurePath(t *testing.T) {
	status := targetVariantStatus("failed", "postprocessing")
	if status != "failed" {
		t.Fatalf("expected failed status for failed callback, got %s", status)
	}
}

func TestTargetJobStatus_CompleteIsTerminal(t *testing.T) {
	status := targetJobStatus("success", "complete")
	if status != "complete" {
		t.Fatalf("expected terminal complete status, got %s", status)
	}
}

func TestTargetJobStatus_SuccessRecoveryPath(t *testing.T) {
	status := targetJobStatus("success", "failed")
	if status != "postprocessing" {
		t.Fatalf("expected postprocessing recovery path, got %s", status)
	}
}

func TestTargetJobStatus_FailurePath(t *testing.T) {
	status := targetJobStatus("failed", "generating")
	if status != "failed" {
		t.Fatalf("expected failed status for failed callback, got %s", status)
	}
}
