# QVORA
## User Journey & Onboarding Flow
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Draft

---

## Overview

This document maps the end-to-end journey of Qvora's primary user — an **Agency Media Buyer or Creative Director** — from first awareness through a fully launched, Signal-connected campaign. It covers activation milestones, screen-level flow, empty states, and key decision points.

**Design principle:** The user creates a real, useful artifact within 60 seconds of signup. Every screen must earn its place.

---

## Journey Phases

```
Phase 0 — Awareness & Acquisition
Phase 1 — Signup & Onboarding (Day 0)
Phase 2 — First Brief (Day 0 — Activation Moment)
Phase 3 — First Video Set (Day 0–1)
Phase 4 — First Export & Launch (Day 1–3)
Phase 5 — Retention & Habit Formation (Week 2–4)
Phase 6 — Expansion (Month 2+)
Phase 7 — Signal Activation (V2, Month 3+)
```

---

## Phase 0 — Awareness & Acquisition

**User state:** Has a problem. Spending too long briefing, editing, or waiting on creative. Probably just saw a competitor's creative volume and felt behind.

**Entry points:**
| Channel | Message | CTA |
|---|---|---|
| LinkedIn paid (agency targeting) | *"Your competitor launched 40 creatives this week. You launched 3."* | Start free trial |
| Google Search ("AI video ad maker for agencies") | Qvora vs. Creatify comparison landing page | Try Qvora free |
| Word of mouth / agency Slack communities | *"We cut brief-to-launch from 5 days to 45 minutes"* | Book a demo |
| Product Hunt launch | Hero: *"URL → Strategy → Video Ads in 15 minutes"* | Get early access |

**Key landing page message:**
> *"Paste a URL. Get your strategy. Get your ads. Know what wins."*
> Strategy-first. Not just video.

---

## Phase 1 — Signup & Onboarding

### Screen 1.1 — Signup

**Goal:** Account created in < 2 minutes with no friction.

```
┌─────────────────────────────────────────────────┐
│  ◉  QVORA                                       │
│                                                 │
│  Start your free 7-day trial                    │
│  No credit card required.                       │
│                                                 │
│  [G] Continue with Google                       │
│  ─────────── or ───────────                     │
│  Work email    [________________]               │
│  Password      [________________]               │
│                                                 │
│  [  Create account  ]                           │
│                                                 │
│  Already have an account? Sign in              │
└─────────────────────────────────────────────────┘
```

**Rules:**
- Google OAuth is the primary CTA (reduces friction)
- Only 2 fields for email path — no name, company, phone at this step
- "No credit card" removes the single biggest signup barrier

---

### Screen 1.2 — Role & Context

**Goal:** Route user to the right empty state. Collect context that powers personalization.

```
┌─────────────────────────────────────────────────┐
│  Quick setup — takes 60 seconds                 │
│                                                 │
│  What describes you best?                       │
│                                                 │
│  ○  Media Buyer / Performance Marketer          │
│  ○  Creative Director                           │
│  ○  Agency Owner / Account Manager             │
│  ○  In-house Brand Manager                     │
│                                                 │
│  Monthly ad spend you manage                   │
│  ○  < $10K    ○  $10K–$100K                    │
│  ○  $100K–$500K    ○  > $500K                  │
│                                                 │
│  Company / Agency name  [________________]      │
│                                                 │
│  [  Continue →  ]                               │
└─────────────────────────────────────────────────┘
```

**Routing logic:**
| Role | Routed to |
|---|---|
| Media Buyer | Brief-first dashboard: "Let's build your first test set" |
| Creative Director | Brief-first dashboard: "Let's build your first campaign brief" |
| Agency Owner | Brand setup: "Set up your first client brand" |
| In-house Brand Manager | Brand setup: "Set up your brand" |

---

### Screen 1.3 — First Brand Setup

**Goal:** Minimum viable brand kit before first generation. Must complete in < 2 minutes.

```
┌─────────────────────────────────────────────────┐
│  Set up your first brand                        │
│  This takes 2 minutes. You can add more later.  │
│                                                 │
│  Brand name   [________________]                │
│                                                 │
│  Primary color                                  │
│  [#______]  ████  (color picker)               │
│                                                 │
│  Logo (optional for now)                        │
│  [  Upload PNG or SVG  ]  or  Skip →            │
│                                                 │
│  [  Create brand →  ]                           │
│                                                 │
│  ─────────────────────────────────             │
│  You can add fonts, intro/outro, and           │
│  tone notes in Brand Settings later.           │
└─────────────────────────────────────────────────┘
```

