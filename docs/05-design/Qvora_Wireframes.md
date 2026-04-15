# QVORA
## Wireframes — Screen Layout Brief
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Lo-Fi Draft
**Coverage:** 12 key screens across all user journey phases

---

## Screen Index

| # | Screen | Phase | Priority |
|---|---|---|---|
| S-01 | Marketing / Landing Page | Acquisition | P0 |
| S-02 | Signup | Onboarding | P0 |
| S-03 | Role & Context | Onboarding | P0 |
| S-04 | Brand Setup | Onboarding | P0 |
| S-05 | Dashboard — Empty State | Activation | P0 |
| S-06 | Dashboard — Populated | Retention | P0 |
| S-07 | Brief Generator — URL Input | Core flow | P0 |
| S-08 | Brief View & Edit | Core flow | P0 |
| S-09 | Generation Settings | Core flow | P0 |
| S-10 | Generation Progress | Core flow | P0 |
| S-11 | Asset Library | Core flow | P0 |
| S-12 | Export Panel | Core flow | P0 |
| S-13 | Trial Locked State | Conversion | P0 |
| S-14 | Plan Limit / Upgrade | Conversion | P0 |
| S-15 | Signal Dashboard (V2) | Expansion | P1 |

---

## App Shell (persistent across S-05 to S-13)

```
┌────────────────────────────────────────────────────────────────────┐
│ TOPBAR                                                    64px     │
│ [◉ QVORA]   [Acme Co ▼]   ──────────────  [16/20 ads] [+ New] [👤]│
├──────────┬─────────────────────────────────────────────────────────┤
│ SIDEBAR  │  CONTENT AREA                                           │
│ 240px    │  max-w-7xl · px-8 · py-8                               │
│          │                                                         │
│ [🏠] Dashboard         │                                          │
│ [✨] Briefs            │                                          │
│ [▶] Studio             │                                          │
│ [📦] Assets            │                                          │
│ [📊] Signal  [V2 pill] │                                          │
│          │                                                         │
│ ──────── │                                                         │
│ [🏷] Brands            │                                          │
│ [👥] Team              │                                          │
│ [⚙] Settings          │                                          │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Topbar notes:**
- Brand switcher `[Acme Co ▼]` — click opens dropdown of all brand kits
- Usage pill `[16/20 ads]` — turns red at 80%; click goes to billing
- `[+ New]` — primary CTA, always visible; opens brief creation
- Avatar `[👤]` — dropdown: profile, settings, logout

**Sidebar notes:**
- Active nav item: `bg-accent border-l-2 border-primary`
- Signal item shows `[V2]` badge until ad account connected
- Collapses to icon-only on `md` breakpoint; drawer on mobile

---

## S-01 — Marketing / Landing Page

**Goal:** Communicate the value proposition in one scroll. Get the signup click.

```
┌────────────────────────────────────────────────────────────────────┐
│  NAV   [◉ QVORA]  ──────────────────  [Pricing] [Login] [Try free]│
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  HERO                                     bg: #0A0A0F             │
│  ┌──────────────────────────────────┐                             │
│  │                                  │                             │
│  │  "Born to Convert."              │   [Product screenshot /     │
│  │                                  │    animated demo]           │
│  │  URL → Strategy → Video Ads      │                             │
│  │  in 15 minutes.                  │                             │
│  │                                  │                             │
│  │  [Try free — no card needed]     │                             │
│  │  [Watch 90s demo]                │                             │
│  │                                  │                             │
│  └──────────────────────────────────┘                             │
│                                                                    │
├────────────────────────────────────────────────────────────────────┤
│  SOCIAL PROOF BAR                                                  │
│  Trusted by [logo] [logo] [logo] [logo]  · "3x creative output"   │
├────────────────────────────────────────────────────────────────────┤
│  THE PROBLEM (3-column)                                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │
│  │ 3-day brief  │  │ 1-week edit  │  │ 1-week wait  │            │
│  │ cycle        │  │ cycle        │  │ for results  │            │
│  └──────────────┘  └──────────────┘  └──────────────┘            │
│  "By the time you know what worked, the moment is gone."          │
├────────────────────────────────────────────────────────────────────┤
│  HOW IT WORKS (numbered steps)                                     │
│  ①  Paste URL         ②  Get Strategy     ③  Generate Ads        │
│  ④  Export Test Set   ⑤  Signal Learns    ⑥  Next Brief Wins     │
├────────────────────────────────────────────────────────────────────┤
│  FEATURE BENTO GRID (2×2 + 1 wide)                                │
│  ┌────────────────────┐  ┌─────────┐  ┌─────────┐                │
│  │  Qvora Brief       │  │ Studio  │  │ Signal  │                │
│  │  (wide card)       │  │         │  │  (V2)   │                │
│  │                    │  └─────────┘  └─────────┘                │
│  └────────────────────┘                                           │
├────────────────────────────────────────────────────────────────────┤
│  PRICING (3 cards)                                                 │
│  [Starter $99] [Growth $149] [Scale $399]                         │
├────────────────────────────────────────────────────────────────────┤
│  FINAL CTA                                                         │
│  "Your competitor launched 40 creatives this week."               │
│  [Start free trial]                                               │
└────────────────────────────────────────────────────────────────────┘
```

**Annotations:**
- Hero bg: deep black `#0A0A0F`; headline: Clash Display 48px, white
- Product screenshot: animated — shows brief → video generation loop
- Problem section: each box has `Signal Red` top accent line
- Feature bento: cards use glassmorphism — `bg-card/60 backdrop-blur-md`
- CTA button: Volt `#7B2FFF` with glow on hover

