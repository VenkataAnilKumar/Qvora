---
title: User Personas & ICP
category: product
tags: [personas, ICP, jobs-to-be-done, agency, media-buyer, creative-director]
sources: [Qvora_Product-Definition]
updated: 2026-04-15
---

# User Personas & ICP

## TL;DR
Qvora's V1 ICP is the **mid-market performance marketing agency**. Three personas exist: the Media Buyer (primary generator), the Creative Director (primary strategist), and the Account Manager (read-only reviewer). DTC Brand Manager is Phase 2 only.

---

## Priority Ranking

| Priority | Persona | V1 Role |
|---|---|---|
| P1 | Agency Media Buyer | Primary user — generates, tests, iterates |
| P2 | Agency Creative Director | Primary strategist — owns briefs and angles |
| P3 | Agency Account Manager | Reviewer only — no generation features |
| P4 | DTC Brand Manager | **Phase 2 only. Not in V1.** |

---

## P1 — Agency Media Buyer

**Who they are:** Works at a performance marketing agency. Manages $50K–$500K/month in ad spend across 3–15 brand clients.

**Their job:** Allocate spend efficiently, maximize ROAS, keep clients happy. Creative output is the lever they pull most.

**Pain:** Creative production is a bottleneck. They can't test enough variants. Fatigue detection is manual. Performance data doesn't feed back into the brief.

**What they want from Qvora:**
- Generate 10 video variants from a URL in minutes, not days
- Know which ones are winning before the budget runs out
- Brief → ad in one system, not five

**Success metric for them:** More creatives tested per week without more production cost.

---

## P2 — Agency Creative Director

**Who they are:** Owns creative strategy and brand voice for agency clients. Approves all outgoing creative.

**Their job:** Ensure every ad reflects the brand, resonates with the audience, and has a strong hook and angle.

**Pain:** Briefing junior designers/editors takes as long as doing it. Strategy gets lost in translation between brief and final video.

**What they want from Qvora:**
- Generate the brief angles themselves (no interpretation layer)
- Edit hooks and angles inline
- Regenerate a specific angle without touching the rest

**Success metric for them:** Brand voice preserved across all variants; fewer revision rounds.

---

## P3 — Agency Account Manager

**Role:** Client-facing liaison. Presents creative to clients for approval.

**V1 scope:** Read-only. Can view generated ads and share with clients. Cannot generate, edit, or export. No generation features built for this persona in V1.

---

## P4 — DTC Brand Manager (Phase 2 only)

> ⚠️ **Not V1.** Do not build V1 features targeting this persona.

**Who they are:** In-house marketer at a DTC brand. Lower ad spend than agencies, more brand-control sensitivity.

**Why Phase 2:** Requires different onboarding, different brand-safety guardrails, different pricing sensitivity. V1 is optimized entirely for agency workflows.

---

## Jobs To Be Done

| Job | Persona | V1 |
|---|---|---|
| "Help me make 10 ad variants today, not next week" | Media Buyer | ✅ |
| "Make sure every angle reflects the brand voice" | Creative Director | ✅ |
| "Tell me which creative is burning out before the client sees bad results" | Media Buyer | V2 (Signal) |
| "Let me present creative to my client without giving them platform access" | Account Manager | ✅ |

---

## Open Questions
- [ ] What is the typical team size at target agencies? (Helps scope collaboration features)
- [ ] Do Creative Directors want to write the initial brief or edit what Qvora generates?

## Related Pages
- [[qvora-overview]] — product one-liner and strategic objective
- [[pricing]] — which tiers map to which personas
- [[features]] — feature scope per persona
