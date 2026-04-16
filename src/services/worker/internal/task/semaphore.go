package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrSemaphoreFull is returned when the per-org FAL.AI concurrency limit is reached.
var ErrSemaphoreFull = errors.New("fal concurrency limit reached for this workspace")

const (
	// falMaxConcurrent is FAL.AI's hard per-user concurrency limit.
	falMaxConcurrent = 2

	// falSemaphoreTTL is how long a semaphore slot is held before auto-release.
	// Set longer than the slowest FAL generation (Veo 3.1 = ~150s).
	falSemaphoreTTL = 20 * time.Minute

	// costHourlyKeyTTL keeps hourly cost keys alive for 2 hours.
	costHourlyKeyTTL = 2 * time.Hour
)

// FalSemaphore manages per-workspace concurrency against FAL.AI.
// FAL.AI enforces a hard limit of 2 concurrent requests per API key per user.
// Without this guard, concurrent asynq workers silently hit 429s and retry,
// burning credits and creating cascading failures.
type FalSemaphore struct {
	rdb *redis.Client
}

// NewFalSemaphore creates a semaphore backed by the given Redis client.
func NewFalSemaphore(rdb *redis.Client) *FalSemaphore {
	return &FalSemaphore{rdb: rdb}
}

// Acquire attempts to claim one FAL.AI concurrency slot for the workspace.
// Returns ErrSemaphoreFull if the workspace is at capacity.
// Returns the slot key so the caller can release it explicitly.
func (s *FalSemaphore) Acquire(ctx context.Context, workspaceID, variantID string) (slotKey string, err error) {
	for slot := 1; slot <= falMaxConcurrent; slot++ {
		key := falSlotKey(workspaceID, slot)
		// SETNX: set only if not exists → atomic claim
		set, err := s.rdb.SetNX(ctx, key, variantID, falSemaphoreTTL).Result()
		if err != nil {
			return "", fmt.Errorf("semaphore acquire redis: %w", err)
		}
		if set {
			return key, nil // slot claimed
		}
	}
	return "", ErrSemaphoreFull
}

// Release frees the semaphore slot after generation completes or fails.
// Safe to call multiple times (idempotent via DEL).
func (s *FalSemaphore) Release(ctx context.Context, slotKey string) error {
	if slotKey == "" {
		return nil
	}
	return s.rdb.Del(ctx, slotKey).Err()
}

// ReleaseByVariant scans all slots for a workspace and releases the one
// held by variantID. Used in webhook handler when slot key is not available.
func (s *FalSemaphore) ReleaseByVariant(ctx context.Context, workspaceID, variantID string) error {
	for slot := 1; slot <= falMaxConcurrent; slot++ {
		key := falSlotKey(workspaceID, slot)
		val, err := s.rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return fmt.Errorf("semaphore release scan: %w", err)
		}
		if val == variantID {
			return s.rdb.Del(ctx, key).Err()
		}
	}
	return nil // slot already released or not found — safe
}

func falSlotKey(workspaceID string, slot int) string {
	return fmt.Sprintf("fal:sem:%s:%d", workspaceID, slot)
}

// =============================================================================
// Cost Circuit Breaker
// Tracks estimated hourly FAL spend per workspace.
// Workers check this before every FAL submission.
// =============================================================================

// CostLimits defines per-tier hourly FAL spend caps in USD.
var CostLimits = map[string]float64{
	"starter": 2.00,
	"growth":  8.00,
	"agency":  20.00,
}

// ModelCost is the estimated USD cost per FAL generation by model.
var ModelCost = map[string]float64{
	"veo-3.1":     0.50,
	"kling-3.0":   0.25,
	"runway-gen4": 0.20,
	"sora-2":      0.40,
}

// CostCircuitBreaker tracks per-workspace hourly spend in Redis.
type CostCircuitBreaker struct {
	rdb *redis.Client
}

// NewCostCircuitBreaker creates a circuit breaker backed by Redis.
func NewCostCircuitBreaker(rdb *redis.Client) *CostCircuitBreaker {
	return &CostCircuitBreaker{rdb: rdb}
}

// CheckAndIncrement checks whether the workspace is within its hourly cost limit,
// then atomically increments the counter if allowed.
// Returns an error if the limit would be exceeded.
func (cb *CostCircuitBreaker) CheckAndIncrement(ctx context.Context, workspaceID, planTier, model string) error {
	limit, ok := CostLimits[planTier]
	if !ok {
		limit = CostLimits["starter"]
	}
	cost, ok := ModelCost[model]
	if !ok {
		cost = 0.50 // conservative default
	}

	key := costHourlyKey(workspaceID)

	// INCRBYFLOAT is atomic — no WATCH/MULTI needed
	newVal, err := cb.rdb.IncrByFloat(ctx, key, cost).Result()
	if err != nil {
		// Fail open: if Redis is unavailable, allow the request
		return nil
	}

	// Set TTL on first increment
	if newVal == cost {
		cb.rdb.Expire(ctx, key, costHourlyKeyTTL)
	}

	if newVal > limit {
		// Roll back the increment — we won't be making this call
		cb.rdb.IncrByFloat(ctx, key, -cost)
		return fmt.Errorf("hourly cost limit $%.2f exceeded for workspace %s (plan: %s)",
			limit, workspaceID, planTier)
	}
	return nil
}

// CurrentHourlyCost returns the current estimated spend for the workspace this hour.
func (cb *CostCircuitBreaker) CurrentHourlyCost(ctx context.Context, workspaceID string) float64 {
	val, err := cb.rdb.Get(ctx, costHourlyKey(workspaceID)).Float64()
	if err != nil {
		return 0
	}
	return val
}

func costHourlyKey(workspaceID string) string {
	now := time.Now().UTC()
	return fmt.Sprintf("fal:cost:%s:%s:%02d", workspaceID, now.Format("20060102"), now.Hour())
}