**Design rules:**
- Logo is optional — don't block activation on brand completeness
- Color picker defaults to a neutral (agency can change later)
- "Skip →" for logo routes to same next screen
- Progressive disclosure: advanced brand settings revealed after first ad is generated

---

## Phase 2 — First Brief (Activation Moment)

> **Definition of activation:** User generates their first complete brief with ≥ 1 approved angle.
> **Target:** < 15 minutes from signup. Ideally < 5 minutes.

### Screen 2.1 — URL Input (The Hook)

**Goal:** The first screen after brand setup must show immediate value. No empty dashboard.

```
┌─────────────────────────────────────────────────┐
│  ◉  QVORA             Acme Co ▼    [+ New Brand]│
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │                                           │ │
│  │  Paste your product URL                   │ │
│  │                                           │ │
│  │  [  https://yourproduct.com/page  ] [→]  │ │
│  │                                           │ │
│  │  Or  [paste a product brief instead]      │ │
│  │                                           │ │
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ────────────────────────────────────           │
│  💡 Works with Shopify, WooCommerce,            │
│     App Store, Google Play, and landing pages  │
└─────────────────────────────────────────────────┘
```

**Empty state copy:** *"Paste your product URL. Get your brief in 15 seconds."*

**What happens on submit:**
1. Loading state: *"Reading your product page..."* (< 10s)
2. Extraction complete: *"Building your creative brief..."* (< 15s)
3. Brief preview appears

---

### Screen 2.2 — Brief Preview

**Goal:** Show the output in full. Make it feel intelligent — not generic.

```
┌──────────────────────────────────────────────────────────────┐
│  Creative Brief — Acme Protein Bar     [Edit] [Approve →]    │
│  Based on: acme.com/protein-bar · Generated in 12s           │
│                                                              │
│  Product Summary                                             │
│  Clean-label, 25g protein bar for active consumers.         │
│  Key differentiator: real food ingredients, no artificial    │
│  sweeteners. Target: gym-goers and health-conscious adults.  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Angle 1 — CONVERSION   ● Desire                     │   │
│  │  "The bar that doesn't taste like compromise"        │   │
│  │  Funnel stage: Conversion · Emotion: Aspiration      │   │
│  │                                                      │   │
│  │  Hooks:                                              │   │
│  │  H1 [desire]    "Real food. 25g protein. Finally."  │   │
│  │  H2 [problem]   "Tired of protein bars that taste   │   │
│  │                  like chalk? Meet Acme."            │   │
│  │  H3 [soc.proof] "10K athletes switched. Here's why."│   │
│  │                            [Regenerate hooks]        │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  [+ Angle 2]  [+ Angle 3]  ...                               │
│                                                              │
│  Recommended formats: ① UGC-style  ② Product Demo           │
│  Platform fit: Meta Feed · TikTok · Instagram Stories        │
│                                                              │
│  [  ← Edit brief  ]    [  Approve & Generate Videos →  ]    │
└──────────────────────────────────────────────────────────────┘
```

**Key UX rules:**
- User sees real content — angles with rationale, hooks with types — not placeholder text
- Edit is always one click away — no locked-in brief
- "Approve & Generate" is the primary CTA — single click to next phase
- Brief quality indicator (confidence score) shown subtly if extraction was partial

---

## Phase 3 — First Video Set

### Screen 3.1 — Generation Settings

**Goal:** Let the user choose how they want to generate — brief-driven formats or direct AI generation primitives. Don't overwhelm: show the smart default first, reveal power features on demand.

```
┌──────────────────────────────────────────────────────────────┐
│  Generate your ad set                                       │
│  3 angles approved · Generating 1 video per angle           │
│                                                             │
│  How do you want to generate?                               │
│                                                             │
│  ┌─── From Brief (Recommended) ──────────────────────────┐  │
│  │  [●] UGC-style    Avatar + voiceover, talking-head    │  │
│  │  [ ] Product Demo  Product shots + B-roll + voiceover │  │
│  │  [ ] Text-Motion   Animated copy + product image      │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─── AI Generation Modes ───────────────────────────────┐  │
│  │  [ ] Text → Video   Script/prompt → cinematic video   │  │
│  │  [ ] Image → Video  Product image → animated clip     │  │
│  │  [ ] Voice → Video  Your voice → lip-synced avatar    │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                             │
│  Platform   [●] Meta Feed (1:1)  [ ] TikTok (9:16) [ ] Both │
│  Duration   [ ] 15s  [●] 30s  [ ] 60s                       │
│                                                             │
│  Avatar     [  Select avatar  ▼  ]                          │
│  Voice      [  Select voice   ▼  ]                          │
│                                                             │
│  [  Generate 3 videos →  ]                                  │
│  Estimated time: ~4 minutes                                 │
└──────────────────────────────────────────────────────────────┘
```

