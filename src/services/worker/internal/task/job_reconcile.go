package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type JobReconcilePayload struct {
	MaxAgeMinutes int `json:"max_age_minutes"`
}

func NewJobReconcileTask(maxAgeMinutes int) (*asynq.Task, error) {
	if maxAgeMinutes <= 0 {
		maxAgeMinutes = 45
	}
	data, err := json.Marshal(JobReconcilePayload{MaxAgeMinutes: maxAgeMinutes})
	if err != nil {
		return nil, fmt.Errorf("marshal job reconcile payload: %w", err)
	}
	return asynq.NewTask(TypeJobReconcileStuck, data, asynq.Queue("low")), nil
}

func HandleJobReconcile(ctx context.Context, t *asynq.Task) error {
	var payload JobReconcilePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal job reconcile payload: %w", err)
	}
	if payload.MaxAgeMinutes <= 0 {
		payload.MaxAgeMinutes = 45
	}
	if err := runJobStuckReconciliation(ctx, payload.MaxAgeMinutes); err != nil {
		return fmt.Errorf("run job reconcile: %w", err)
	}
	return nil
}
