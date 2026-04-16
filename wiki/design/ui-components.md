---
title: UI Component Specification
category: design
tags: [ui-components, shadcn, radix, button, card, dialog, sidebar]
sources: [Qvora_UI-Spec]
updated: 2026-04-15
---

# UI Component Specification

## TL;DR
25 components built on shadcn/ui + Radix + Tailwind v4. Primary theme: dark-first with Volt glow on CTAs. Every component has defined hover/focus/loading/disabled states. Framer Motion for all motion.

---

## Component Index

| # | Component | Base | Key Customization |
|---|---|---|---|
| C-01 | Button | `Button` | Volt glow on primary; `success` variant added |
| C-02 | Input | `Input` | Search-bar variant; URL input |
| C-03 | Textarea | `Textarea` | Brief manual input |
| C-04 | Badge | `Badge` | Angle type, hook type, platform, status |
| C-05 | Card | `Card` | Brief card, asset card, metric card |
| C-06 | Progress | `Progress` | Generation progress; usage meter |
| C-07 | Sidebar | `Sidebar` | Collapsible; brand switcher in topbar |
| C-08 | Topbar | custom | Brand switcher + usage pill + "New Brief" CTA |
| C-09 | Dialog / Modal | `Dialog` | Video preview, confirmation dialogs |
| C-10 | Sheet | `Sheet` | Export panel, settings drawers |
| C-11 | Dropdown Menu | `DropdownMenu` | Brand switcher, action menus |
| C-12 | Tabs | `Tabs` | Brief angles view, signal breakdowns |
| C-13 | Avatar Selector | custom | Popover grid, filter, preview |
| C-14 | Voice Selector | custom | List with audio preview |
| C-15 | Video Player | Mux Player | Asset preview, generation complete |
| C-16 | Skeleton | `Skeleton` | All loading states |
| C-17 | Toast | `Sonner` | Success, error, info notifications |
| C-18 | Colour Picker | custom (Radix) | Brand kit color input |
| C-19 | File Upload | custom | Logo, voice, image upload |
| C-20 | Data Table | `DataTable` | Asset library, signal metrics |
| C-21 | Chart | Recharts (shadcn) | Signal bar charts, usage line |
| C-22 | Command | `Command` | Global search (⌘K) |
| C-23 | Onboarding Checklist | custom | Progress checklist, first 7 days |
| C-24 | Empty State | custom | All empty screens |
| C-25 | Generation Status Card | custom | Per-video progress + complete state |

---

## C-01 — Button

### Variants

| Variant | Background | Use |
|---|---|---|
| `default` | `bg-primary` (Volt) | Primary CTAs: Generate, Approve, Export |
| `secondary` | `bg-secondary` | Secondary: Edit, Duplicate, Share |
| `outline` | Transparent + border | Tertiary: Cancel, Back, Skip |
| `ghost` | `hover:bg-accent` | Nav items, icon-only toolbar |
| `destructive` | `bg-destructive` (Red) | Delete, remove, disconnect |
| `success` | `bg-success` (Green) | Activation: Ads ready, Export complete |

### Sizes

| Size | Height | Use |
|---|---|---|
| `sm` | 32px | Table row actions, badge buttons |
| `default` | 40px | Standard UI actions |
| `lg` | 48px | Primary CTAs in forms, hero |
| `icon` | 40×40px | Icon-only buttons |

### States

```
hover:    bg-primary/90  + box-shadow: 0 0 16px rgba(123,47,255,0.4)  ← Volt glow
focus:    ring-2 ring-ring ring-offset-2
active:   scale(0.98), 100ms spring
disabled: opacity-50 cursor-not-allowed
loading:  Loader2 spinner replaces icon
```

---

## C-04 — Badge

| Variant | Color | Use |
|---|---|---|
| `default` | Volt/purple tint | Angle type label |
| `secondary` | Muted gray | Hook variant label |
| `outline` | Border only | Platform (TikTok, Meta, etc.) |
| `success` | Convert Green | "Winning" status |
| `destructive` | Signal Red | "Fatiguing" / "Failed" status |
| `data` | Data Blue | Metrics annotation |

---

## C-05 — Card

Three variants in practice:

**Brief Card** — brief list view
- Thumbnail + brand logo + product name + angle count + status badge + timestamp
- Hover: border-primary glow

**Asset Card** — asset library
- Video thumbnail (Mux) + video duration + angle tag + platform badge + download button
- Generation Status Card (C-25) is a special variant during active generation

**Metric Card** — Signal dashboard (V2)
- Large metric number (JetBrains Mono) + label + trend indicator

---

## C-08 — Topbar

Right-to-left layout:
1. **Brand Switcher** (leftmost after logo) — shows active brand name + logo, dropdown to switch
2. **Usage Pill** — `X / Y videos this month` with progress bar inline; click → billing page
3. **New Brief** CTA (`default` button, `lg` size)
4. **User avatar** — dropdown with account settings, sign out

Trial users: **Trial banner** replaces or sits above topbar with countdown: `X days left in your trial → Upgrade`.

---

## C-15 — Video Player (Mux)

- `<MuxPlayer>` component from `@mux/mux-player-react`
- Signed playback tokens (workspace-scoped — no public URLs)
- Controls: play/pause, volume, fullscreen, download
- Autoplay on dialog open (muted)
- Aspect ratio: 9:16 (vertical) lockced — other ratios letterboxed

---

## C-24 — Empty State

Pattern for every zero-state screen:

```
[Icon — 48px, muted-foreground color]
[Heading — text-lg, foreground]
[Body copy — text-sm, muted-foreground]
[CTA Button — primary, "Get started"]
```

Specific copy examples:
- Briefs empty: *"Your first brief is one URL away."* → [Paste a URL]
- Assets empty: *"Generate your first video to see it here."*
- Exports empty: *"No exports yet. Export an asset to download your ads."*

---

## C-25 — Generation Status Card

Real-time state per video in generation queue:

```
[Thumbnail / spinner]  [Video title]
                       [Status badge: Queued / Rendering / Processing / Complete]
                       [Progress bar: 0–100%]
                       [ETA text: "~2 min remaining"]
```

Powered by SSE stream (`/api/generation/[jobId]/stream`). Updates in real time without polling.

---

## Open Questions
- [ ] Does C-13 (Avatar Selector) need a filter by gender / style in V1?
- [ ] C-14 (Voice Selector) — does audio preview play on hover or on click?
- [ ] Is C-22 (Command / ⌘K) in V1 or V2?

## Related Pages
- [[design-system]] — tokens and motions these components are built on
- [[wireframes]] — screen layouts assembling these components
- [[features]] — which components belong to which modules
