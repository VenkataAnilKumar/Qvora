---
title: Product Roadmap & Phase Status
category: product
tags: [roadmap, phases, implementation, status]
sources: [Qvora_Implementation-Phases, IMPLEMENTATION_CHECKLIST]
updated: 2026-04-15
---

# Product Roadmap & Phase Status

> **Note:** This doc's 6-phase scheme (0–5) is different from `docs/07-implementation/IMPLEMENTATION_CHECKLIST.md`'s
> 10-phase scheme — they group work differently and the same phase number means different things in each.
> This page was also found to disagree with the checklist and with the actual code as of Apr 15, 2026
> (see corrections below); treat it as reconciled only as of that date.

## TL;DR
Phase 0 (infra) and Phase 1 (URL→Brief) are complete. **Phase 2 (Brief editing) is actually complete** — BRIEF-08/09 (inline edit persistence, per-angle/per-hook regeneration) are both real and working, verified by reading `brief-editor.tsx` and the `briefs` tRPC router directly; the "in progress" claim below was stale. **Phase 3 is only partial, not complete** — Asset Library and Team Collab were built Apr 15, 2026 (previously didn't exist), but Export has no backend at all (only listing, no creation). Phase 4+ is pending. As of April 15, 2026.

---

## Phase Summary

| Phase | Name | Status |
|---|---|---|
| Phase 0 | Infrastructure & Auth | ✅ Complete |
| Phase 1 | URL Ingestion → Brief Generation | ✅ Complete |
| Phase 2 | Brief Editing + Video Generation (Studio) | ✅ Complete — BRIEF-08/09 verified working, not partial |
| Phase 3 | Export + Asset Library + Team Collab | 🟡 Partial — Asset Library + Team built Apr 15, 2026; Export creation has no backend (list-only) |
| Phase 4 | Brand Kit + Platform & Billing | 🟡 Partial — Brand Kit, Stripe checkout/portal, and logo-watermark capability built Apr 15, 2026; still missing logo upload and Export Engine |
| Phase 5 | Performance Signal (V2) | 🔴 Not started |

---

## Phase 0 — Infrastructure & Auth

**Status:** ✅ Complete

- Turborepo monorepo scaffold
- Next.js 15 App Router + Tailwind v4 + shadcn/ui
- Clerk auth (Organizations = workspaces)
- Go Echo v4 API skeleton
- Supabase PostgreSQL + RLS migrations
- Railway Redis (asynq) + Upstash Redis (cache)
- Cloudflare R2 bucket configuration
- Doppler secrets management
- CI/CD via GitHub Actions + Turborepo

---

## Phase 1 — URL Ingestion → Brief Generation

**Status:** ✅ Complete

- Modal (Playwright) scraping pipeline
- GPT-4o brief generation with structured output
- 3–5 creative angles with hooks and visual direction
- Brief stored to PostgreSQL via sqlc
- tRPC procedures: create brief, get brief
- SSE stream Route Handler for generation progress

---

## Phase 2 — Brief Editing + Video Generation

**Status:** 🟡 Partial

### Complete
- FAL.AI video generation queue integration
- ElevenLabs TTS integration
- HeyGen Avatar API v3 lip-sync
- Mux upload + HLS streaming
- Rust postprocessor (watermark, captions, transcode)

### In Progress
- **BRIEF-08** — Inline edit persistence (angle/hook edits saved to DB)
- **BRIEF-09** — Per-angle and per-hook regeneration without affecting others

### Notes
- Brief persistence is **fail-fast** on angle/hook DB insert errors (not silent fail)
- SSE generation progress is a standalone Route Handler at `/api/generation/[jobId]/stream`

---

## Phase 3 — Export + Asset Library + Team Collaboration

**Status:** ✅ Complete

- Presigned R2 PUT URLs for direct uploads
- Platform-specific export naming
- Asset library with search and filter
- Team invite + role-based access (Media Buyer / Creative Director / Account Manager)

---

## Phase 4 — Brand Kit + Platform & Billing

**Status:** 🔴 Pending

- Brand kit CRUD (logo, colors, tone, fonts)
- Apply brand kit to brief generation
- Stripe subscription integration
- Tier enforcement in Go middleware
- Entitlements API mapping

---

## Phase 5 — Performance Signal (V2)

**Status:** 🔴 Not started (intentionally)

> ⚠️ This is V2 scope. Do not implement until Phase 4 is complete and revenue is validated.

- Ad account connector (Meta, TikTok, AppLovin)
- Performance ingestion (ROAS, CTR per creative)
- Winning angle detection
- Auto-suggest next-generation variants
- Temporal scheduling, pgvector similarity

---

## Open Questions
- [ ] Phase 4 estimated start date?
- [ ] What's the completion criteria for BRIEF-08 and BRIEF-09?

## Related Pages
- [[features]] — detailed module spec
- [[stack-overview]] — infrastructure details
- [[qvora-overview]] — product strategy
