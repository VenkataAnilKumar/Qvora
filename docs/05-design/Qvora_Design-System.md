# QVORA
## Design System
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Foundation Draft
**Stack:** Next.js 15 · shadcn/ui · Tailwind CSS v4 · Framer Motion

---

## Design Philosophy

**Dark-first.** The dark theme is designed as the primary experience. The light theme is adapted from it — not the other way around. Performance marketers working in Meta Ads Manager and TikTok Ads live in dark UIs. Qvora matches their environment.

**Dense but breathable.** Information-rich without feeling cluttered. Data is grouped, not scattered. Whitespace is generous inside components; tight between sections.

**Speed as a value.** Interactions feel instant. Transitions are short (150–250ms). Skeletons appear immediately. Nothing makes the user wait without visual feedback.

**Intelligence visible.** AI outputs feel considered, not generated. Brief angles show rationale. Generation shows progress. Signal shows data — not just numbers.

---

## Color Tokens

### Mapping — Brand → shadcn CSS Variables

shadcn/ui uses CSS variable convention: `--background`, `--foreground`, `--primary`, etc. Qvora brand colors map as follows.

```css
/* globals.css — dark theme (default) */
:root {
  /* Base surfaces */
  --background:        9 9 15;        /* #0A0A0F — Qvora Black */
  --foreground:        245 245 247;   /* #F5F5F7 — Qvora White */

  /* Card / elevated surface */
  --card:              15 15 22;      /* #0F0F16 — slightly lifted from bg */
  --card-foreground:   245 245 247;

  /* Popover / dropdown */
  --popover:           18 18 28;      /* #12121C */
  --popover-foreground: 245 245 247;

  /* Primary — Qvora Volt */
  --primary:           123 47 255;    /* #7B2FFF */
  --primary-foreground: 245 245 247;

  /* Secondary — subtle surface */
  --secondary:         26 26 38;      /* #1A1A26 */
  --secondary-foreground: 180 180 195;

  /* Muted — de-emphasized text / surfaces */
  --muted:             26 26 38;
  --muted-foreground:  120 120 140;

  /* Accent — used for hover states */
  --accent:            35 35 55;      /* #232337 */
  --accent-foreground: 245 245 247;

  /* Destructive — Signal Red */
  --destructive:       255 61 61;     /* #FF3D3D */
  --destructive-foreground: 245 245 247;

  /* Success — Convert Green */
  --success:           0 232 122;     /* #00E87A */
  --success-foreground: 9 9 15;

  /* Data — Data Blue */
  --data:              46 156 255;    /* #2E9CFF */
  --data-foreground:   9 9 15;

  /* Border */
  --border:            35 35 55;      /* #232337 — subtle, low-contrast */
  --input:             26 26 38;
  --ring:              123 47 255;    /* Focus ring = Volt */

  /* Radius */
  --radius:            0.5rem;        /* 8px default */
}

/* Light theme override */
.light {
  --background:        255 255 255;
  --foreground:        9 9 15;
  --card:              250 250 252;
  --card-foreground:   9 9 15;
  --popover:           255 255 255;
  --popover-foreground: 9 9 15;
  --primary:           123 47 255;
  --primary-foreground: 245 245 247;
  --secondary:         240 240 245;
  --secondary-foreground: 50 50 65;
  --muted:             240 240 245;
  --muted-foreground:  100 100 120;
  --accent:            235 235 245;
  --accent-foreground: 9 9 15;
  --destructive:       220 38 38;
  --destructive-foreground: 255 255 255;
  --success:           0 180 90;
  --success-foreground: 255 255 255;
  --data:              37 99 235;
  --data-foreground:   255 255 255;
  --border:            220 220 235;
  --input:             240 240 245;
  --ring:              123 47 255;
}

/* Tailwind v4 @theme block — registers CSS vars as utility classes */
/* Add to globals.css alongside :root and .light blocks */
@theme {
  /* Font families */
  --font-display: "Clash Display", sans-serif;
  --font-heading: "Space Grotesk", sans-serif;
  --font-sans: "Inter", sans-serif;
  --font-mono: "JetBrains Mono", monospace;

  /* Qvora extension colors — enables bg-success, text-success, etc. */
  --color-success:            rgb(var(--success) / <alpha-value>);
  --color-success-foreground: rgb(var(--success-foreground) / <alpha-value>);
  --color-data:               rgb(var(--data) / <alpha-value>);
  --color-data-foreground:    rgb(var(--data-foreground) / <alpha-value>);
}
```

> **Note — Tailwind v4 is CSS-first.** Fonts and custom tokens are declared in `globals.css` inside `@theme {}`, not in a `tailwind.config.ts`. There is no JS config file in a Tailwind v4 project. The `@theme` block above replaces the old `extend.colors` and `fontFamily` config options entirely.