---

## S-02 — Signup

**Goal:** Account created in < 90 seconds. Google OAuth is primary.

```
┌────────────────────────────────────────────────────────────────────┐
│  [◉ QVORA]                                    [Already have account?│
│                                                         Sign in →] │
├──────────────────────────────────┬─────────────────────────────────┤
│                                  │                                 │
│  FORM  (centered, max-w-sm)      │  SOCIAL PROOF                  │
│                                  │                                 │
│  "Start your free 7-day trial"   │  "Join agencies producing       │
│  No credit card required.        │   3x more creative."           │
│                                  │                                 │
│  [G  Continue with Google]       │  ┌─────────────────────┐       │
│                                  │  │  "Cut brief-to-       │       │
│  ── or ──                        │  │   launch from 5 days │       │
│                                  │  │   to 45 minutes."   │       │
│  Work email  [________________]  │  │   — Agency CD       │       │
│  Password    [________________]  │  └─────────────────────┘       │
│                                  │                                 │
│  [  Create account  ]            │  ┌─────────────────────┐       │
│                                  │  │  "3x output at the   │       │
│  By signing up you agree to      │  │   same retainer."   │       │
│  Terms · Privacy                 │  │   — Media Buyer     │       │
│                                  │  └─────────────────────┘       │
└──────────────────────────────────┴─────────────────────────────────┘
```

**Annotations:**
- Google button: white bg, `#0A0A0F` text, Google logo — highest conversion
- No name field on signup — collected on next screen
- Right panel: testimonials rotate every 4s (Framer Motion auto-play)
- Mobile: right panel hidden; form full-width

---

## S-03 — Role & Context

**Goal:** Role selection in one screen. 3 fields max.

```
┌────────────────────────────────────────────────────────────────────┐
│  [◉ QVORA]                              Step 1 of 2  [●○]         │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  "Quick setup — takes 60 seconds"                                  │
│                                                                    │
│  What describes you best?                                          │
│  ┌──────────────────────────────────────────────────────┐         │
│  │  ○  Media Buyer / Performance Marketer               │         │
│  │  ○  Creative Director                                │         │
│  │  ○  Agency Owner / Account Manager                   │         │
│  │  ○  In-house Brand Manager                           │         │
│  └──────────────────────────────────────────────────────┘         │
│                                                                    │
│  Monthly ad spend you manage                                       │
│  ┌──────────────────────────────────────────────────────┐         │
│  │  ○  < $10K    ○  $10K–$100K                          │         │
│  │  ○  $100K–$500K    ○  > $500K                        │         │
│  └──────────────────────────────────────────────────────┘         │
│                                                                    │
│  Company / Agency name   [________________________________]        │
│                                                                    │
│                                     [  Continue →  ]              │
└────────────────────────────────────────────────────────────────────┘
```

