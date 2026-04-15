package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type SignalRecommendationsRefreshPayload struct {
	Days int `json:"days"`
}

func NewSignalRecommendationsRefreshTask(days int) (*asynq.Task, error) {
	if days <= 0 {
		days = 90
	}
	data, err := json.Marshal(SignalRecommendationsRefreshPayload{Days: days})
	if err != nil {
		return nil, fmt.Errorf("marshal signal recommendations payload: %w", err)
	}
	return asynq.NewTask(TypeSignalRecommendationsRefresh, data, asynq.Queue("low")), nil
}

func HandleSignalRecommendationsRefresh(ctx context.Context, t *asynq.Task) error {
	var payload SignalRecommendationsRefreshPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal signal recommendations payload: %w", err)
	}
	if payload.Days <= 0 {
		payload.Days = 90
	}
	if err := refreshAllSignalRecommendations(ctx, payload.Days); err != nil {
		return fmt.Errorf("refresh signal recommendations: %w", err)
	}
	return nil
}