### Semantic Color Usage

| Token | Hex | Use |
|---|---|---|
| `--background` | `#0A0A0F` | Page background, full-bleed sections |
| `--card` | `#0F0F16` | Card surfaces, panels, sidebar bg |
| `--secondary` | `#1A1A26` | Input backgrounds, tag pills, secondary buttons |
| `--accent` | `#232337` | Hover states, selected rows, active nav items |
| `--border` | `#232337` | All dividers, card borders, input borders |
| `--primary` (Volt) | `#7B2FFF` | CTAs, active states, brand highlights, progress fills |
| `--success` (Green) | `#00E87A` | Generation complete, export success, KPI wins, CTR positive |
| `--destructive` (Red) | `#FF3D3D` | Fatigue alerts, errors, warnings, plan limit |
| `--data` (Blue) | `#2E9CFF` | Charts, analytics values, Signal data, metrics |
| `--muted-foreground` | `#78788C` | Placeholder text, secondary labels, timestamps |

### Do Not Use
- Raw hex values in components — always use CSS variables via Tailwind `bg-background`, `text-foreground` etc.
- Accent colors decoratively — Convert Green, Signal Red, and Data Blue are **functional only**
- More than 2 accent colors on a single screen

---

## Typography

### Font Stack

| Role | Font | Weight | Tailwind Class |
|---|---|---|---|
| **Display / Hero** | Clash Display | 700 (Bold) | `font-display font-bold` |
| **Heading** | Space Grotesk | 600 (SemiBold) | `font-heading font-semibold` |
| **UI / Body** | Inter | 400, 500 | `font-sans` |
| **Data / Metrics** | JetBrains Mono | 400, 500 | `font-mono` |

```css
/* globals.css — Tailwind v4 @theme block (CSS-first, no tailwind.config.ts) */
@theme {
  --font-display: "Clash Display", sans-serif;
  --font-heading: "Space Grotesk", sans-serif;
  --font-sans: "Inter", sans-serif;
  --font-mono: "JetBrains Mono", monospace;
}
```

### Type Scale

| Name | Size | Line Height | Use |
|---|---|---|---|
| `text-4xl` | 36px | 1.1 | Hero headings (marketing) |
| `text-3xl` | 30px | 1.2 | Page titles |
| `text-2xl` | 24px | 1.3 | Section headings |
| `text-xl` | 20px | 1.4 | Card headings, modal titles |
| `text-lg` | 18px | 1.5 | Sub-headings, emphasis |
| `text-base` | 16px | 1.6 | Body copy, descriptions |
| `text-sm` | 14px | 1.5 | Labels, secondary text, table rows |
| `text-xs` | 12px | 1.4 | Timestamps, badges, metadata |

### Text Color Usage

| Color | Class | Use |
|---|---|---|
| `--foreground` | `text-foreground` | Primary content, headings |
| `--muted-foreground` | `text-muted-foreground` | Secondary labels, hints, timestamps |
| `--primary` | `text-primary` | Links, active states, volt highlights |
| `--success` | `text-success` | Positive metrics, success messages |
| `--destructive` | `text-destructive` | Errors, warnings, fatigue alerts |
| `--data` | `text-data` | Metric values in charts, analytics |

---

## Spacing Scale

Base unit: **4px (0.25rem)**. Tailwind's default spacing scale maps directly.

| Token | px | Use |
|---|---|---|
| `space-1` | 4px | Icon-to-label gap, tight inline spacing |
| `space-2` | 8px | Badge padding, button icon gap |
| `space-3` | 12px | Input padding (y), list item gap |
| `space-4` | 16px | Card padding (small), section gap (small) |
| `space-5` | 20px | — |
| `space-6` | 24px | Card padding (default), modal padding |
| `space-8` | 32px | Section spacing, between card groups |
| `space-10` | 40px | Page top padding, major section gap |
| `space-12` | 48px | Hero padding |
| `space-16` | 64px | Full-bleed section margins |

---

## Border Radius

| Token | Value | Use |
|---|---|---|
| `rounded-sm` | 4px | Badges, tags, small pills |
| `rounded` / `rounded-md` | 8px | Buttons, inputs, cards (default) |
| `rounded-lg` | 12px | Modals, large panels, video previews |
| `rounded-xl` | 16px | Feature cards on marketing page |
| `rounded-full` | 9999px | Avatar circles, toggle switches |

---

## Shadows & Elevation

Dark backgrounds don't take standard drop shadows well. Elevation is expressed through:

