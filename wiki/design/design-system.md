---
title: Design System
category: design
tags: [design-system, colors, typography, tokens, dark-mode, tailwind-v4]
sources: [Qvora_Design-System]
updated: 2026-04-15
---

# Design System

## TL;DR
Dark-first, dense-but-breathable UI. All tokens live in `globals.css` as `@theme {}` — no JS config. shadcn/ui CSS variable convention mapped to Qvora brand. Transitions: 150–250ms.

---

## Design Philosophy

| Principle | Meaning in Practice |
|---|---|
| **Dark-first** | Dark is the primary experience; light is adapted from it. Agency users live in dark UIs (Meta Ads Manager, TikTok). |
| **Dense but breathable** | Information-rich without clutter. Whitespace generous inside components; tight between sections. |
| **Speed as a value** | Transitions 150–250ms. Skeletons appear immediately. Nothing waits without visual feedback. |
| **Intelligence visible** | Brief angles show rationale. Generation shows progress. Signal shows data — not arbitrary numbers. |

---

## Color Tokens

### shadcn CSS Variable Mapping (Dark Theme Default)

```css
/* globals.css — dark theme */
:root {
  --background:        9 9 15;        /* #0A0A0F — Qvora Black */
  --foreground:        245 245 247;   /* #F5F5F7 — Qvora White */
  --card:              15 15 22;      /* #0F0F16 — elevated surface */
  --card-foreground:   245 245 247;
  --popover:           18 18 28;      /* #12121C */
  --popover-foreground: 245 245 247;
  --primary:           123 47 255;    /* #7B2FFF — Volt */
  --primary-foreground: 245 245 247;
  --secondary:         26 26 38;      /* #1A1A26 */
  --secondary-foreground: 180 180 195;
  --muted:             26 26 38;
  --muted-foreground:  120 120 140;
  --accent:            35 35 55;      /* #232337 */
  --accent-foreground: 245 245 247;
  --destructive:       255 61 61;     /* #FF3D3D — Signal Red */
  --destructive-foreground: 245 245 247;
  --success:           0 232 122;     /* #00E87A — Convert Green */
  --success-foreground: 9 9 15;
  --data:              46 156 255;    /* #2E9CFF — Data Blue */
  --data-foreground:   9 9 15;
  --border:            35 35 55;      /* #232337 — subtle */
  --input:             26 26 38;
  --ring:              123 47 255;    /* Focus ring = Volt */
  --radius:            0.5rem;        /* 8px */
}
```

> ⚠️ **Rule:** All tokens in `@theme {}` block in `globals.css`. **No `tailwind.config.ts`** is allowed.

### Semantic Color Usage

| Token | Hex | Usage |
|---|---|---|
| `--primary` (Volt) | `#7B2FFF` | CTAs, active nav, focus rings, highlights |
| `--success` (Convert Green) | `#00E87A` | Performance wins, export complete, activation moments |
| `--destructive` (Signal Red) | `#FF3D3D` | Errors, delete, fatigue alerts, warnings |
| `--data` (Data Blue) | `#2E9CFF` | Analytics values, charts, metrics |
| `--background` (Black) | `#0A0A0F` | Main app background |
| `--foreground` (White) | `#F5F5F7` | Body text, default icon color |
| `--muted-foreground` | `#78788C` | Helper text, labels, secondary info |
| `--border` | `#232337` | All dividers, input borders, card outlines |

---

## Typography

| Role | Font | Weight | Use |
|---|---|---|---|
| **Hero / Display** | Clash Display | Bold (700) | Landing hero, major section headers |
| **Display / Subheading** | Space Grotesk Bold | Bold (700) | Dashboard section heads, card titles |
| **UI / Body** | Inter | Regular (400) / Medium (500) | Body text, labels, forms |
| **Data / Mono** | JetBrains Mono | Regular (400) | Metrics, IDs, code, API keys |

### Type Scale (Tailwind v4 CSS vars)

| Token | Size | Line Height | Tracking | Use |
|---|---|---|---|---|
| `--text-xs` | 12px | 1.5 | 0.025em | Badges, timestamps, helper text |
| `--text-sm` | 14px | 1.5 | 0 | Table data, form labels, body |
| `--text-base` | 16px | 1.6 | 0 | Primary body copy |
| `--text-lg` | 18px | 1.4 | -0.01em | Card headings, section labels |
| `--text-xl` | 20px | 1.3 | -0.02em | Page titles (dashboard), modal headers |
| `--text-2xl` | 24px | 1.2 | -0.025em | Section hero heads |
| `--text-3xl` | 30px | 1.1 | -0.03em | Major page headers |
| `--text-4xl+` | 36–72px | 1.0 | -0.04em | Marketing / landing hero |

---

## Spacing

- **Spacing unit:** 4px base
- **Page padding:** 24px horizontal (desktop), 16px (mobile)
- **Section gap:** 32px between major sections
- **Component gap:** 16px between related components, 8px within
- **Sidebar width (default):** 240px; collapsed: 64px (icon-only)
- **Topbar height:** 56px

---

## Motion

| Scenario | Duration | Easing |
|---|---|---|
| Micro-interactions (hover, focus) | 150ms | ease-out |
| Component transitions (dropdown, sheet) | 200ms | ease-in-out |
| Page transitions | 250ms | ease-in-out |
| Loading skeleton pulse | 1.5s | ease-in-out (loop) |
| Generation progress | real-time SSE driven | — |

**Rule:** Nothing should feel choppy or sluggish. 250ms max for any user-triggered transition.

---

## Surface Hierarchy

```
Background    #0A0A0F  (--background)
  └─ Card     #0F0F16  (--card)         ↑ 6 lightness units above bg
      └─ Popover  #12121C  (--popover)  ↑ elevated above card
          └─ Modal / Sheet overlay on top
```

Borders always `--border` (#232337). Never white borders.

---

## Open Questions
- [ ] Is there a sanctioned light-mode toggle in V1, or dark-only?
- [ ] Does the design system need a storybook or is Tailwind class documentation enough?
- [ ] Pattern for empty states — branded illustration or icon + copy only?

## Related Pages
- [[ui-components]] — component specs built on this system
- [[wireframes]] — screen layouts using these tokens
- [[stack-overview]] — Tailwind v4 rule reference
