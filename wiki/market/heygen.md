---
title: HeyGen — Competitor & Vendor Profile
category: market
tags: [heygen, competitor, vendor, avatar, lip-sync, V2V]
sources: [Qvora_Competitive-Analysis, Qvora_Architecture-Stack]
updated: 2026-04-15
---

# HeyGen — Competitor & Vendor Profile

## TL;DR
HeyGen is both a market competitor (avatar-first video platform) and a Qvora vendor component (used for V2V lip-sync). Qvora uses HeyGen Avatar API **v3** — v4 does not exist; any "v4" reference in docs is an error.

---

## Market Profile

| Field | Value |
|---|---|
| Website | heygen.com |
| Tagline | AI Video Creation Platform |
| Core strength | Avatar quality, multilingual video, translation |
| Primary weakness | Not built for paid ads; no brief/strategy engine |
| Pricing entry | $29/mo |

---

## What They Do

- High-quality digital avatars (photorealistic, diverse)
- Video translation + lip-sync (multilingual)
- Spokesperson-style video creation
- Strong enterprise adoption for explainer and training videos

---

## Why HeyGen Is in Qvora's Stack

Qvora uses HeyGen **Avatar API v3** for Voice-to-Video (V2V) lip-sync generation — the spokesperson format where a real or digital avatar speaks the brand's voiceover.

> ⚠️ **Non-negotiable:** HeyGen = v3 only. Active platform: `developers.heygen.com`. V2V lip-sync is a v3-only capability. There is NO HeyGen v4. Any "v4" reference in any document is an error.

---

## HeyGen as Competitor vs. Vendor

| View | Status |
|---|---|
| As competitor | Low overlap — HeyGen is not optimized for performance ad workflows. Their users make explainer videos, not ad variants. |
| As vendor | Critical dependency — Qvora relies on HeyGen API v3 for the V2V format. Risk: HeyGen API pricing changes or rate limits could affect Qvora unit economics. |

> 🔍 Gap: Document HeyGen API v3 rate limits and cost per rendered minute. Important for pricing model validation.

---

## Open Questions
- [ ] What is HeyGen's pricing per API render minute?
- [ ] Is there a viable alternative to HeyGen for V2V (e.g., D-ID, Synthesia API)?
- [ ] Does HeyGen have a caching API for pre-rendered avatar base assets?

## Related Pages
- [[competitive-landscape]] — full competitor comparison
- [[ai-layer]] — HeyGen in Qvora's AI stack
- [[features]] — V2V feature module (V2V)