**AI Generation Mode sub-screens (shown when selected):**

```
TEXT → VIDEO selected:
┌──────────────────────────────────────────────────────────────┐
│  Text → Video                                               │
│                                                             │
│  Prompt   [Use brief script ✓]  or  [Write custom prompt]  │
│  Model    [●] Auto  [ ] Veo 3.1  [ ] Kling 3.0  [ ] Runway │
│  Style    [●] Cinematic  [ ] UGC  [ ] Product  [ ] Abstract │
│  Duration [ ] 5s  [ ] 10s  [●] 15s  [ ] 30s                 │
│                                                             │
│  Veo 3.1 selected → [●] Include native audio (SFX)         │
│  Runway selected  → Camera: [ ] Static [●] Zoom  [ ] Pan   │
└──────────────────────────────────────────────────────────────┘

IMAGE → VIDEO selected:
┌──────────────────────────────────────────────────────────────┐
│  Image → Video                                              │
│                                                             │
│  Image   [Use product images from brief ✓]  or  [Upload]   │
│           ┌──────┐  ┌──────┐  ┌──────┐                    │
│           │ img1 │  │ img2 │  │ + Add│                    │
│           └──────┘  └──────┘  └──────┘                    │
│                                                             │
│  Motion  [  gentle zoom in on product, soft light  ]       │
│          Templates: [Subtle Focus] [Rotation] [Pan] [Splash]│
│  Intensity  [ ] Subtle  [●] Medium  [ ] Dynamic            │
│  Duration   [ ] 3s  [●] 5s  [ ] 8s  [ ] 10s               │
│  Model      [●] Kling 3.0  [ ] Runway  [ ] SVD             │
└──────────────────────────────────────────────────────────────┘

VOICE → VIDEO selected:
┌──────────────────────────────────────────────────────────────┐
│  Voice → Video                                              │
│                                                             │
│  Voice source                                               │
│  [●] Upload audio   [  Choose file  ]  (MP3, WAV, M4A)     │
│  [ ] ElevenLabs     [  Select voice  ▼  ]                   │
│  [ ] Brand voice clone  ← Growth+ only                     │
│                                                             │
│  Avatar    [  Select avatar  ▼  ]                           │
│  Language  [●] English  [ ] Other ▼  (175+ supported)      │
│  Background [●] Lifestyle  [ ] Solid  [ ] Product-matched  │
└──────────────────────────────────────────────────────────────┘
```

**Rules:**
- "From Brief" modes are the default (recommended path) — AI Generation Modes are secondary
- Selecting an AI generation mode collapses the "From Brief" section and expands mode-specific settings
- Default brief-driven format = UGC-style; default AI mode if selected = Text → Video
- Avatar and voice pre-populated from brand kit if set
- "Estimated time" shown per mode: T2V ~1–2 min, I2V ~30–60s, V2V ~2–3 min

---

### Screen 3.2 — Generation Progress

**Goal:** Keep user engaged during the wait. Don't let them navigate away thinking it failed.

```
┌──────────────────────────────────────────────────────┐
│  Building your ad set...                            │
│                                                     │
│  Angle 1 — Conversion (Desire)      ████████░ 80%  │
│  Angle 2 — Awareness (Curiosity)    █████░░░░ 50%  │
│  Angle 3 — Consideration (Fear)     ██░░░░░░░ 20%  │
│                                                     │
│  ⏱ ~3 minutes remaining                            │
│                                                     │
│  We'll email you when they're ready if you need     │
│  to step away.                                      │
│                                                     │
│  [  Notify me by email  ]                           │
└──────────────────────────────────────────────────────┘
```

---

### Screen 3.3 — Ad Set Ready

**Goal:** Deliver the moment of delight. Make results feel premium and ready to launch.

```
┌──────────────────────────────────────────────────────┐
│  ✓ Your ad set is ready                             │
│  3 videos · Acme Protein Bar · 30s · Meta Feed      │
│                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  [▶ vid] │  │  [▶ vid] │  │  [▶ vid] │          │
│  │ Angle 1  │  │ Angle 2  │  │ Angle 3  │          │
│  │ Desire   │  │ Curiosity│  │ Fear     │          │
│  └──────────┘  └──────────┘  └──────────┘          │
│                                                     │
│  [  Export all  ]    [  Add variants  ]             │
│  [  Edit a video  ]  [  Generate more formats  ]    │
└──────────────────────────────────────────────────────┘
```