1. **Border highlight** — subtle top-border `border-t border-white/5` on elevated surfaces
2. **Background step-up** — each elevation level lightens bg slightly (`#0A0A0F` → `#0F0F16` → `#12121C`)
3. **Glow (Volt)** — Volt glow used sparingly on primary CTAs and active states

```css
/* Elevation levels */
--surface-0: #0A0A0F   /* page background */
--surface-1: #0F0F16   /* cards, panels */
--surface-2: #12121C   /* popovers, dropdowns */
--surface-3: #161624   /* tooltips, top-layer modals */

/* Volt glow — primary CTA only */
box-shadow: 0 0 20px rgba(123, 47, 255, 0.3);

/* Subtle border highlight */
border-top: 1px solid rgba(255, 255, 255, 0.05);
```

---

## Motion & Animation

**Principle:** Fast, purposeful, never decorative. Every animation communicates state — not personality.

| Type | Duration | Easing | Use |
|---|---|---|---|
| **Micro** | 100–150ms | `ease-out` | Button hover, checkbox, toggle |
| **Standard** | 200–250ms | `ease-in-out` | Panel slide, modal open, tab switch |
| **Complex** | 300–400ms | `spring` (Framer) | Generation progress, card reveal |
| **Skeleton** | `pulse` loop | — | Loading states — all skeletons use pulse |

### Framer Motion presets

```ts
// Fade in (cards, content areas)
export const fadeIn = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.2, ease: 'easeOut' }
}

// Slide in from right (panels, drawers)
export const slideInRight = {
  initial: { opacity: 0, x: 24 },
  animate: { opacity: 1, x: 0 },
  transition: { duration: 0.25, ease: 'easeOut' }
}

// Scale up (modals)
export const scaleUp = {
  initial: { opacity: 0, scale: 0.96 },
  animate: { opacity: 1, scale: 1 },
  transition: { duration: 0.2, ease: 'easeOut' }
}

// Generation progress (spring feel)
export const progressReveal = {
  initial: { width: '0%' },
  animate: { width: targetPercent },
  transition: { type: 'spring', stiffness: 80, damping: 20 }
}
```

---

## Layout System

### App Shell

```
┌──────────────────────────────────────────────────────────────┐
│  TOPBAR  64px                                               │
│  [◉ QVORA]  [Brand: Acme Co ▼]        [Usage] [+ New] [Avatar]│
├────────────┬─────────────────────────────────────────────────┤
│            │                                                 │
│  SIDEBAR   │  MAIN CONTENT AREA                              │
│  240px     │  max-w-6xl, centered, px-8                      │
│  (collaps- │                                                 │
│   ible to  │                                                 │
│   60px)    │                                                 │
│            │                                                 │
└────────────┴─────────────────────────────────────────────────┘
```

### Sidebar widths
| State | Width |
|---|---|
| Expanded | 240px |
| Collapsed (icon-only) | 60px |
| Mobile drawer | Full-screen overlay |

### Content max-widths
| Area | Max Width |
|---|---|
| Dashboard / library | `max-w-7xl` (1280px) |
| Brief editor | `max-w-4xl` (896px) |
| Settings | `max-w-2xl` (672px) |
| Onboarding wizard | `max-w-lg` (512px) |

### Grid system
- **Bento grid** for dashboard metrics: `grid-cols-4` (desktop) → `grid-cols-2` (tablet) → `grid-cols-1` (mobile)
- **Asset library grid**: `grid-cols-3` (desktop) → `grid-cols-2` (tablet) → `grid-cols-1` (mobile)
- **Brief angles**: single column, full-width cards stacked vertically

---

## Iconography

**Library:** Lucide Icons (ships with shadcn/ui)

| Icon | Use |
|---|---|
| `Wand2` | Brief generation, AI actions |
| `Video` | Studio, generation |
| `BarChart2` | Signal, performance |
| `Layers` | Brand kits |
| `Users` | Team |
| `Download` | Export |
| `Sparkles` | AI-generated content indicator |
| `Zap` | Fast generation, batch |
| `Target` | Qvora logo concept (reticle) |
| `AlertTriangle` | Fatigue alert |
| `CheckCircle2` | Success states |
| `RefreshCw` | Regenerate |
| `Clock` | Generation in progress |

**Size standards:**
- `size-4` (16px) — inline with text, table actions
- `size-5` (20px) — button icons, nav items
- `size-6` (24px) — section headings, empty state icons
- `size-8` (32px) — feature icons in cards
- `size-12` (48px) — empty state hero icons

---

## Core Component Styles

### Button variants

