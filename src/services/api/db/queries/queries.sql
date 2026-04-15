-- name: GetWorkspaceByOrgID :one
SELECT * FROM workspaces
WHERE org_id = $1
LIMIT 1;

-- name: UpsertWorkspace :one
INSERT INTO workspaces (org_id, plan_tier, sub_status, trial_ends_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (org_id) DO UPDATE SET
    plan_tier     = EXCLUDED.plan_tier,
    sub_status    = EXCLUDED.sub_status,
    trial_ends_at = EXCLUDED.trial_ends_at,
    updated_at    = NOW()
RETURNING *;

-- name: CreateJob :one
INSERT INTO jobs (workspace_id, product_url, model)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetJobByID :one
SELECT * FROM jobs
WHERE id = $1 AND workspace_id = $2
LIMIT 1;

-- name: ListJobsByWorkspace :many
SELECT * FROM jobs
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateVariant :one
INSERT INTO variants (job_id, workspace_id, angle)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateVariantComplete :one
UPDATE variants
SET status = 'complete',
    mux_asset_id    = $2,
    mux_playback_id = $3,
    r2_key          = $4,
    duration_secs   = $5,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: ListVariantsByJob :many
SELECT * FROM variants
WHERE job_id = $1
ORDER BY created_at ASC;
