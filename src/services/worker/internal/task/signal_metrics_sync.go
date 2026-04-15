package task

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type SignalMetricsSyncPayload struct{}

func NewSignalMetricsSyncTask() (*asynq.Task, error) {
	return asynq.NewTask(TypeSignalMetricsSync, []byte(`{}`), asynq.Queue("low")), nil
}

func HandleSignalMetricsSync(ctx context.Context, t *asynq.Task) error {
	_ = t
	if err := syncSignalMetrics(ctx); err != nil {
		return fmt.Errorf("sync signal metrics: %w", err)
	}
	return nil
}
