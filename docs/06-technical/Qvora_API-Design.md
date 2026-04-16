# Qvora API Design

**Version:** 2.0  
**Date:** April 16, 2026  
**Status:** Transitional (V1 runtime active, Phase 8+ microservice target)  
**Base URL:** `https://api.qvora.com/v1`  
**Internal BFF:** `http://localhost:3001/trpc` (tRPC, Next.js → Go bridge)  
**Auth:** Bearer token (Clerk JWT) in `Authorization` header
**Canonical architecture:** `docs/06-technical/Qvora_Microservice-Architecture.md`

---

## Table of Contents

1. [Conventions](#conventions)
2. [Authentication & Authorization](#authentication--authorization)
3. [Error Codes](#error-codes)
4. [Rate Limits](#rate-limits)
5. [Endpoints](#endpoints)
   - [Briefs](#briefs)
   - [Assets](#assets)
   - [Generation Jobs](#generation-jobs)
   - [Exports](#exports)
   - [Projects](#projects)
   - [Brands](#brands)
   - [Org / Billing](#org--billing)
   - [Webhooks (FAL.AI Callbacks)](#webhooks-falai-callbacks)
  - [Realtime Status Channels](#realtime-status-channels)
6. [tRPC BFF Procedures](#trpc-bff-procedures)
7. [Data Models](#data-models)

---

## Conventions

### Versioning
All public endpoints are prefixed with `/v1`. Breaking changes increment the major version. Non-breaking additions are additive within the same version.

### Request Format
- Content-Type: `application/json`
- Dates: ISO 8601 (`2026-04-14T10:30:00Z`)
- UUIDs: lowercase hyphenated (`550e8400-e29b-41d4-a716-446655440000`)
- Pagination: cursor-based via `cursor` + `limit` (default 20, max 100)

### Response Envelope
All responses follow a consistent envelope:

```json
{
  "data": { ... },
  "meta": {
    "request_id": "req_01HZ...",
    "timestamp": "2026-04-14T10:30:00Z"
  }
}
```

List responses include pagination:
```json
{
  "data": [ ... ],
  "pagination": {
    "cursor": "eyJpZCI6IjEyMyJ9",
    "has_more": true,
    "total": 147
  },
  "meta": { ... }
}
```

### Idempotency
POST requests that create resources require `X-Idempotency-Key` on job/brief creation endpoints. Duplicate requests with the same key return the original response (deduped by workspace + key).

---

## Authentication & Authorization

All API requests require a valid Clerk JWT in the `Authorization` header:

```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

The JWT payload must include:
```json
{
  "sub": "user_01HZ...",
  "org_id": "org_01HZ...",
  "org_role": "admin | member",
  "plan": "starter | growth | agency"
}
```

**Scopes enforced per endpoint:**

| Scope | Description |
|---|---|
| `read:briefs` | View briefs in org |
| `write:briefs` | Create/update/delete briefs |
| `read:assets` | View generated assets |
| `write:assets` | Trigger generation, update metadata |
| `read:exports` | Download export packages |
| `write:exports` | Create export jobs |
| `admin:org` | Manage org settings, billing, members |

**Row-Level Security**: All queries are scoped to `org_id` extracted from JWT. Cross-org access returns `403`.

---

## Error Codes

```json
{
  "error": {
    "code": "BRIEF_NOT_FOUND",
    "message": "Brief with id '550e...' not found or not accessible",
    "status": 404,
    "request_id": "req_01HZ..."
  }
}
```

| HTTP Status | Error Code | Description |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Request body failed schema validation |
| 400 | `INVALID_CURSOR` | Pagination cursor is malformed or expired |
| 400 | `QUOTA_EXCEEDED` | Monthly generation quota reached for plan |
| 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| 401 | `TOKEN_EXPIRED` | JWT is expired |
| 403 | `FORBIDDEN` | JWT valid but insufficient permissions |
| 403 | `PLAN_REQUIRED` | Feature requires higher plan tier |
| 404 | `BRIEF_NOT_FOUND` | Brief does not exist or not in org |
| 404 | `ASSET_NOT_FOUND` | Asset not found |
| 404 | `JOB_NOT_FOUND` | Generation job not found |
| 409 | `DUPLICATE_IDEMPOTENCY_KEY` | Duplicate request (returns original response) |
| 422 | `URL_UNREACHABLE` | Provided URL could not be fetched |
| 422 | `SCRAPE_FAILED` | URL fetched but content extraction failed |
| 422 | `GENERATION_FAILED` | FAL.AI / HeyGen job returned error |
| 429 | `RATE_LIMITED` | Too many requests, see `Retry-After` header |
| 500 | `INTERNAL_ERROR` | Unexpected server error |
| 503 | `SERVICE_UNAVAILABLE` | Upstream AI service unavailable |

---

## Rate Limits

| Plan | Requests/min | Generation jobs/day | Concurrent jobs |
|---|---|---|---|
| Starter | 60 | 20 | 2 |
| Growth | 120 | 50 | 5 |
| Agency | 300 | 200 | 15 |

Rate limit headers on every response:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1713091800
```

---

## Endpoints

---

### Briefs

Briefs are the central resource — AI-generated creative strategy documents derived from a URL.

---

#### `POST /briefs`

Extract URL and generate a creative brief via URL extraction + structured LLM pipeline.

**Request:**
```json
{
  "url": "https://brand.com/product-page",
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "options": {
    "tone": "bold | conversational | professional | playful",
    "audience_override": "string (optional, overrides extracted audience)",
    "angle_count": 3
  }
}
```

**Response `202 Accepted`:**
```json
{
  "data": {
    "id": "brief_01HZ...",
    "status": "processing",
    "job_id": "job_01HZ...",
    "status_channel": "generation_jobs:job_01HZ...",
    "estimated_seconds": 25
  }
}
```

**Notes:**
- Brief generation is async. Poll `GET /briefs/{id}` for V1.
- Realtime status updates move to Supabase Realtime in Phase 8. SSE is a V1 compatibility path.
- `options.angle_count` defaults to 3, max 5 (Agency plan only).

---

#### `GET /briefs`

List briefs for the authenticated org.

**Query Parameters:**

| Param | Type | Default | Description |
|---|---|---|---|
| `project_id` | UUID | — | Filter by project |
| `status` | string | — | `processing\|ready\|failed` |
| `cursor` | string | — | Pagination cursor |
| `limit` | int | 20 | Max 100 |
| `sort` | string | `created_at:desc` | `created_at:asc\|desc` |

**Response `200 OK`:**
```json
{
  "data": [
    {
      "id": "brief_01HZ...",
      "project_id": "550e...",
      "url": "https://brand.com/product-page",
      "brand_name": "Acme Co",
      "status": "ready",
      "angles_count": 3,
      "created_at": "2026-04-14T10:30:00Z",
      "updated_at": "2026-04-14T10:30:45Z"
    }
  ],
  "pagination": {
    "cursor": "eyJpZCI6ImJyaWVmXzAxSFoifQ==",
    "has_more": false,
    "total": 12
  }
}
```

---

#### `GET /briefs/{id}`

Fetch a single brief with full content.

**Response `200 OK`:**
```json
{
  "data": {
    "id": "brief_01HZ...",
    "project_id": "550e...",
    "url": "https://brand.com/product-page",
    "brand_name": "Acme Co",
    "status": "ready",
    "content": {
      "product_summary": "string",
      "target_audience": "string",
      "pain_points": ["string"],
      "usp": "string",
      "tone": "bold",
      "angles": [
        {
          "id": "angle_01",
          "type": "problem_solution | social_proof | urgency | lifestyle | feature_focus",
          "headline": "string",
          "hook": "string",
          "body": "string",
          "cta": "string",
          "recommended_format": "vertical_9x16 | square_1x1 | horizontal_16x9",
          "recommended_duration": 15
        }
      ],
      "visual_direction": "string",
      "color_palette": ["#hex"],
      "competitor_contrast": "string"
    },
    "raw_extraction": {
      "title": "string",
      "meta_description": "string",
      "og_image": "https://...",
      "extracted_text_chars": 2847
    },
    "created_at": "2026-04-14T10:30:00Z",
    "updated_at": "2026-04-14T10:30:45Z"
  }
}
```

---

#### `PATCH /briefs/{id}`

Update editable fields on a brief (human overrides on AI-generated content).

**Request:**
```json
{
  "content": {
    "angles": [
      {
        "id": "angle_01",
        "headline": "Updated headline",
        "hook": "Updated hook"
      }
    ]
  }
}
```

**Response `200 OK`:** Returns updated brief object (same as GET).

**Notes:** Only `content.angles[*].headline`, `hook`, `body`, `cta` are patchable. Structural fields (`type`, `recommended_format`) are locked post-generation.

---

#### `DELETE /briefs/{id}`

Soft-delete a brief (sets `deleted_at`, cascades to asset `brief_id` references).

**Response `204 No Content`**

---

#### `POST /briefs/{id}/regenerate`

Trigger re-generation of the brief using the same URL and options.

**Request:**
```json
{
  "reason": "too_generic | wrong_tone | bad_angles | other"
}
```

**Response `202 Accepted`:** Same as `POST /briefs`.

---

### Assets

Assets are generated video/image files attached to a brief angle.

---

#### `POST /assets/generate`

Queue a video generation job from a brief angle or direct prompt.

**Request (from brief):**
```json
{
  "brief_id": "brief_01HZ...",
  "angle_id": "angle_01",
  "generation_mode": "t2v | i2v | v2v",
  "format": "vertical_9x16 | square_1x1 | horizontal_16x9",
  "duration": 15,
  "t2v_options": {
    "model": "auto | kling_3 | veo_3 | runway_gen4 | sora_2",
    "style": "cinematic | ugc | product_demo | documentary | animated",
    "camera_motion": "static | slow_zoom | pan_left | pan_right | orbit",
    "negative_prompt": "string (optional)"
  },
  "i2v_options": {
    "source_image_url": "https://r2.qvora.com/...",
    "motion_prompt": "string",
    "motion_intensity": 0.7,
    "model": "auto | kling_3 | runway_gen4 | svd"
  },
  "v2v_options": {
    "voice_source": "upload | record | clone",
    "voice_asset_id": "asset_01HZ...",
    "avatar_id": "heygen_avatar_id",
    "language": "en | es | fr | de | pt | ja | ko | zh",
    "voice_clone_consent": true
  },
  "brand_slug": "acme-co"
}
```

**Response `202 Accepted`:**
```json
{
  "data": {
    "id": "asset_01HZ...",
    "job_id": "job_01HZ...",
    "status": "queued",
    "status_channel": "generation_jobs:job_01HZ...",
    "estimated_seconds": 120
  }
}
```

---

#### `GET /assets`

List assets for the org.

**Query Parameters:**

| Param | Type | Default | Description |
|---|---|---|---|
| `brief_id` | UUID | — | Filter by brief |
| `project_id` | UUID | — | Filter by project |
| `status` | string | — | `queued\|processing\|ready\|failed` |
| `generation_mode` | string | — | `t2v\|i2v\|v2v` |
| `format` | string | — | `vertical_9x16\|square_1x1\|horizontal_16x9` |
| `cursor` | string | — | Pagination cursor |
| `limit` | int | 20 | Max 100 |

**Response `200 OK`:**
```json
{
  "data": [
    {
      "id": "asset_01HZ...",
      "brief_id": "brief_01HZ...",
      "angle_id": "angle_01",
      "status": "ready",
      "generation_mode": "t2v",
      "format": "vertical_9x16",
      "duration_seconds": 15,
      "file_name": "acme-co_problem_solution_vertical_9x16_hook_a_20260414_v1.mp4",
      "storage_url": "https://r2.qvora.com/...",
      "thumbnail_url": "https://r2.qvora.com/...",
      "play_url": "https://stream.mux.com/...",
      "file_size_bytes": 18500000,
      "model_used": "kling_3",
      "generation_cost_usd": 0.12,
      "created_at": "2026-04-14T10:30:00Z"
    }
  ]
}
```

---

#### `GET /assets/{id}`

Fetch single asset with full metadata.

**Response `200 OK`:**
```json
{
  "data": {
    "id": "asset_01HZ...",
    "brief_id": "brief_01HZ...",
    "angle_id": "angle_01",
    "angle_type": "problem_solution",
    "hook_variant": "hook_a",
    "status": "ready",
    "generation_mode": "t2v",
    "format": "vertical_9x16",
    "duration_seconds": 15,
    "model_used": "kling_3",
    "prompt_used": "string (the final prompt sent to FAL.AI)",
    "file_name": "acme-co_problem_solution_vertical_9x16_hook_a_20260414_v1.mp4",
    "storage_url": "https://r2.qvora.com/...",
    "thumbnail_url": "https://r2.qvora.com/...",
    "play_url": "https://stream.mux.com/...",
    "file_size_bytes": 18500000,
    "generation_cost_usd": 0.12,
    "tags": {
      "brand_slug": "acme-co",
      "angle_type": "problem_solution",
      "format": "vertical_9x16",
      "hook_variant": "hook_a"
    },
    "created_at": "2026-04-14T10:30:00Z",
    "updated_at": "2026-04-14T10:30:00Z"
  }
}
```

---

#### `PATCH /assets/{id}`

Update asset tags and human-readable labels.

**Request:**
```json
{
  "tags": {
    "hook_variant": "hook_b"
  },
  "notes": "Approved for Meta campaign Q2"
}
```

**Response `200 OK`:** Returns updated asset object.

---

#### `DELETE /assets/{id}`

Soft-delete asset and schedule R2 storage cleanup (24h delay for safety).

**Response `204 No Content`**

---

#### `POST /assets/{id}/retry`

Re-queue a failed generation job with the same parameters.

**Response `202 Accepted`:** Same envelope as `POST /assets/generate`.

---

### Generation Jobs

Jobs track async AI generation progress across FAL.AI and HeyGen.

---

#### `GET /jobs/{id}`

Poll job status.

**Response `200 OK`:**
```json
{
  "data": {
    "id": "job_01HZ...",
    "type": "brief_generation | video_generation | export_packaging",
    "status": "queued | processing | completed | failed",
    "resource_id": "asset_01HZ...",
    "resource_type": "brief | asset | export",
    "progress": {
      "percent": 65,
      "stage": "rendering",
      "stages_completed": ["queued", "prompt_built", "submitted_to_falai"],
      "current_stage": "rendering",
      "stages_remaining": ["post_processing", "upload", "thumbnail"]
    },
    "external_job_id": "fal_task_abc123",
    "started_at": "2026-04-14T10:30:00Z",
    "completed_at": null,
    "error": null,
    "created_at": "2026-04-14T10:29:50Z"
  }
}
```

---

#### `GET /jobs`

List recent jobs for org.

**Query Parameters:**

| Param | Type | Description |
|---|---|---|
| `type` | string | `brief_generation\|video_generation\|export_packaging` |
| `status` | string | `queued\|processing\|completed\|failed` |
| `resource_id` | UUID | Filter by resource |
| `limit` | int | Default 20 |

---

### Exports

Exports package selected assets into structured ZIP files for ad platform upload.

---

#### `POST /exports`

Create an export package.

**Request:**
```json
{
  "name": "Meta Q2 Campaign Pack",
  "asset_ids": ["asset_01HZ...", "asset_02HZ..."],
  "options": {
    "include_brief_pdf": true,
    "include_captions": true,
    "naming_convention": "standard | custom",
    "custom_prefix": "BRAND_Q2_"
  }
}
```

**Response `202 Accepted`:**
```json
{
  "data": {
    "id": "export_01HZ...",
    "status": "packaging",
    "job_id": "job_01HZ...",
    "estimated_seconds": 30
  }
}
```

---

#### `GET /exports/{id}`

Get export status and download URL when ready.

**Response `200 OK`:**
```json
{
  "data": {
    "id": "export_01HZ...",
    "name": "Meta Q2 Campaign Pack",
    "status": "ready | packaging | failed",
    "asset_count": 6,
    "file_size_bytes": 112000000,
    "download_url": "https://r2.qvora.com/exports/export_01HZ....zip?X-Amz-Expires=3600&...",
    "download_url_expires_at": "2026-04-14T11:30:00Z",
    "manifest": {
      "assets": [
        {
          "file_name": "acme-co_problem_solution_vertical_9x16_hook_a_20260414_v1.mp4",
          "format": "vertical_9x16",
          "angle_type": "problem_solution",
          "duration_seconds": 15
        }
      ]
    },
    "created_at": "2026-04-14T10:30:00Z",
    "completed_at": "2026-04-14T10:30:35Z"
  }
}
```

---

#### `GET /exports`

List exports for org.

**Query Parameters:** `project_id`, `status`, `cursor`, `limit`

---

#### `DELETE /exports/{id}`

Delete export package and revoke download URL.

**Response `204 No Content`**

---

### Projects

Projects group briefs and assets by campaign or client.

---

#### `POST /projects`

**Request:**
```json
{
  "name": "Spring Campaign 2026",
  "description": "string (optional)",
  "color": "#7B2FFF"
}
```

**Response `201 Created`:**
```json
{
  "data": {
    "id": "proj_01HZ...",
    "org_id": "org_01HZ...",
    "name": "Spring Campaign 2026",
    "description": "string",
    "color": "#7B2FFF",
    "briefs_count": 0,
    "assets_count": 0,
    "created_at": "2026-04-14T10:30:00Z"
  }
}
```

---

#### `GET /projects`

List all projects for org. No pagination needed (max 50 projects expected).

**Response `200 OK`:**
```json
{
  "data": [ { ... } ]
}
```

---

#### `GET /projects/{id}`

Fetch project with summary stats.

---

#### `PATCH /projects/{id}`

Update `name`, `description`, `color`.

---

#### `DELETE /projects/{id}`

Soft-delete project. Briefs and assets are retained but `project_id` is nullified.

---

### Brands

Brand profiles store extracted brand identity for reuse across briefs.

---

#### `POST /brands`

Manually create or save a brand profile (auto-created from brief extraction).

**Request:**
```json
{
  "name": "Acme Co",
  "slug": "acme-co",
  "website": "https://acme.com",
  "logo_url": "https://...",
  "color_palette": ["#FF5733", "#2ECC71"],
  "tone": "bold",
  "notes": "string (optional)"
}
```

---

#### `GET /brands`

List saved brand profiles for org.

---

#### `GET /brands/{id}`

---

#### `PATCH /brands/{id}`

---

#### `DELETE /brands/{id}`

---

### Org / Billing

---

#### `GET /org`

Get current org profile and usage stats.

**Response `200 OK`:**
```json
{
  "data": {
    "id": "org_01HZ...",
    "name": "Acme Agency",
    "plan": "growth",
    "usage": {
      "period_start": "2026-04-01T00:00:00Z",
      "period_end": "2026-04-30T23:59:59Z",
      "ads_used": 18,
      "ads_limit": 50,
      "briefs_generated": 24,
      "exports_created": 7,
      "storage_used_bytes": 2400000000,
      "storage_limit_bytes": 10737418240
    },
    "seats_used": 3,
    "seats_limit": 5
  }
}
```

---

#### `GET /org/members`

List org members (admin only).

**Response `200 OK`:**
```json
{
  "data": [
    {
      "id": "member_01HZ...",
      "user_id": "user_01HZ...",
      "email": "jane@acme.com",
      "name": "Jane Smith",
      "role": "admin | member",
      "joined_at": "2026-03-01T00:00:00Z"
    }
  ]
}
```

---

#### `POST /org/members/invite`

Invite a new member (admin only).

**Request:**
```json
{
  "email": "newmember@acme.com",
  "role": "member"
}
```

**Response `201 Created`:**
```json
{
  "data": {
    "invite_id": "inv_01HZ...",
    "email": "newmember@acme.com",
    "expires_at": "2026-04-21T10:30:00Z"
  }
}
```

---

#### `DELETE /org/members/{member_id}`

Remove member from org (admin only).

---

#### `GET /org/billing`

Get billing info and upcoming invoice (admin only).

**Response `200 OK`:**
```json
{
  "data": {
    "plan": "growth",
    "status": "active",
    "current_period_end": "2026-05-14T00:00:00Z",
    "amount_due_usd": 149.00,
    "payment_method": {
      "type": "card",
      "last4": "4242",
      "brand": "visa"
    },
    "portal_url": "https://billing.stripe.com/session/..."
  }
}
```

---

### Webhooks (FAL.AI Callbacks)

Internal endpoints called by FAL.AI. Not exposed to API consumers. Protected by HMAC signature verification.

---

#### `POST /internal/webhooks/falai`

Receives FAL.AI job completion callbacks.

**Headers:**
```
X-Fal-Signature: sha256=abc123...
Content-Type: application/json
```

**Request body (FAL.AI format):**
```json
{
  "request_id": "fal_task_abc123",
  "status": "COMPLETED | FAILED",
  "payload": {
    "video": {
      "url": "https://fal.media/files/...",
      "content_type": "video/mp4",
      "file_name": "output.mp4",
      "file_size": 18500000
    }
  },
  "error": null
}
```

**Processing flow:**
1. Verify `X-Fal-Signature` HMAC
2. Lookup `generation_jobs` by `external_job_id = request_id`
3. If COMPLETED: download from FAL.AI URL → upload to R2 → generate thumbnail → update `assets.status = ready` → push SSE event
4. If FAILED: update `assets.status = failed`, `generation_jobs.status = failed` → push SSE error event

V1 path: status updates may be delivered via SSE compatibility endpoints.
Phase 8+ target: status updates are delivered via Supabase Realtime channels.

**Response `200 OK`:** `{ "received": true }`

---

#### `POST /internal/webhooks/heygen`

Receives HeyGen Avatar v3 job completion callbacks. Same signature verification pattern.

---

### Realtime Status Channels

Primary target architecture is Supabase Realtime (Phase 8+). Legacy SSE endpoints remain documented for V1 compatibility.

---

#### `GET /stream/briefs/{brief_id}`

Stream brief generation progress.

**Headers:**
```
Accept: text/event-stream
Authorization: Bearer ...
```

**Event stream:**
```
event: progress
data: {"stage": "extracting", "percent": 10, "message": "Fetching URL content..."}

event: progress
data: {"stage": "parsing", "percent": 30, "message": "Parsing brand signals..."}

event: progress
data: {"stage": "generating_angles", "percent": 60, "message": "Writing creative angles..."}

event: progress
data: {"stage": "validating", "percent": 85, "message": "Validating output quality..."}

event: complete
data: {"brief_id": "brief_01HZ...", "status": "ready"}

event: error
data: {"code": "SCRAPE_FAILED", "message": "Could not extract content from URL"}
```

---

#### `GET /stream/assets/{asset_id}`

Stream video generation progress.

```
event: progress
data: {"stage": "queued", "percent": 5, "message": "Job queued..."}

event: progress
data: {"stage": "prompt_built", "percent": 15, "message": "Prompt assembled..."}

event: progress
data: {"stage": "submitted", "percent": 20, "message": "Submitted to Kling 3.0..."}

event: progress
data: {"stage": "rendering", "percent": 65, "message": "Model rendering video..."}

event: progress
data: {"stage": "post_processing", "percent": 85, "message": "Processing and uploading..."}

event: complete
data: {"asset_id": "asset_01HZ...", "play_url": "https://stream.mux.com/...", "thumbnail_url": "..."}

event: error
data: {"code": "GENERATION_FAILED", "message": "Model returned error: content policy violation"}
```

---

## tRPC BFF Procedures

The tRPC BFF layer lives in Next.js and proxies authenticated calls to the Go API. It provides type-safety end-to-end and handles:
- Clerk session → JWT extraction
- Request deduplication for concurrent React Query calls
- Optimistic updates (brief list, asset list)

### Key Procedures

```typescript
// briefs router
trpc.briefs.create(input: CreateBriefInput): Promise<BriefJob>
trpc.briefs.list(input: ListBriefsInput): Promise<PaginatedBriefs>
trpc.briefs.get(input: { id: string }): Promise<Brief>
trpc.briefs.patch(input: PatchBriefInput): Promise<Brief>
trpc.briefs.delete(input: { id: string }): Promise<void>
trpc.briefs.regenerate(input: { id: string; reason: string }): Promise<BriefJob>

// assets router
trpc.assets.generate(input: GenerateAssetInput): Promise<AssetJob>
trpc.assets.list(input: ListAssetsInput): Promise<PaginatedAssets>
trpc.assets.get(input: { id: string }): Promise<Asset>
trpc.assets.patch(input: PatchAssetInput): Promise<Asset>
trpc.assets.delete(input: { id: string }): Promise<void>
trpc.assets.retry(input: { id: string }): Promise<AssetJob>

// exports router
trpc.exports.create(input: CreateExportInput): Promise<ExportJob>
trpc.exports.get(input: { id: string }): Promise<Export>
trpc.exports.list(input: ListExportsInput): Promise<PaginatedExports>
trpc.exports.delete(input: { id: string }): Promise<void>

// projects router
trpc.projects.create(input: CreateProjectInput): Promise<Project>
trpc.projects.list(): Promise<Project[]>
trpc.projects.get(input: { id: string }): Promise<Project>
trpc.projects.patch(input: PatchProjectInput): Promise<Project>
trpc.projects.delete(input: { id: string }): Promise<void>

// org router (admin-gated)
trpc.org.get(): Promise<OrgProfile>
trpc.org.members.list(): Promise<Member[]>
trpc.org.members.invite(input: InviteMemberInput): Promise<Invite>
trpc.org.members.remove(input: { memberId: string }): Promise<void>
trpc.org.billing.get(): Promise<BillingInfo>
```

### Input Schemas (Zod)

```typescript
const CreateBriefInput = z.object({
  url: z.string().url(),
  project_id: z.string().uuid().optional(),
  options: z.object({
    tone: z.enum(["bold", "conversational", "professional", "playful"]).optional(),
    audience_override: z.string().max(200).optional(),
    angle_count: z.number().int().min(1).max(5).default(3),
  }).optional(),
})

const GenerateAssetInput = z.object({
  brief_id: z.string().uuid(),
  angle_id: z.string(),
  generation_mode: z.enum(["t2v", "i2v", "v2v"]),
  format: z.enum(["vertical_9x16", "square_1x1", "horizontal_16x9"]),
  duration: z.number().int().min(6).max(60).default(15),
  t2v_options: z.object({
    model: z.enum(["auto", "kling_3", "veo_3", "runway_gen4", "sora_2"]).default("auto"),
    style: z.enum(["cinematic", "ugc", "product_demo", "documentary", "animated"]).optional(),
    camera_motion: z.enum(["static", "slow_zoom", "pan_left", "pan_right", "orbit"]).optional(),
    negative_prompt: z.string().max(500).optional(),
  }).optional(),
  i2v_options: z.object({
    source_image_url: z.string().url(),
    motion_prompt: z.string().max(500).optional(),
    motion_intensity: z.number().min(0).max(1).default(0.7),
    model: z.enum(["auto", "kling_3", "runway_gen4", "svd"]).default("auto"),
  }).optional(),
  v2v_options: z.object({
    voice_source: z.enum(["upload", "record", "clone"]),
    voice_asset_id: z.string().uuid().optional(),
    avatar_id: z.string(),
    language: z.enum(["en", "es", "fr", "de", "pt", "ja", "ko", "zh"]).default("en"),
    voice_clone_consent: z.boolean(),
  }).optional(),
  brand_slug: z.string().max(50),
})
```

---

## Data Models

### TypeScript types (generated from Zod, used across frontend and BFF)

```typescript
type Brief = {
  id: string
  project_id: string | null
  url: string
  brand_name: string
  status: "processing" | "ready" | "failed"
  content: BriefContent | null
  raw_extraction: RawExtraction | null
  created_at: string
  updated_at: string
}

type BriefContent = {
  product_summary: string
  target_audience: string
  pain_points: string[]
  usp: string
  tone: "bold" | "conversational" | "professional" | "playful"
  angles: BriefAngle[]
  visual_direction: string
  color_palette: string[]
  competitor_contrast: string
}

type BriefAngle = {
  id: string
  type: "problem_solution" | "social_proof" | "urgency" | "lifestyle" | "feature_focus"
  headline: string
  hook: string
  body: string
  cta: string
  recommended_format: "vertical_9x16" | "square_1x1" | "horizontal_16x9"
  recommended_duration: number
}

type Asset = {
  id: string
  brief_id: string
  angle_id: string
  angle_type: string
  hook_variant: string
  status: "queued" | "processing" | "ready" | "failed"
  generation_mode: "t2v" | "i2v" | "v2v"
  format: "vertical_9x16" | "square_1x1" | "horizontal_16x9"
  duration_seconds: number
  model_used: string | null
  prompt_used: string | null
  file_name: string | null
  storage_url: string | null
  thumbnail_url: string | null
  play_url: string | null
  file_size_bytes: number | null
  generation_cost_usd: number | null
  tags: AssetTags
  created_at: string
  updated_at: string
}

type AssetTags = {
  brand_slug: string
  angle_type: string
  format: string
  hook_variant: string
}

type GenerationJob = {
  id: string
  type: "brief_generation" | "video_generation" | "export_packaging"
  status: "queued" | "processing" | "completed" | "failed"
  resource_id: string
  resource_type: "brief" | "asset" | "export"
  progress: JobProgress
  external_job_id: string | null
  started_at: string | null
  completed_at: string | null
  error: string | null
  created_at: string
}

type JobProgress = {
  percent: number
  stage: string
  stages_completed: string[]
  current_stage: string
  stages_remaining: string[]
}
```

---

## Endpoint Summary

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/briefs` | JWT | Create brief from URL |
| GET | `/briefs` | JWT | List org briefs |
| GET | `/briefs/{id}` | JWT | Get brief detail |
| PATCH | `/briefs/{id}` | JWT | Edit brief content |
| DELETE | `/briefs/{id}` | JWT | Delete brief |
| POST | `/briefs/{id}/regenerate` | JWT | Re-run brief generation |
| POST | `/assets/generate` | JWT | Queue video generation |
| GET | `/assets` | JWT | List org assets |
| GET | `/assets/{id}` | JWT | Get asset detail |
| PATCH | `/assets/{id}` | JWT | Update asset tags |
| DELETE | `/assets/{id}` | JWT | Delete asset |
| POST | `/assets/{id}/retry` | JWT | Retry failed job |
| GET | `/jobs/{id}` | JWT | Poll job status |
| GET | `/jobs` | JWT | List recent jobs |
| POST | `/exports` | JWT | Create export package |
| GET | `/exports/{id}` | JWT | Get export + download URL |
| GET | `/exports` | JWT | List exports |
| DELETE | `/exports/{id}` | JWT | Delete export |
| POST | `/projects` | JWT | Create project |
| GET | `/projects` | JWT | List projects |
| GET | `/projects/{id}` | JWT | Get project |
| PATCH | `/projects/{id}` | JWT | Update project |
| DELETE | `/projects/{id}` | JWT | Delete project |
| POST | `/brands` | JWT | Create brand |
| GET | `/brands` | JWT | List brands |
| GET | `/brands/{id}` | JWT | Get brand |
| PATCH | `/brands/{id}` | JWT | Update brand |
| DELETE | `/brands/{id}` | JWT | Delete brand |
| GET | `/org` | JWT | Org profile + usage |
| GET | `/org/members` | JWT+Admin | List members |
| POST | `/org/members/invite` | JWT+Admin | Invite member |
| DELETE | `/org/members/{id}` | JWT+Admin | Remove member |
| GET | `/org/billing` | JWT+Admin | Billing info |
| GET | `/stream/briefs/{id}` | JWT | SSE: brief progress |
| GET | `/stream/assets/{id}` | JWT | SSE: asset generation progress |
| POST | `/internal/webhooks/falai` | HMAC | FAL.AI callback (internal) |
| POST | `/internal/webhooks/heygen` | HMAC | HeyGen callback (internal) |