**Annotations:**
- Radio cards: full-width clickable rows — not tiny radio buttons
- Selected row: `bg-accent border-l-2 border-primary`
- "Continue" disabled until role selected
- Progress indicator top-right: `[●●○]` — 2 of 3 steps

---

## S-04 — Brand Setup

**Goal:** Minimum viable brand kit in < 2 minutes.

```
┌────────────────────────────────────────────────────────────────────┐
│  [◉ QVORA]                              Step 2 of 2  [●●]         │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  "Set up your first brand"                                         │
│  You can add more detail in Brand Settings later.                  │
│                                                                    │
│  Brand name   [________________________________]                   │
│                                                                    │
│  Primary color                                                     │
│  [#7B2FFF]  ████  (colour picker popover)                         │
│                                                                    │
│  Logo                                                              │
│  ┌──────────────────────────────────┐                             │
│  │   [↑]  Upload PNG or SVG         │                             │
│  │        Drag and drop here        │  ← dashed border, dropzone  │
│  └──────────────────────────────────┘                             │
│  [Skip for now →]  ← text link, muted                             │
│                                                                    │
│  ────────────────────────────────────────────────────────         │
│                                                                    │
│  [  Create brand & continue →  ]                                   │
│                                                                    │
│  ⓘ You can add fonts, colours, and intro/outro in                  │
│     Brand Settings anytime.                                        │
└────────────────────────────────────────────────────────────────────┘
```

**Annotations:**
- Logo upload: drag-and-drop with dashed `border-dashed border-border` zone
- "Skip for now" is visible — don't block activation on logo
- Color picker: Radix popover with hex input + visual swatch
- On submit: brief creation screen immediately (no dashboard redirect)

---

## S-05 — Dashboard (Empty State)

**Goal:** New user's first dashboard — direct them to action, not data.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  DASHBOARD                                              │
│          │                                                         │
│          │  Good morning, Jordan. 👋                               │
│          │                                                         │
│          │  ┌────────────────────────────────────────────────┐    │
│          │  │  ONBOARDING CHECKLIST                          │    │
│          │  │  ●  Brand created           ✓                  │    │
│          │  │  ○  Generate your first brief                  │    │
│          │  │  ○  Export your first ad set                   │    │
│          │  │  ○  Connect ad account (optional)              │    │
│          │  │                 ██████░░░░  50% complete        │    │
│          │  └────────────────────────────────────────────────┘    │
│          │                                                         │
│          │  ┌────────────────────────────────────────────────┐    │
│          │  │                                                │    │
│          │  │    [✨]                                        │    │
│          │  │                                                │    │
│          │  │   Paste your product URL.                      │    │
│          │  │   Get your brief in 15 seconds.               │    │
│          │  │                                                │    │
│          │  │   [  + New brief  ]                           │    │
│          │  │                                                │    │
│          │  └────────────────────────────────────────────────┘    │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- Checklist: shows for first 7 days only; dismisses on completion
- Empty state hero: centered, icon `Wand2` size-12, muted violet tint bg
- `[+ New brief]` button: Volt primary, glow effect — unmissable
- No metrics cards on empty state — only shown when data exists

---

## S-06 — Dashboard (Populated)

**Goal:** Returning user sees progress at a glance. Quick resume.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  DASHBOARD                                              │
│          │                                                         │
│          │  Good morning, Jordan.            [+ New brief]         │
│          │                                                         │
│          │  BENTO METRICS  (4-col grid)                           │
│          │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌────────┐ │
│          │  │ Ads this  │ │ Briefs    │ │ Brands    │ │ Ads    │ │
│          │  │ month     │ │ active    │ │ active    │ │ used   │ │
│          │  │  12       │ │  3        │ │  2        │ │ 12/20  │ │
│          │  │ ↑4 vs last│ │           │ │           │ │ ██░░░  │ │
│          │  └───────────┘ └───────────┘ └───────────┘ └────────┘ │
│          │                                                         │
│          │  RECENT BRIEFS                           [View all →]  │
│          │  ┌──────────────────────┐ ┌──────────────────────┐    │
│          │  │ Acme Protein Bar     │ │ Acme Energy Drink    │    │
│          │  │ 3 angles · 9 ads     │ │ Draft · 0 ads        │    │
│          │  │ Exported Apr 10      │ │ [Continue →]         │    │
│          │  └──────────────────────┘ └──────────────────────┘    │
│          │                                                         │
│          │  RECENT ASSETS (3-col grid)              [View all →]  │
│          │  ┌──────────┐ ┌──────────┐ ┌──────────┐               │
│          │  │ [▶ vid]  │ │ [▶ vid]  │ │ [▶ vid]  │               │
│          │  │ Angle 1  │ │ Angle 2  │ │ Angle 3  │               │
│          │  └──────────┘ └──────────┘ └──────────┘               │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- Metrics row: bento grid — `12/20 ads` bar turns `--destructive` at 80%
- Brief cards: `hover:border-primary/30` transition 150ms
- Draft brief: different visual treatment — muted border + "Continue" CTA
- Asset previews: Mux Player thumbnail + play overlay on hover

