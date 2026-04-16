---
title: User Stories — Epics & Stories
category: product
tags: [user-stories, epics, acceptance-criteria, P1, P2, P3]
sources: [Qvora_User-Stories]
updated: 2026-04-15
---

# User Stories — Epics & Stories

## TL;DR
8 epics covering the full product lifecycle. EPIC 7 (Signal) is V2. EPIC 1–6 and EPIC 8 are V1. P3 (Account Manager) has no generation stories — read-only only.

---

## Epic Structure

| Epic | Name | Scope |
|---|---|---|
| EPIC 1 | Onboarding & Activation | V1 |
| EPIC 2 | Qvora Brief (Strategy Layer) | V1 |
| EPIC 3 | Qvora Studio (Video Generation) | V1 |
| EPIC 4 | Export & Structured Testing | V1 |
| EPIC 5 | Brand Kit & Multi-Brand Management | V1 |
| EPIC 6 | Team & Collaboration | V1 |
| EPIC 7 | Qvora Signal (Performance Learning) | **V2 only** |
| EPIC 8 | Platform & Administration | V1 |

**Activation target:** User creates a real, useful artifact within 60 seconds of signup. The brief IS the artifact.

---

## EPIC 1 — Onboarding & Activation

> **Goal:** First complete ad set in < 15 minutes. Real artifact in first 60 seconds.

| Story | Summary | Key Acceptance Criteria |
|---|---|---|
| US-01 | Signup + role selection | Role routes to appropriate empty state; Google OAuth; email verified before first gen |
| US-02 | First brand setup | Brand wizard first screen after verify; min viable: name + color + logo; < 2 min |
| US-03 | First brief generation (**Activation Moment**) | URL → brief in 15s; 3 angles + 3 hooks/angle shown; 1-click proceed to video |
| US-04 | Onboarding checklist | Checklist days 1–7 only; milestones: brand → brief → export → ad account (Growth+) |

**Research signal:** Role-based onboarding lifts 7-day retention by 35% (UserPilot, 2026).

---

## EPIC 2 — Qvora Brief (Strategy Layer)

> **Goal:** LLM generates creative strategy; user edits and approves before any video is rendered.

Key behaviors:
- Generate 3–5 creative angles with hooks, visual direction, rationale, and recommended format
- Inline edit of any angle name, hook text, visual direction
- Per-angle regeneration without touching other angles (**BRIEF-09** — in progress)
- Brief persisted to DB with fail-fast on insert errors (**BRIEF-08** — in progress)
- Brief history — view previous briefs for same product URL

---

## EPIC 3 — Qvora Studio (Video Generation)

> **Goal:** Brief angle → ready-to-export video, with real-time progress.

Key behaviors:
- Select angle(s) → choose format (UGC / Spokesperson / Demo / Voiceover)
- Select avatar (V2V path) and voice (ElevenLabs)
- Submit generation job → SSE progress stream in UI
- Variant count enforced by tier (3 / 10 / unlimited)
- Model selection: Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2

---

## EPIC 4 — Export & Structured Testing

> **Goal:** Organized, platform-ready export with naming convention.

Key behaviors:
- Export individual video or full ad set ZIP
- Platform presets: TikTok / Reels / Meta / YouTube Shorts
- Naming: `[brand]-[angle]-[variant]-[platform]`
- Export history — re-download any past export

---

## EPIC 5 — Brand Kit & Multi-Brand Management

> **Goal:** Persist and apply brand identity across all generations.

Key behaviors:
- Brand kit: name, primary color, secondary color, logo, tone-of-voice notes, intro/outro bumper
- Auto-apply brand kit to brief generation
- Brand switcher in topbar — switch active brand in one click
- Tier limits: 1 (Starter) / 3 (Growth) / unlimited (Agency)

---

## EPIC 6 — Team & Collaboration

> **Goal:** Agency workflow — multiple people, defined roles.

Key roles:
- **Admin** — full access, billing, team management
- **Creator** (Media Buyer / Creative Director) — generation, editing, export
- **Reviewer** (Account Manager) — view and share only; no generation

Key behaviors:
- Email invite to workspace
- Role-based action gating (creator vs. reviewer)
- Seat limits by tier

**Note:** US-19 explicitly defines the Account Manager as reviewer-only. No generation stories for P3.

---

## EPIC 7 — Qvora Signal (V2 only)

> ⚠️ Not built in V1.

- Connect ad accounts (Meta, TikTok, AppLovin)
- Ingest ROAS, CTR, spend per creative
- Classify creatives: winning / fatiguing / dead
- Auto-suggest next-gen variants based on winners
- Requires Temporal scheduling, pgvector semantic similarity

---

## EPIC 8 — Platform & Administration

Key behaviors:
- Stripe subscription management (upgrade, downgrade, cancel)
- Usage meter in dashboard (variants used / limit)
- Trial countdown banner (Day 6, Day 8)
- Billing portal (Stripe Customer Portal)
- Org settings (name, logo, seats)

---

## Open Questions
- [ ] What's the fallback UX when brief generation takes > 30s?
- [ ] Can a Reviewer share a video externally with a public link (no auth)?
- [ ] Does per-angle regeneration count against the variant quota?

## Related Pages
- [[personas]] — who each epic is for
- [[features]] — module-level detail for each epic
- [[user-journey]] — flow-level breakdown of EPIC 1 onboarding
- [[pricing]] — tier limits referenced in EPIC 3, 5, 6
