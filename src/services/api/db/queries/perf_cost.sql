-- name: InsertPerfEvent :exec
INSERT INTO video_performance_events (
  workspace_id, variant_id, job_id,
  stage, duration_ms, model,
  fal_request_id, error_type, error_msg, recorded_at
) VALUES (
  $1, $2, $3,
  $4, $5, $6,
  $7, $8, $9, $10
);

-- name: InsertCostEvent :exec
INSERT INTO cost_events (
  workspace_id, variant_id, job_id,
  source, model,
  estimated_usd, credits, recorded_at
) VALUES (
  $1, $2, $3,
  $4, $5,
  $6, $7, $8
);

-- name: GetWorkspaceMonthCost :one
SELECT COALESCE(SUM(estimated_usd), 0)::NUMERIC AS total_usd
FROM cost_events
WHERE workspace_id = $1
  AND date_trunc('month', recorded_at) = date_trunc('month', NOW());

-- name: ListPerfEventsByVariant :many
SELECT stage, duration_ms, model, fal_request_id, error_type, recorded_at
FROM video_performance_events
WHERE variant_id = $1
ORDER BY recorded_at ASC;

-- name: UpdateVariantAvatarJob :exec
UPDATE variants
SET avatar_job_id    = $2,
    avatar_provider  = $3,
    updated_at       = NOW()
WHERE id = $1;

-- name: CreateJobIdempotent :one
-- ON CONFLICT returns the existing row — client gets the same job_id on retry.
INSERT INTO jobs (
  workspace_id, product_url, status, model, idempotency_key
) VALUES (
  $1, $2, 'queued', $3, $4
)
ON CONFLICT (workspace_id, idempotency_key)
WHERE idempotency_key IS NOT NULL
DO UPDATE SET updated_at = NOW()
RETURNING *;