---

## S-07 — Brief Generator (URL Input)

**Goal:** URL submitted in one action. No navigation required.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  NEW BRIEF                                              │
│          │                                                         │
│          │  ← Back          New Brief · Acme Co                   │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │                                                  │  │
│          │  │  Paste your product URL                          │  │
│          │  │                                                  │  │
│          │  │  [ https://yourproduct.com/page      ] [→ Go]   │  │
│          │  │                                                  │  │
│          │  │  ── or ──                                        │  │
│          │  │                                                  │  │
│          │  │  [  Paste a product brief instead  ]             │  │
│          │  │                                                  │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │ 💡 Works with Shopify · WooCommerce · App Store  │  │
│          │  │    Google Play · Custom landing pages            │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  Industry template (optional)                           │
│          │  [Auto-detect ▼]  DTC · Mobile App · SaaS · Beauty...  │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Loading state (after submit):**
```
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  [⠿]  Reading your product page...          3s / 10s    │  │
│  │       ████████░░░░░░░░░░  40%                            │  │
│  │       acme.com/protein-bar                               │  │
│  └──────────────────────────────────────────────────────────┘  │
```

**Annotations:**
- Input: large, full-width, `text-base` — feels like a search bar, not a form
- `[→ Go]` button: Volt, attached to input right edge (search-bar pattern)
- Industry template: optional, collapsed by default
- Loading: inline progress below input — no page change

---

## S-08 — Brief View & Edit

**Goal:** Full brief visible. Every part editable. One click to generate.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  BRIEF — Acme Protein Bar                               │
│          │  acme.com/protein-bar · Generated 12s ago   [Share ↗]  │
│          │                                                         │
│          │  PRODUCT SUMMARY                     [Edit]             │
│          │  Clean-label 25g protein bar for active consumers.      │
│          │  Key differentiator: real food, no artificial sweet.    │
│          │                                                         │
│          │  ANGLES  (3 of 5)                    [+ Add angle]      │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │ [CONVERSION] ● Desire    ↕ drag  [↺] [✎] [✕]   │  │
│          │  │ "The bar that doesn't taste like compromise"     │  │
│          │  │ Funnel: Conversion · Emotion: Aspiration         │  │
│          │  │                                                  │  │
│          │  │ H1 [desire]    "Real food. 25g protein..."      │  │
│          │  │ H2 [problem]   "Tired of protein bars that..."  │  │
│          │  │ H3 [soc-proof] "10K athletes switched..."       │  │
│          │  │                        [+ Add hook] [↺ Regen]   │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │ [AWARENESS] ● Curiosity  ↕ drag  [↺] [✎] [✕]   │  │
│          │  │  ...                                             │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  RECOMMENDED FORMAT                                     │
│          │  ① UGC-style   ② Product Demo   ③ Text-Motion         │
│          │  Platforms: Meta Feed · TikTok · Stories               │
│          │                                                         │
│          │  ────────────────────────────────────────────────────  │
│          │  [  ← Edit  ]         [  Approve & Generate Videos →  ]│
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- Angle cards: draggable (drag handle `↕` on left) for reorder
- Inline edit: click angle name → becomes input; click outside → saves
- `[↺ Regen]` per angle and per hook section
- Hook rows: `[desire]` badge + opening line text
- Approved brief: border turns `border-success`, header badge `[✓ Approved]`
- Bottom bar: sticky on scroll — CTA always visible

---

## S-09 — Generation Settings

**Goal:** Format confirmed in < 30 seconds. Smart defaults pre-selected.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  GENERATE — Acme Protein Bar                            │
│          │  3 angles approved · 1 video per angle                  │
│          │                                                         │
│          │  ┌────── From Brief (Recommended) ───────────────────┐  │
│          │  │  [●] UGC-style    Avatar + voiceover              │  │
│          │  │  [ ] Product Demo  Product shots + B-roll         │  │
│          │  │  [ ] Text-Motion   Animated copy + image          │  │
│          │  └───────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  ┌────── AI Generation Modes ────────────────────────┐  │
│          │  │  [ ] Text → Video    Script → cinematic           │  │
│          │  │  [ ] Image → Video   Product image → motion       │  │
│          │  │  [ ] Voice → Video   Your voice → lip-sync        │  │
│          │  └───────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  Platform  [●] Meta Feed (1:1)  [ ] TikTok (9:16)      │
│          │            [ ] Both                                     │
│          │                                                         │
│          │  Duration  [ ] 15s  [●] 30s  [ ] 60s                   │
│          │                                                         │
│          │  Avatar    [  Select avatar  ▼  ] (100+ avatars)       │
│          │  Voice     [  Select voice   ▼  ] (20+ voices)         │
│          │                                                         │
│          │  Estimated time: ~4 minutes · 3 videos                 │
│          │                                                         │
│          │  [  ← Back to brief  ]    [  Generate 3 videos →  ]    │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- "From Brief" section expanded by default; "AI Generation Modes" collapsed
- Selecting an AI mode collapses "From Brief" and expands mode settings inline
- Avatar selector: popover grid (3×3 thumbnails + filter bar)
- Voice selector: list with play-preview button per voice
- Estimated time updates when platform or duration changes

---

## S-10 — Generation Progress

**Goal:** Keep user informed. Reduce drop-off during wait.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  GENERATING — Acme Protein Bar                          │
│          │  3 videos · UGC-style · 30s · Meta Feed                 │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │  ⏳ Angle 1 — Conversion (Desire)                │  │
│          │  │  ████████████░░░░░░░░  80%                       │  │← Volt bar
│          │  │  Rendering lip-sync...                           │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │  ⏳ Angle 2 — Awareness (Curiosity)              │  │
│          │  │  █████░░░░░░░░░░░░░░  40%                        │  │
│          │  │  Generating script...                            │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │  ⏳ Angle 3 — Consideration (Fear)               │  │
│          │  │  ██░░░░░░░░░░░░░░░░░░  15%                       │  │
│          │  │  Extracting product assets...                    │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  ⏱ ~3 minutes remaining                                 │
│          │                                                         │
│          │  [  Notify me by email when ready  ]                    │
│          │  You can navigate away safely.                          │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Complete state (single card):**
```
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  ✓ Angle 1 — Conversion (Desire)           Done in 1:42 │  │← green border
│  │  [▶ Preview]  [↓ Download]                              │  │
│  └──────────────────────────────────────────────────────────┘  │
```

**Annotations:**
- Progress bars: Volt fill, animated spring
- Status text updates in real-time via SSE
- Completed cards: border → `border-success`, icon → `CheckCircle2`
- When all complete: success banner + `[View ad set →]` CTA

---

## S-11 — Asset Library

**Goal:** Browse, preview, and select assets for export in one screen.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  ASSETS                              [+ Generate more]  │
│          │                                                         │
│          │  FILTERS                                                │
│          │  [Brand ▼]  [Format ▼]  [Angle ▼]  [Date ▼]  [Clear]  │
│          │                                                         │
│          │  9 assets · Acme Protein Bar · UGC · Apr 14            │
│          │                                                         │
│          │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│          │  │ [▶ 0:30] │  │ [▶ 0:30] │  │ [▶ 0:30] │             │
│          │  │          │  │          │  │          │             │
│          │  │ Angle 1  │  │ Angle 2  │  │ Angle 3  │             │
│          │  │ Desire   │  │ Curiosity│  │ Fear     │             │
│          │  │ UGC · 30s│  │ UGC · 30s│  │ UGC · 30s│             │
│          │  ├──────────┤  ├──────────┤  ├──────────┤             │
│          │  │ [↓][✎][⋯]│  │ [↓][✎][⋯]│  │ [↓][✎][⋯]│             │
│          │  └──────────┘  └──────────┘  └──────────┘             │
│          │                                                         │
│          │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│          │  │  ...     │  │  ...     │  │  ...     │             │
│          │  └──────────┘  └──────────┘  └──────────┘             │
│          │                                                         │
│          │  ☐ Select all                        [Export selected] │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Asset card hover state:**
```
┌──────────┐
│ [▶ PLAY] │  ← overlay with large play button
│  ● ● ●  │  ← bottom-left: select checkbox appears
│ [↓] [✎] │  ← bottom-right: quick actions appear
└──────────┘
```

**Annotations:**
- Filter bar: sticky below topbar on scroll
- Asset cards: checkbox appears on hover or when "Select all" clicked
- `[Export selected]` appears in sticky bottom bar when ≥1 selected
- Video previews: Mux Player thumbnail; full preview on click → modal

---

## S-12 — Export Panel (Sheet/Drawer)

**Goal:** Confirm export settings and download. Right-side sheet pattern.

```
┌──────────┬──────────────────────────────┬──────────────────────────┐
│ SIDEBAR  │  ASSET LIBRARY (dimmed)      │  EXPORT PANEL            │
│          │  (bg overlay 40% opacity)    │  ──────────────────────  │
│          │                              │  Export 3 assets         │
│          │                              │  Acme Protein Bar        │
│          │                              │                          │
│          │                              │  Format                  │
│          │                              │  [●] MP4 H.264  [ ] MOV  │
│          │                              │                          │
│          │                              │  Resolution              │
│          │                              │  [●] 1080p  [ ] 720p     │
│          │                              │                          │
│          │                              │  Naming convention       │
│          │                              │  ┌──────────────────┐   │
│          │                              │  │ acme-co_          │   │
│          │                              │  │ conversion_ugc_  │   │
│          │                              │  │ desire_          │   │
│          │                              │  │ 20260414_v1.mp4  │   │
│          │                              │  └──────────────────┘   │
│          │                              │  ✓ manifest.csv included │
│          │                              │                          │
│          │                              │  Est. size: 185 MB       │
│          │                              │                          │
│          │                              │  [  Download ZIP  ]      │
│          │                              │  Ready in ~20 seconds    │
│          │                              │                          │
│          │                              │  [  Cancel  ]            │
└──────────┴──────────────────────────────┴──────────────────────────┘
```

**Annotations:**
- Sheet slides in from right (Framer Motion `slideInRight`)
- Naming convention preview: monospace font, shows actual generated filenames
- `[Download ZIP]` → triggers async CDN bundle → progress bar replaces button
- On complete: `✓ Download ready` + auto-triggers browser download

---

## S-13 — Trial Locked State

**Goal:** Blocked user is motivated to upgrade, not confused. Assets feel safe.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  (full-page overlay — sidebar visible but inactive)     │
│ (muted)  │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │                                                  │  │
│          │  │  ⏳  Your trial has ended.                       │  │
│          │  │                                                  │  │
│          │  │  "Your 3 ad sets are saved and ready."           │  │
│          │  │  Upgrade to keep generating.                     │  │
│          │  │                                                  │  │
│          │  │  ┌────────────────────────────────────────────┐  │  │
│          │  │  │ ✅ Starter  $99/mo  · Best value           │  │  │
│          │  │  │    20 ads · 3 brands · full export         │  │  │
│          │  │  │    [ Upgrade to Starter → ]                │  │  │  ← Volt CTA
│          │  │  ├────────────────────────────────────────────┤  │  │
│          │  │  │  Growth  $149/mo  (expand ▼)               │  │  │
│          │  │  │  Scale   $399/mo  (expand ▼)               │  │  │
│          │  │  └────────────────────────────────────────────┘  │  │
│          │  │                                                  │  │
│          │  │  ⓘ Assets deleted after 30 days if no plan      │  │
│          │  │    selected.                                     │  │
│          │  │                                                  │  │
│          │  └──────────────────────────────────────────────────┘  │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- Full-page modal: `bg-background/95 backdrop-blur-sm` — content inaccessible behind
- Starter card: `border-primary bg-primary/5` with Volt glow CTA
- Growth and Scale: collapsed rows with expand chevron
- Asset retention note: `text-muted-foreground text-sm` — informative, not threatening
- No dismiss option — shown on every login until upgrade
- Triggered when `account.status === 'trial_expired'`

---

## S-14 — Plan Limit / Upgrade

**Goal:** User hits their ad cap — blocked inline with clear path to upgrade. Never cold-dumped to billing.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  GENERATE — Acme Protein Bar                            │
│          │                                                         │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │  ⚠  You've used all 20 ads this month.           │  │  ← red border
│          │  │                                                  │  │
│          │  │  Your Starter limit resets on May 1.             │  │
│          │  │  Upgrade to keep generating now.                 │  │
│          │  │                                                  │  │
│          │  │  ┌────────────────────────────────────────────┐  │  │
│          │  │  │  You're on Starter ($99/mo)                │  │  │
│          │  │  │  20 ads/mo  ████████████████████  20/20   │  │  │  ← red fill
│          │  │  ├────────────────────────────────────────────┤  │  │
│          │  │  │  Growth  $149/mo → 50 ads/mo               │  │  │
│          │  │  │  [ Upgrade to Growth → ]   ← Volt CTA      │  │  │
│          │  │  ├────────────────────────────────────────────┤  │  │
│          │  │  │  Scale   $399/mo → 200 ads/mo              │  │  │
│          │  │  └────────────────────────────────────────────┘  │  │
│          │  │                                                  │  │
│          │  │  [ Wait until next reset ]  ← ghost, muted      │  │
│          │  └──────────────────────────────────────────────────┘  │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- Triggered inline on the Generate screen — not a separate page navigation
- Limit banner: `border-destructive bg-destructive/5`
- Usage bar: fully red `bg-destructive` fill at 100%
- Current plan shown first for context; next tier is the CTA (Growth)
- Scale always visible — pricing anchoring toward Growth
- `[ Wait until next reset ]` ghost link — always provide an escape hatch
- State detected from `usage.ads_used >= plan.ads_limit`

---

## S-15 — Signal Dashboard (V2)

**Goal:** Show which creative variables are winning. Drive next brief.

```
┌──────────┬─────────────────────────────────────────────────────────┐
│ SIDEBAR  │  SIGNAL — Acme Co             Last sync: 2h ago  [↻]   │
│          │                                                         │
│          │  [Last 30 days ▼]    [Meta ▼]    [All formats ▼]       │
│          │                                                         │
│          │  KPI BENTO (4-col)                                      │
│          │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │
│          │  │ Avg CTR │ │ Best CPA│ │ Top     │ │ Fatigue │      │
│          │  │  3.2%   │ │  $4.10  │ │ Angle   │ │ Alerts  │      │
│          │  │ ↑0.4%   │ │ ↓$0.80  │ │ Convert │ │   2     │      │
│          │  └─────────┘ └─────────┘ └─────────┘ └─────────┘      │
│          │                                                         │
│          │  BY ANGLE TYPE                   BY HOOK TYPE           │
│          │  ┌──────────────────────┐  ┌──────────────────────┐    │
│          │  │ Conversion  3.8% CTR │  │ Desire   4.1% ▲ Best │    │
│          │  │ ████████████░░  Best │  │ Problem  2.8%        │    │
│          │  │ Awareness   1.9% CTR │  │ Soc Proo 2.1%        │    │
│          │  │ ████░░░░░░░░░░       │  │                      │    │
│          │  └──────────────────────┘  └──────────────────────┘    │
│          │                                                         │
│          │  ⚠ FATIGUE ALERTS                                       │
│          │  ┌──────────────────────────────────────────────────┐  │
│          │  │ ⚠ acme_conversion_ugc_desire_v1                  │  │
│          │  │   CTR dropped 35% over 3 days                    │  │
│          │  │   [  Refresh this ad →  ]                        │  │
│          │  └──────────────────────────────────────────────────┘  │
│          │                                                         │
│          │  [  Start next brief with Signal insights  ]            │
└──────────┴─────────────────────────────────────────────────────────┘
```

**Annotations:**
- KPI cards: metric in `font-mono text-2xl text-data`; delta in `text-success` / `text-destructive`
- Bar charts: relative width, Volt fill for top performer, muted for others
- Fatigue alert card: `border-destructive bg-destructive/5`
- `[Start next brief]` — Volt primary, bottom of page — drives the loop

---

*Wireframes v1.0 — Qvora*
*April 14, 2026 — Confidential*
