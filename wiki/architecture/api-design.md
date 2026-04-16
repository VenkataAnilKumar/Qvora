---
title: API Design
category: architecture
tags: [api, rest, trpc, sse, go-echo, endpoints, auth, jwt]
sources: [Qvora_API-Design]
updated: 2026-04-15
---

# API Design

## TL;DR
Two API layers: tRPC BFF (TypeScript, Vercel) for frontend ↔ Go bridge, and Go Echo v4 REST API (Railway) for business logic. SSE generation stream is a **standalone Next.js Route Handler** — not tRPC. Bearer JWT from Clerk on every request. Cursor-based pagination.

---

## Base URLs

| Layer | URL |
|---|---|
| Public REST API | `https://api.qvora.com/v1` |
| Internal tRPC BFF | `http://localhost:3001/trpc` (Next.js → Go bridge) |
| SSE Stream | `/api/generation/[jobId]/stream` (Next.js Route Handler) |

---

## Conventions

- **Auth:** `Authorization: Bearer <Clerk JWT>` on every request
- **Content-Type:** `application/json`
- **Dates:** ISO 8601
- **UUIDs:** lowercase hyphenated
- **Pagination:** cursor-based (`cursor` + `limit`; default 20, max 100)
- **Idempotency:** POST resource creation accepts `Idempotency-Key` header; duplicate within 24h returns original response

### Response Envelope

```json
{
  "data": { ... },
  "meta": {
    "request_id": "req_01HZ...",
    "timestamp": "2026-04-14T10:30:00Z"
  }
}
```

List responses add:
```json
"pagination": { "cursor": "...", "has_more": true, "total": 147 }
```

---

## JWT Payload

```json
{
  "sub": "user_01HZ...",
  "org_id": "org_01HZ...",
  "org_role": "admin | member",
  "plan": "starter | growth | agency"
}
```

**Tier limits enforced in Go middleware** — never client-side.

---

## Scopes

| Scope | Permission |
|---|---|
| `read:briefs` | View briefs in org |
| `write:briefs` | Create/update/delete briefs |
| `read:assets` | View generated assets |
| `write:assets` | Trigger generation, update metadata |
| `read:exports` | Download export packages |
| `write:exports` | Create export jobs |

---

## Endpoints (Go REST API)

### Briefs

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/briefs` | Create brief from URL or manual input |
| `GET` | `/v1/briefs` | List briefs for org (paginated) |
| `GET` | `/v1/briefs/:id` | Get single brief with angles + hooks |
| `PATCH` | `/v1/briefs/:id` | Update brief metadata |
| `DELETE` | `/v1/briefs/:id` | Soft-delete brief |
| `POST` | `/v1/briefs/:id/regenerate` | Regenerate full brief |
| `POST` | `/v1/briefs/:id/angles/:angleId/regenerate` | Regenerate single angle |

### Generation Jobs

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/generations` | Submit generation job |
| `GET` | `/v1/generations/:jobId` | Get job status + progress |
| `GET` | `/v1/generations` | List jobs for org |

### Assets

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/assets` | List assets in org (filterable by brief, status, platform) |
| `GET` | `/v1/assets/:id` | Get single asset with signed playback URL |
| `PATCH` | `/v1/assets/:id` | Update metadata (title, tags) |
| `DELETE` | `/v1/assets/:id` | Soft-delete asset |

### Exports

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/exports` | Create export job (selected assets + platform preset) |
| `GET` | `/v1/exports/:id` | Get export status + download URL |
| `GET` | `/v1/exports` | List exports for org |

### Brands

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/brands` | Create brand kit |
| `GET` | `/v1/brands` | List brands for org |
| `GET` | `/v1/brands/:id` | Get brand kit |
| `PATCH` | `/v1/brands/:id` | Update brand kit |
| `DELETE` | `/v1/brands/:id` | Delete brand |

### Webhooks — FAL.AI Callbacks

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/webhooks/fal` | Receive FAL.AI job completion callback |

> ⚠️ FAL.AI always uses `fal.queue.submit()` (async). Callbacks arrive at this webhook — never `fal.subscribe()` (blocks worker).

### SSE Stream (NOT Go — Next.js Route Handler)

```
GET /api/generation/[jobId]/stream
```

- Located at `src/apps/web/app/api/generation/[jobId]/stream/route.ts`
- Returns `text/event-stream` with `ReadableStream`
- Events: `queued` | `scraping` | `generating` | `postprocessing` | `complete` | `failed`
- **This is NOT a tRPC subscription** — it is a standalone Next.js Route Handler

---

## tRPC BFF Procedures

The tRPC layer translates between the frontend (TanStack Query) and the Go REST API. Key procedures:

| Router | Procedure | Type | Description |
|---|---|---|---|
| `briefs` | `create` | mutation | POST to Go `/v1/briefs` |
| `briefs` | `list` | query | GET briefs list |
| `briefs` | `byId` | query | GET single brief |
| `briefs` | `regenerate` | mutation | POST regenerate angle |
| `generations` | `submit` | mutation | POST generation job |
| `assets` | `list` | query | GET assets list |
| `exports` | `create` | mutation | POST export job |
| `brands` | `list` | query | GET brands for org |
| `brands` | `create` | mutation | POST new brand |

---

## Error Codes

| Code | HTTP | Meaning |
|---|---|---|
| `AUTH_REQUIRED` | 401 | No valid JWT |
| `FORBIDDEN` | 403 | Valid JWT, wrong scope/role |
| `NOT_FOUND` | 404 | Resource not found or not owned by org |
| `PLAN_LIMIT_EXCEEDED` | 402 | Variant/brand limit hit for tier |
| `VALIDATION_ERROR` | 422 | Request body fails Zod/Go validation |
| `GENERATION_FAILED` | 500 | FAL.AI / ElevenLabs / HeyGen error |

---

## Rate Limits

| Tier | Rate |
|---|---|
| All tiers | 100 requests/minute per org (Upstash Redis sliding window) |
| Generation submit | 10/minute per org |
| Brief create | 20/minute per org |

---

## Open Questions
- [ ] Does the public REST API need API keys for a developer/partner tier or is it Clerk JWT only?
- [ ] Should `/v1/assets` support bulk operations (batch delete, batch download)?

## Related Pages
- [[system-architecture]] — where this API fits in the component diagram
- [[stack-overview]] — Go Echo v4, tRPC, Clerk JWT
- [[database-schema]] — data models behind these endpoints
