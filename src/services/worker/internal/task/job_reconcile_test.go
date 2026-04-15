package task

import (
	"encoding/json"
	"testing"
)

func TestNewJobReconcileTask_DefaultsMaxAge(t *testing.T) {
	task, err := NewJobReconcileTask(0)
	if err != nil {
		t.Fatalf("expected no error creating task: %v", err)
	}

	if task.Type() != TypeJobReconcileStuck {
		t.Fatalf("expected task type %s, got %s", TypeJobReconcileStuck, task.Type())
	}

	var payload JobReconcilePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if payload.MaxAgeMinutes != 45 {
		t.Fatalf("expected default max age 45, got %d", payload.MaxAgeMinutes)
	}
}

func TestNewJobReconcileTask_UsesProvidedMaxAge(t *testing.T) {
	task, err := NewJobReconcileTask(120)
	if err != nil {
		t.Fatalf("expected no error creating task: %v", err)
	}

	var payload JobReconcilePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if payload.MaxAgeMinutes != 120 {
		t.Fatalf("expected max age 120, got %d", payload.MaxAgeMinutes)
	}
}
