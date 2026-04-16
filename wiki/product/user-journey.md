---
title: User Journey & Onboarding Flow
category: product
tags: [user-journey, onboarding, activation, screens, flow]
sources: [Qvora_User-Journey]
updated: 2026-04-15
---

# User Journey & Onboarding Flow

## TL;DR
7 journey phases from awareness to Signal activation. The activation moment is Phase 2 (first brief in < 60s). Every screen must earn its place. Dark-first, low-friction signup with Google OAuth as the primary path.

---

## Design Principle

> *The user creates a real, useful artifact within 60 seconds of signup. Every screen must earn its place.*

The brief IS the artifact. Not a tutorial. Not a walkthrough. A real creative strategy document for their actual product, in under 60 seconds.

---

## Journey Phases

| Phase | Name | Timing |
|---|---|---|
| Phase 0 | Awareness & Acquisition | Pre-signup |
| Phase 1 | Signup & Onboarding | Day 0 |
| Phase 2 | First Brief — **Activation Moment** | Day 0 |
| Phase 3 | First Video Set | Day 0–1 |
| Phase 4 | First Export & Launch | Day 1–3 |
| Phase 5 | Retention & Habit Formation | Week 2–4 |
| Phase 6 | Expansion | Month 2+ |
| Phase 7 | Signal Activation (V2) | Month 3+ |

---

## Phase 0 — Awareness & Acquisition

**Entry points:**

| Channel | Message | CTA |
|---|---|---|
| LinkedIn paid (agency targeting) | *"Your competitor launched 40 creatives this week. You launched 3."* | Start free trial |
| Google Search | Qvora vs. Creatify comparison landing page | Try Qvora free |
| Word-of-mouth / agency Slack | *"We cut brief-to-launch from 5 days to 45 minutes"* | Book a demo |
| Product Hunt launch | *"URL → Strategy → Video Ads in 15 minutes"* | Get early access |

**Key landing page message:**
> *"Paste a URL. Get your strategy. Get your ads. Know what wins."*

The **"strategy-first"** framing is the product's key differentiator vs. Creatify ("templates-first").

---

## Phase 1 — Signup & Onboarding

### Screen 1.1 — Signup
- Google OAuth is the **primary CTA** (reduces friction)
- Email path: only 2 fields — email + password (name/company collected progressively)
- "No credit card required" copy — removes single biggest signup barrier
- Email verification required before first generation

### Screen 1.2 — Role & Context (60 seconds)
- Question 1: Role (Media Buyer / Creative Director / Agency Owner / In-house Brand Manager)
- Question 2: Monthly ad spend range (< $10K / $10K–$100K / $100K–$500K / > $500K)
- Question 3: Company/agency name
- Drives: role-appropriate empty states, personalized placeholder copy in briefs

### Screen 1.3 — Brand Setup (2 minutes)
- **First screen after verify** — not a dashboard
- Minimum: brand name + primary color + logo
- Optional: secondary color, font, tone-of-voice notes, intro/outro bumper
- Progressive disclosure — skip extended inputs, return later

---

## Phase 2 — First Brief (Activation Moment)

- URL input field is the **first CTA after brand setup** — no additional navigation
- System generates brief in < 15 seconds
- Brief preview: 3 angles × 3 hook variants + recommended format + rationale
- One-click approve → proceed to video generation
- Inline editing of any angle/hook before proceeding

**Empty state copy:** *"Paste your product URL. Get your brief in 15 seconds."*

**Why this is the activation moment:** The user has produced a real, intelligent creative strategy for their actual product. This is the value. Everything after is delivery.

---

## Phase 3 — First Video Set

- User selects 1–3 angles from the Brief
- Chooses format per angle (UGC / Spokesperson / Demo / Voiceover)
- Selects avatar (V2V) and voice (ElevenLabs)
- Hits "Generate"
- **SSE progress stream** shows real-time status per video (queued → rendering → complete)
- Video appears in-product on completion, playable via Mux

---

## Phase 4 — First Export & Launch

- User selects videos for export
- Chooses platform preset (TikTok / Reels / Meta / YouTube Shorts)
- System generates ZIP with videos + brief PDF
- Download or share link

---

## Phase 5 — Retention & Habit Formation

- Onboarding checklist disappears (day 7 or all milestones complete)
- Usage habit: open Qvora → new brief per client per campaign cycle
- Key retention lever: **multiple brands** — agency users who set up 3+ brands churn at much lower rates

---

## Phase 6 — Expansion

- Invite team members (Creative Director, Account Manager)
- Upgrade tier for more variants / more brand kits
- Add more client brands

---

## Phase 7 — Signal Activation (V2)

> ⚠️ V2 only.

- Connect ad account (Meta / TikTok)
- Signal dashboard shows which creatives are winning, fatiguing, dead
- Auto-suggest regeneration of top-performing angles

---

## Open Questions
- [ ] What is the day-3 and day-6 email copy?
- [ ] What happens on day 8 trial expiry UX? (Generation locked — what does the screen look like?)
- [ ] Is there a public share link for Account Managers to send videos to clients?

## Related Pages
- [[user-stories]] — story-level detail for each phase
- [[personas]] — who goes through each phase
- [[features]] — module-level feature spec
- [[pricing]] — trial timeline and conversion emails