**Success state copy (brand voice — celebratory, brief):**
> *"3 ads ready. Time to find your winner."*

---

## Phase 4 — First Export & Launch

### Screen 4.1 — Export

```
┌──────────────────────────────────────────────────────┐
│  Export ad set                                      │
│  3 assets selected · Acme Protein Bar               │
│                                                     │
│  Format         [●] MP4 H.264  [ ] MOV             │
│  Resolution     [●] 1080p  [ ] 720p                 │
│                                                     │
│  ✓ Naming convention applied:                       │
│    acme-co_conversion_ugc_desire_20260501_v1.mp4    │
│    acme-co_awareness_ugc_curiosity_20260501_v1.mp4  │
│    acme-co_consideration_ugc_fear_20260501_v1.mp4   │
│                                                     │
│  ✓ manifest.csv included                           │
│                                                     │
│  [  Download ZIP  ]                                 │
│  Approx. 180 MB · Ready in ~20 seconds             │
└──────────────────────────────────────────────────────┘
```

**Activation milestone triggered:** "First export" milestone marked complete on onboarding checklist.

---

## Phase 5 — Retention & Habit Formation (Week 2–4)

**Goal:** User returns twice a week. Qvora becomes the first step in every campaign.

### Key retention triggers

| Trigger | Timing | Channel | Message |
|---|---|---|---|
| "Finish your setup" | Day 2 if brand kit incomplete | Email | *"Your brand kit is 60% done. Add your font and outro to complete it."* |
| "You haven't created a brief this week" | Day 7 if no activity | Email | *"Your competitors are testing. Are you?"* |
| Weekly usage digest | Every Monday | Email | *"Last week: 12 ads generated · 3 angles tested · 0 winners yet — let's change that."* |
| Plan limit warning | At 80% usage | In-app + email | *"You've used 16 of 20 ads this month. Upgrade to Growth for 100/mo."* |

### Retention screen — Dashboard (returning user)

```
┌────────────────────────────────────────────────────────────────┐
│  ◉  QVORA             Acme Co ▼                [+ New Brief]   │
│                                                               │
│  Good morning, Sarah.                                         │
│                                                               │
│  This month           This week                              │
│  12 ads generated     3 briefs                               │
│  2 brands active      1 export                               │
│  8 ads exported                                              │
│                                                               │
│  Recent briefs                              [View all]        │
│  ┌──────────────────────┐ ┌──────────────────────┐           │
│  │ Acme Protein Bar     │ │ Acme Energy Drink    │           │
│  │ 3 angles · 9 ads     │ │ Draft · 0 ads        │           │
│  │ Exported Apr 10      │ │ [Continue →]         │           │
│  └──────────────────────┘ └──────────────────────┘           │
│                                                               │
│  [  + New brief  ]                                            │
└────────────────────────────────────────────────────────────────┘
```

---

## Phase 6 — Expansion

### Expansion trigger points

| Moment | Trigger | CTA |
|---|---|---|
| Hit brand kit limit | User tries to add 4th brand on Growth plan | *"You've reached your 3-brand limit. Upgrade to Agency for unlimited brands."* |
| Hit ad limit | 80% of monthly ads used | *"Running low? Agency tier = unlimited ads."* |
| Team invite needed | User tries to invite 6th member on Growth | *"Add unlimited seats on Agency."* |
| Signal curiosity | User asks about performance | *"Connect your Meta account to see which angles are actually winning."* (V2 upgrade prompt) |

---

## Phase 7 — Signal Activation (V2, Month 3+)

**Goal:** Turn Qvora from a generation tool into a learning system. This is the moat.

### Signal onboarding flow

```
Step 1: Connect ad account
  → "Connect your Meta account to start learning from your campaigns."
  → OAuth flow: < 3 steps
  → First sync: 24 hours

Step 2: Asset matching
  → Qvora matches exported filenames to ad creative IDs in Meta
  → "We found 8 of your Qvora-generated ads in your Meta account."

Step 3: Performance appears
  → Performance dashboard unlocked after 1,000 impressions minimum
  → "Your Desire hook is outperforming Problem hooks by 2.4x CTR."

Step 4: Recommendations
  → Next brief for this brand pre-populated with top-performing angle
  → "Based on your last campaign, start with Conversion · Desire hooks."
```

### Signal dashboard screen