| Variant | Background | Text | Use |
|---|---|---|---|
| `default` | `--primary` (#7B2FFF) | White | Primary CTA — "Generate", "Approve", "Export" |
| `secondary` | `--secondary` | `--foreground` | Secondary actions — "Edit", "Duplicate" |
| `outline` | Transparent | `--foreground` | Tertiary — "Cancel", "Skip" |
| `ghost` | Transparent | `--muted-foreground` | Nav items, icon buttons |
| `destructive` | `--destructive` | White | Delete, remove |
| `success` | `--success` | `--background` | Activation moments — "Ads ready" |

**Volt glow on primary button (hover state):**
```css
button[variant="default"]:hover {
  box-shadow: 0 0 16px rgba(123, 47, 255, 0.4);
}
```

### Badge / Tag variants

| Variant | Use |
|---|---|
| `default` (Volt outline) | Angle type labels |
| `secondary` | Format tags (ugc, demo) |
| `success` | CTR win, export complete |
| `destructive` | Fatigue detected |
| `outline` (Data Blue) | Platform tags (Meta, TikTok) |

### Input states

| State | Border | Background |
|---|---|---|
| Default | `--border` | `--input` |
| Focus | `--primary` (Volt) | `--input` |
| Error | `--destructive` | `--input` |
| Disabled | `--border/50` | `--muted/30` |

### Progress bar (generation)

```
Background: --secondary
Fill:        --primary (Volt) with animated gradient
             linear-gradient(90deg, #7B2FFF, #9B5FFF)
Height:      6px
Radius:      rounded-full
Animation:   progressReveal (spring)
```

---

## Qvora-Specific Patterns

### Brief Card
```
┌─────────────────────────────────────────────────────┐
│  [CONVERSION]  ● Desire          [↺ Regen] [✎ Edit] │  ← Header row: badge + actions
│                                                     │
│  "The bar that doesn't taste like compromise"       │  ← Angle name, text-lg font-semibold
│  Funnel: Conversion · Emotion: Aspiration           │  ← Meta row, text-xs text-muted-foreground
│                                                     │
│  Hooks                                              │
│  ├─ H1 [desire]     "Real food. 25g protein..."    │  ← Hook rows with type badge
│  ├─ H2 [problem]    "Tired of protein bars that..." │
│  └─ H3 [soc-proof]  "10K athletes switched..."     │
└─────────────────────────────────────────────────────┘
```
- Border: `border border-border`
- Background: `bg-card`
- Hover: `border-primary/30` with subtle Volt glow
- Selected/approved: `border-primary bg-primary/5`

### Asset Card (video tile)
```
┌──────────────┐
│  [▶ VIDEO ]  │  ← 16:9 or 9:16 preview — Mux Player
│   PREVIEW    │
├──────────────┤
│ Angle 1      │  ← text-sm font-medium
│ Desire · UGC │  ← badges, text-xs
│ Apr 14 · 30s │  ← timestamp + duration
├──────────────┤
│ [↓] [✎] [⋯] │  ← download, edit, more actions
└──────────────┘
```
- Aspect ratio: 9:16 card for TikTok/Stories assets; 1:1 for Feed
- Hover: overlay with play button centred + `scale(1.02)` transform

### Generation Progress Card
```
┌─────────────────────────────────────────────────────┐
│  ⏳ Angle 1 — Conversion (Desire)                   │
│  ████████████░░░░ 75%                               │  ← Volt progress bar
│  UGC · 30s · 9:16                                   │
└─────────────────────────────────────────────────────┘
```
- Complete state: border turns `--success`, icon → `CheckCircle2`
- Failed state: border turns `--destructive`, icon → `AlertTriangle`

### Signal Metric Row
```
┌─────────────────────────────────────────────────────┐
│  Desire hook          4.1% CTR   $3.90 CPA   ▲ Best │
│                       ████████████░░░░              │  ← relative bar chart
└─────────────────────────────────────────────────────┘
```
- CTR/CPA values in `font-mono text-data`
- `▲ Best` in `text-success`
- Fatigue row: `text-destructive` + `AlertTriangle` icon

---

## Responsive Breakpoints

| Breakpoint | Width | Layout change |
|---|---|---|
| `sm` | 640px | Single column, stacked cards |
| `md` | 768px | 2-column grids, sidebar collapses to drawer |
| `lg` | 1024px | 3-column asset grid, sidebar visible |
| `xl` | 1280px | Full 4-column bento dashboard |
| `2xl` | 1536px | Max content width — no wider layouts |

---

*Design System v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources:**
- [ShadCN UI in 2026 — DEV Community](https://dev.to/whoffagents/shadcn-ui-in-2026-the-component-library-that-changed-how-we-build-uis-296o)
- [UI Design Trends 2026 — Tubik Blog](https://blog.tubikstudio.com/ui-design-trends-2026/)
- [Best Dashboard Design Examples 2026 — Muzli](https://muz.li/blog/best-dashboard-design-examples-inspirations-for-2026/)
