package task

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type SignalGDPRCleanupPayload struct{}

func NewSignalGDPRCleanupTask() (*asynq.Task, error) {
	return asynq.NewTask(TypeSignalGDPRCleanup, []byte(`{}`), asynq.Queue("low")), nil
}

func HandleSignalGDPRCleanup(ctx context.Context, t *asynq.Task) error {
	_ = t
	if err := runSignalGDPRCleanup(ctx); err != nil {
		return fmt.Errorf("run signal gdpr cleanup: %w", err)
	}
	return nil
}