```
┌────────────────────────────────────────────────────────────────┐
│  Qvora Signal — Acme Co              Last sync: 2h ago  [↻]   │
│                                                               │
│  Last 30 days  ▼          [Meta] [TikTok]                    │
│                                                               │
│  By Angle Type            CTR     CPA      Spend             │
│  ─────────────────────────────────────────────────           │
│  ● Conversion             3.8%    $4.20    $1,200            │
│  ● Awareness              1.9%    $8.10    $800              │
│  ● Consideration          2.4%    $6.40    $600              │
│                                                               │
│  By Hook Type             CTR     CPA                        │
│  ─────────────────────────────────────────────────           │
│  Desire                   4.1%    $3.90    ▲ Best            │
│  Problem                  2.8%    $5.20                      │
│  Social Proof             2.1%    $6.80                      │
│                                                               │
│  ⚠ Fatigue detected: acme_conversion_ugc_desire_v1           │
│    CTR dropped 35% over 3 days. [Refresh this ad →]          │
│                                                               │
│  [  Start new brief with Signal insights  ]                  │
└────────────────────────────────────────────────────────────────┘
```

---

## Activation Milestones Summary

| Milestone | Target | In-product Trigger |
|---|---|---|
| **M1 — Signed up** | < 2 min from landing page | Welcome email |
| **M2 — Brand created** | Day 0 | Checklist item 1 ✓ |
| **M3 — First brief generated** | < 15 min from signup | Checklist item 2 ✓ |
| **M4 — First video exported** | Day 0–1 | Checklist item 3 ✓; "You're live" banner |
| **M5 — Second brief (habit)** | Day 3–7 | No prompt needed — user returns organically |
| **M6 — Second brand added** | Week 2–4 | Signals agency ICP fit |
| **M7 — Team member invited** | Month 1 | Expansion signal |
| **M8 — Ad account connected (V2)** | Month 3+ | Signal activation; retention anchor |

---

## Empty States

| Screen | Empty State Copy | Primary CTA |
|---|---|---|
| Brief Library (no briefs yet) | *"No briefs yet. Paste your product URL and watch what happens."* | Paste a URL |
| Asset Library (no assets yet) | *"No ads yet. Approve a brief to generate your first set."* | Go to briefs |
| Signal Dashboard (not connected) | *"Connect your Meta account to see which ads are actually winning."* | Connect account |
| Signal Dashboard (connected, < 1K impressions) | *"Collecting data… Signal unlocks after 1,000 impressions per variant."* | View assets |
| Team (no members yet) | *"It's just you for now. Invite your team to collaborate on briefs and creatives."* | Invite teammate |

---

## Error States

| Error | Copy | Action |
|---|---|---|
| URL extraction failed | *"We couldn't read that page. Try pasting your product description instead."* | Open manual input |
| Generation timeout | *"This one's taking longer than expected. We'll notify you when it's ready."* | Email notification |
| Ad account connection failed | *"Connection interrupted. Check your permissions and try again."* | Retry / Help doc link |
| Plan limit reached | *"You've used all 20 ads this month. Upgrade to continue generating."* | Upgrade CTA |
| Export failed | *"Export failed. Try again — your assets are safe."* | Retry |

---

## Key UX Principles Applied

| Principle | Application |
|---|---|
| **60-second artifact** | First brief is visible before user even configures anything — URL input is step 1 post-brand-setup |
| **Role-based routing** | Media Buyer → brief-first; Agency Owner → brand-setup-first |
| **Progressive disclosure** | Advanced brand kit fields, team settings, and Signal shown only after activation milestones |
| **Expectation setting** | Generation time always shown; email fallback always offered |
| **No dead ends** | Every empty state and error state has a primary CTA |
| **Brand voice** | All copy is direct, confident, brief — *"3 ads ready. Time to find your winner."* not *"Your generation process has completed successfully."* |

---

*User Journey v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources used in research:**
- [SaaS Onboarding Flows That Actually Convert in 2026 — SaaSUI](https://www.saasui.design/blog/saas-onboarding-flows-that-actually-convert-2026)
- [How Top AI Tools Onboard New Users in 2026 — UserGuiding](https://userguiding.com/blog/how-top-ai-tools-onboard-new-users)
- [SaaS Onboarding in 2026: Activation Checklist — DAR Design](https://dardesign.io/blog/saas-onboarding-2026-activation-checklist-reduce-churn)
- [Arcads Review — Marketer Milk](https://www.marketermilk.com/blog/arcads-review) *(bare-bones onboarding gap)*
- [Digital Marketing Workflows 2026 — Octopus Marketing](https://www.octopusmarketing.agency/blog/digital-marketing-workflows-process-documentation-2026-the-ultimate-scaling-blueprint-for-teams-agencies/)
