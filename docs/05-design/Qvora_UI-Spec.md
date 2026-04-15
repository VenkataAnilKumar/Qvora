# QVORA
## UI Component Specification
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Draft
**Stack:** shadcn/ui · Radix UI · Tailwind CSS v4 · Framer Motion · Lucide Icons

---

## Overview

This document specifies every UI component used in Qvora — variants, states, props, interaction behaviour, and accessibility requirements. It is the design-to-engineering handoff companion to the Wireframes and Design System documents.

---

## Component Index

| # | Component | shadcn base | Qvora customisation |
|---|---|---|---|
| C-01 | Button | `Button` | Volt glow on primary; success variant added |
| C-02 | Input | `Input` | Search-bar variant; URL input |
| C-03 | Textarea | `Textarea` | Brief manual input |
| C-04 | Badge | `Badge` | Angle type, hook type, platform, status |
| C-05 | Card | `Card` | Brief card, asset card, metric card |
| C-06 | Progress | `Progress` | Generation progress; usage meter |
| C-07 | Sidebar | `Sidebar` | Collapsible; brand switcher in topbar |
| C-08 | Topbar | custom | Brand switcher, usage pill, new brief CTA |
| C-09 | Dialog / Modal | `Dialog` | Video preview, confirmation dialogs |
| C-10 | Sheet | `Sheet` | Export panel, settings drawers |
| C-11 | Dropdown Menu | `DropdownMenu` | Brand switcher, action menus |
| C-12 | Tabs | `Tabs` | Brief angles view, signal breakdowns |
| C-13 | Avatar Selector | custom | Popover grid, filter, preview |
| C-14 | Voice Selector | custom | List with audio preview |
| C-15 | Video Player | Mux Player | Asset preview, generation complete |
| C-16 | Skeleton | `Skeleton` | All loading states |
| C-17 | Toast | `Sonner` | Success, error, info notifications |
| C-18 | Colour Picker | custom (Radix) | Brand kit colour input |
| C-19 | File Upload | custom | Logo upload, voice upload, image upload |
| C-20 | Data Table | `DataTable` | Asset library, signal metrics |
| C-21 | Chart | Recharts via shadcn | Signal bar charts, usage line |
| C-22 | Command | `Command` | Global search (⌘K) |
| C-23 | Onboarding Checklist | custom | Progress checklist, first 7 days |
| C-24 | Empty State | custom | All empty screens |
| C-25 | Generation Status Card | custom | Per-video progress + complete state |

---

## C-01 — Button

### Variants

| Variant | Class pattern | Use |
|---|---|---|
| `default` | `bg-primary text-primary-foreground` | Primary CTAs — "Generate", "Approve", "Export" |
| `secondary` | `bg-secondary text-secondary-foreground` | Secondary — "Edit", "Duplicate", "Share" |
| `outline` | `border border-input bg-transparent` | Tertiary — "Cancel", "Back", "Skip" |
| `ghost` | `hover:bg-accent` | Nav items, icon-only toolbar buttons |
| `destructive` | `bg-destructive text-destructive-foreground` | Delete, remove, disconnect |
| `success` | `bg-success text-success-foreground` | Activation moments — "Ads ready", "Export complete" |

### Sizes

| Size | Height | Padding | Font | Use |
|---|---|---|---|---|
| `sm` | 32px | `px-3` | `text-xs` | Table row actions, badge buttons |
| `default` | 40px | `px-4` | `text-sm` | Standard UI actions |
| `lg` | 48px | `px-6` | `text-base` | Primary CTAs in forms, hero |
| `icon` | 40×40px | `p-2` | — | Icon-only buttons |

### States

```
default:   bg-primary
hover:     bg-primary/90  +  box-shadow: 0 0 16px rgba(123,47,255,0.4)
focus:     ring-2 ring-ring ring-offset-2
active:    scale(0.98) — 100ms spring
disabled:  opacity-50 cursor-not-allowed
loading:   icon replaced with Loader2 spinner (spin animation)
```

### Loading button pattern
```tsx
<Button disabled={isLoading}>
  {isLoading ? <Loader2 className="size-4 animate-spin mr-2" /> : null}
  {isLoading ? 'Generating...' : 'Generate videos'}
</Button>
```

### Accessibility
- `aria-disabled` when loading (not HTML `disabled` — keeps focusable for screen readers)
- Loading state announces "Loading" via `aria-live="polite"` region

---

## C-02 — Input

### Variants

| Variant | Use |
|---|---|
| `default` | Standard form inputs (brand name, email, etc.) |
| `search` | URL input — larger, icon on right, attached button |
| `inline-edit` | Brief angle/hook editing — borderless until focused |

### URL Input (search variant)

```tsx
<div className="flex items-center border border-input rounded-md bg-input
                focus-within:ring-2 focus-within:ring-ring">
  <Input
    className="border-0 bg-transparent flex-1 text-base"
    placeholder="https://yourproduct.com/page"
  />
  <Button size="sm" className="rounded-l-none m-1">
    <ArrowRight className="size-4" />
  </Button>
</div>
```

### Inline Edit (brief fields)

```
default:  no border, transparent bg, cursor text
hover:    bg-accent/40, border-b border-border dashed
focus:    border-b border-primary solid, bg-secondary
```

### States

| State | Visual |
|---|---|
| Default | `border-input bg-input` |
| Focus | `ring-2 ring-ring` (Volt) |
| Error | `border-destructive` + error message below `text-destructive text-xs` |
| Disabled | `opacity-50 cursor-not-allowed` |
| Success | `border-success` + `CheckCircle2` icon on right |

---

## C-03 — Textarea

Used for manual brief input (US-06). Same states as Input.

```tsx
<Textarea
  className="min-h-[120px] resize-y font-sans text-sm"
  placeholder="Describe your product — features, target audience, key differentiator, CTA..."
  maxLength={2000}
/>
<p className="text-xs text-muted-foreground text-right mt-1">
  {charCount}/2000
</p>
```

---

## C-04 — Badge

### Variants & colours

| Variant | Colour | Use |
|---|---|---|
| `angle-conversion` | Volt `#7B2FFF` outline | Conversion angle label |
| `angle-awareness` | Data Blue `#2E9CFF` outline | Awareness angle label |
| `angle-consideration` | `--muted-foreground` outline | Consideration angle |
| `angle-retention` | `--success` outline | Retention angle |
| `hook-desire` | Volt subtle fill | Desire hook type |
| `hook-problem` | Destructive subtle fill | Problem hook type |
| `hook-social-proof` | Success subtle fill | Social proof hook |
| `hook-shock` | Orange subtle fill | Shock hook |
| `hook-curiosity` | Blue subtle fill | Curiosity hook |
| `platform-meta` | `#0082FB` fill | Meta platform tag |
| `platform-tiktok` | `#010101` fill + white text | TikTok tag |
| `platform-youtube` | `#FF0000` fill | YouTube tag |
| `status-draft` | Muted outline | Draft state |
| `status-approved` | Success outline | Approved state |
| `status-exported` | Volt outline | Exported state |
| `v2-pill` | Volt subtle + italic | V2 feature label |

### Badge implementation

```tsx
// Angle type badge
<Badge variant="outline" className="text-primary border-primary/50 text-xs">
  CONVERSION
</Badge>

// Hook type badge
<Badge className="bg-primary/10 text-primary border-0 text-xs">
  desire
</Badge>

// V2 pill
<Badge className="bg-primary/10 text-primary text-xs italic">
  V2
</Badge>
```

---

## C-05 — Card

### Brief Card

```tsx
<Card className={cn(
  "border border-border bg-card transition-all duration-150",
  "hover:border-primary/30 hover:shadow-[0_0_12px_rgba(123,47,255,0.15)]",
  isApproved && "border-success bg-success/5",
  isSelected && "border-primary bg-primary/5"
)}>
  <CardHeader className="flex flex-row items-center justify-between pb-2">
    <div className="flex items-center gap-2">
      <Badge variant="angle-conversion">CONVERSION</Badge>
      <Badge variant="hook-desire">● Desire</Badge>
    </div>
    <div className="flex gap-1">
      <Button variant="ghost" size="icon"><RefreshCw className="size-4" /></Button>
      <Button variant="ghost" size="icon"><Pencil className="size-4" /></Button>
      <Button variant="ghost" size="icon"><GripVertical className="size-4 text-muted-foreground" /></Button>
    </div>
  </CardHeader>
  <CardContent>
    {/* Angle name — inline editable */}
    {/* Meta row — funnel stage + emotion */}
    {/* Hook list */}
  </CardContent>
</Card>
```

### Asset Card (video tile)

```tsx
<Card className="overflow-hidden border border-border bg-card group">
  {/* Video thumbnail — 16:9 or 9:16 aspect ratio */}
  <div className="relative aspect-video bg-secondary">
    <MuxPlayer ... />
    {/* Hover overlay */}
    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100
                    transition-opacity flex items-center justify-center">
      <Button size="icon" variant="ghost" className="text-white">
        <Play className="size-8" />
      </Button>
      <Checkbox className="absolute bottom-2 left-2 opacity-0 group-hover:opacity-100" />
    </div>
  </div>
  <CardContent className="p-3">
    <p className="text-sm font-medium truncate">Angle 1 — Conversion</p>
    <div className="flex gap-1 mt-1">
      <Badge variant="hook-desire">Desire</Badge>
      <Badge variant="secondary">UGC</Badge>
    </div>
    <p className="text-xs text-muted-foreground mt-1">Apr 14 · 30s · 1080p</p>
  </CardContent>
  <CardFooter className="p-3 pt-0 flex gap-1">
    <Button variant="ghost" size="sm"><Download className="size-4 mr-1" />Download</Button>
    <Button variant="ghost" size="sm"><Pencil className="size-4 mr-1" />Edit</Button>
    <Button variant="ghost" size="icon" className="ml-auto"><MoreHorizontal className="size-4" /></Button>
  </CardFooter>
</Card>
```

### Metric Card (dashboard bento)

```tsx
<Card className="border border-border bg-card p-6">
  <p className="text-xs text-muted-foreground uppercase tracking-wide">Ads this month</p>
  <p className="text-3xl font-mono font-semibold mt-1">12</p>
  <p className="text-xs text-success mt-1 flex items-center gap-1">
    <TrendingUp className="size-3" /> +4 vs last month
  </p>
</Card>
```

---

## C-06 — Progress

### Generation progress bar

```tsx
<div className="space-y-1">
  <div className="flex justify-between text-xs text-muted-foreground">
    <span>Rendering lip-sync...</span>
    <span>80%</span>
  </div>
  <Progress
    value={80}
    className="h-1.5 bg-secondary [&>div]:bg-gradient-to-r
               [&>div]:from-primary [&>div]:to-violet-400"
  />
</div>
```

### Usage meter (topbar pill)

```tsx
<div className={cn(
  "flex items-center gap-1.5 text-xs px-2 py-1 rounded-full border",
  usage >= 80
    ? "border-destructive/50 text-destructive bg-destructive/10"
    : "border-border text-muted-foreground"
)}>
  <span>{used}/{limit} ads</span>
</div>
```

---

## C-07 — Sidebar

### Structure

```
SidebarProvider (context)
└── Sidebar (240px / 60px collapsed)
    ├── SidebarHeader
    │   └── Logo + collapse toggle
    ├── SidebarContent
    │   ├── SidebarGroup — Main
    │   │   ├── SidebarMenuItem: Dashboard
    │   │   ├── SidebarMenuItem: Briefs
    │   │   ├── SidebarMenuItem: Studio
    │   │   ├── SidebarMenuItem: Assets
    │   │   └── SidebarMenuItem: Signal [V2 badge]
    │   └── SidebarGroup — Workspace
    │       ├── SidebarMenuItem: Brands
    │       ├── SidebarMenuItem: Team
    │       └── SidebarMenuItem: Settings
    └── SidebarFooter
        └── User avatar + plan badge
```

### Active state
```css
SidebarMenuItem[active]:
  background: hsl(var(--accent))
  border-left: 2px solid hsl(var(--primary))
  color: hsl(var(--foreground))
```

### Collapsed state (icon-only, 60px)
- Icons only, no labels
- Tooltip on hover shows label
- `SidebarTrigger` button at top toggles state, persisted in `localStorage`

---

## C-08 — Topbar

### Layout
```
[◉ Logo]  [Brand Switcher ▼]  ────────────  [Usage Pill]  [+ New Brief]  [Avatar ▼]
  48px         auto             flex-1          auto           auto          40px
```

### Brand Switcher
```tsx
<DropdownMenu>
  <DropdownMenuTrigger asChild>
    <Button variant="ghost" className="gap-2">
      <div className="size-4 rounded-sm" style={{ background: brand.color }} />
      <span className="text-sm font-medium">{brand.name}</span>
      <ChevronDown className="size-3 text-muted-foreground" />
    </Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    {brands.map(b => <DropdownMenuItem key={b.id}>{b.name}</DropdownMenuItem>)}
    <DropdownMenuSeparator />
    <DropdownMenuItem><Plus className="size-4 mr-2" />Add brand</DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

---

## C-09 — Dialog / Modal

### Video Preview Modal

```tsx
<Dialog>
  <DialogContent className="max-w-3xl p-0 overflow-hidden bg-card border-border">
    {/* Full video player — 16:9 */}
    <div className="aspect-video bg-black">
      <MuxPlayer playbackId={...} autoPlay />
    </div>
    <div className="p-4 flex items-center justify-between">
      <div>
        <p className="text-sm font-medium">{asset.angleName}</p>
        <div className="flex gap-1 mt-1">{/* badges */}</div>
      </div>
      <div className="flex gap-2">
        <Button variant="outline" size="sm"><Pencil className="size-4 mr-1" />Edit</Button>
        <Button size="sm"><Download className="size-4 mr-1" />Download</Button>
      </div>
    </div>
  </DialogContent>
</Dialog>
```

### Confirmation Dialog

```tsx
<AlertDialog>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete this asset?</AlertDialogTitle>
      <AlertDialogDescription>
        This cannot be undone. The video will be permanently removed.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <AlertDialogCancel>Cancel</AlertDialogCancel>
      <AlertDialogAction className="bg-destructive">Delete</AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
```

---

## C-10 — Sheet (Export Panel)

```tsx
<Sheet>
  <SheetContent side="right" className="w-[420px] bg-card border-l border-border">
    <SheetHeader>
      <SheetTitle>Export {count} assets</SheetTitle>
      <SheetDescription>{brand.name} · {brief.name}</SheetDescription>
    </SheetHeader>
    <div className="py-6 space-y-6">
      {/* Format selector */}
      {/* Resolution selector */}
      {/* Naming convention preview */}
      {/* manifest.csv checkbox */}
    </div>
    <SheetFooter>
      <Button className="w-full" size="lg" onClick={handleExport}>
        <Download className="size-4 mr-2" />Download ZIP
      </Button>
      <p className="text-xs text-muted-foreground text-center mt-2">
        Est. {sizeMB} MB · Ready in ~20 seconds
      </p>
    </SheetFooter>
  </SheetContent>
</Sheet>
```

---

## C-13 — Avatar Selector

### Popover grid pattern

```
┌─────────────────────────────────────────────────────┐
│  Select avatar                          [✕ Close]   │
│  [Search avatars...]                                 │
│  [Gender ▼] [Age ▼] [Style ▼]                       │
│  ────────────────────────────────────────────────── │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐     │
│  │ img  │ │ img  │ │ img  │ │ img  │ │ img  │     │
│  │      │ │      │ │  ✓  │ │      │ │      │     │← selected: ring-2 ring-primary
│  │ M·28 │ │ F·35 │ │ M·42 │ │ F·24 │ │ F·31 │     │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘     │
│  [Load more]                                        │
└─────────────────────────────────────────────────────┘
```

- Grid: `grid-cols-5 gap-2`
- Avatar thumbnail: `aspect-square rounded-lg overflow-hidden object-cover`
- Selected: `ring-2 ring-primary ring-offset-2 ring-offset-card`
- Hover: `scale(1.04)` 150ms

---

## C-14 — Voice Selector

```
┌─────────────────────────────────────────────────────┐
│  Select voice                                       │
│  [Search voices...]   [Gender ▼] [Energy ▼]         │
│  ────────────────────────────────────────────────── │
│  ○  Sarah K.   Female · 30s · Upbeat    [▶ Preview] │
│  ●  Marcus D.  Male · 40s · Authorit.  [▶ Preview] │← selected: bg-accent
│  ○  Priya N.   Female · 25s · Calm     [▶ Preview] │
│  ○  James T.   Male · 35s · Energetic  [▶ Preview] │
└─────────────────────────────────────────────────────┘
```

- `[▶ Preview]` plays a 5s audio sample inline
- Playing state: `[⏹ Stop]`, waveform animation shown
- Selected row: `bg-accent border-l-2 border-primary`

---

## C-16 — Skeleton

All loading states use pulse skeletons — not spinners (except button loading).

### Brief loading skeleton
```tsx
<div className="space-y-4">
  <Skeleton className="h-4 w-1/3" />      {/* Product summary title */}
  <Skeleton className="h-16 w-full" />    {/* Summary text */}
  <div className="space-y-3">
    {[1,2,3].map(i => (
      <Skeleton key={i} className="h-32 w-full rounded-lg" />  {/* Angle cards */}
    ))}
  </div>
</div>
```

### Asset grid loading skeleton
```tsx
<div className="grid grid-cols-3 gap-4">
  {[...Array(6)].map((_, i) => (
    <div key={i} className="space-y-2">
      <Skeleton className="aspect-video rounded-lg" />
      <Skeleton className="h-4 w-3/4" />
      <Skeleton className="h-3 w-1/2" />
    </div>
  ))}
</div>
```

---

## C-17 — Toast (Sonner)

```tsx
// Success
toast.success('Ad set ready', {
  description: '3 videos generated. Time to find your winner.',
  action: { label: 'View assets', onClick: () => router.push('/assets') }
})

// Error
toast.error("Couldn't reach that page", {
  description: 'Try pasting your product description instead.',
  action: { label: 'Manual input', onClick: openManualInput }
})

// Info
toast.info('Generation in progress', {
  description: "We'll notify you when your 3 videos are ready.",
})
```

**Styling:**
```tsx
<Toaster
  position="bottom-right"
  toastOptions={{
    classNames: {
      toast: 'bg-card border-border text-foreground',
      success: 'border-l-4 border-success',
      error: 'border-l-4 border-destructive',
    }
  }}
/>
```

---

## C-19 — File Upload

### Logo / image upload

```tsx
<div
  onDragOver={handleDragOver}
  onDrop={handleDrop}
  className={cn(
    "border-2 border-dashed border-border rounded-lg p-8",
    "flex flex-col items-center gap-3 cursor-pointer",
    "hover:border-primary/50 hover:bg-accent/20 transition-colors",
    isDragOver && "border-primary bg-primary/5"
  )}
>
  <Upload className="size-8 text-muted-foreground" />
  <div className="text-center">
    <p className="text-sm font-medium">Upload PNG or SVG</p>
    <p className="text-xs text-muted-foreground">or drag and drop</p>
  </div>
  <input type="file" className="hidden" accept=".png,.svg,.jpg,.webp" />
</div>
```

### Voice upload (V2V)

Same pattern but `accept=".mp3,.wav,.m4a"` and shows audio waveform preview after upload.

---

## C-23 — Onboarding Checklist

```tsx
<Card className="border border-border bg-card/60 backdrop-blur-sm">
  <CardHeader className="pb-2">
    <CardTitle className="text-sm flex items-center gap-2">
      <Sparkles className="size-4 text-primary" />
      Get started
    </CardTitle>
  </CardHeader>
  <CardContent className="space-y-2">
    {steps.map(step => (
      <div key={step.id} className="flex items-center gap-3">
        {step.complete
          ? <CheckCircle2 className="size-4 text-success shrink-0" />
          : <Circle className="size-4 text-muted-foreground shrink-0" />
        }
        <span className={cn(
          "text-sm",
          step.complete ? "text-muted-foreground line-through" : "text-foreground"
        )}>
          {step.label}
        </span>
      </div>
    ))}
    <Progress value={completedPercent} className="h-1 mt-3" />
  </CardContent>
</Card>
```

- Shown for first 7 days only (stored in user metadata)
- Auto-dismisses when 100% complete — `fadeOut` 300ms then `unmount`

---

## C-24 — Empty State

```tsx
<div className="flex flex-col items-center justify-center py-24 text-center">
  <div className="size-16 rounded-full bg-accent flex items-center justify-center mb-4">
    <icon className="size-8 text-muted-foreground" />
  </div>
  <h3 className="text-lg font-semibold">{title}</h3>
  <p className="text-sm text-muted-foreground mt-1 max-w-sm">{description}</p>
  <Button className="mt-6" size="lg">{cta}</Button>
</div>
```

### All empty states

| Screen | Icon | Title | Description | CTA |
|---|---|---|---|---|
| Brief Library | `Wand2` | No briefs yet | Paste your product URL and watch what happens. | Paste a URL |
| Asset Library | `Video` | No ads yet | Approve a brief to generate your first set. | Go to briefs |
| Signal (not connected) | `BarChart2` | Connect to start learning | See which angles and hooks are winning — automatically. | Connect Meta account |
| Signal (< 1K impressions) | `Clock` | Collecting data… | Signal unlocks after 1,000 impressions per variant. | View assets |
| Team | `Users` | Just you for now | Invite your team to collaborate on briefs and creatives. | Invite teammate |
| Brand Library | `Layers` | No brands yet | Set up your first client brand to get started. | Add brand |

---

## C-25 — Generation Status Card

```tsx
// In-progress
<Card className={cn(
  "border transition-all",
  status === 'complete' ? "border-success bg-success/5" :
  status === 'error'    ? "border-destructive bg-destructive/5" :
                          "border-border"
)}>
  <CardContent className="p-4 space-y-2">
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        {status === 'complete' ? <CheckCircle2 className="size-4 text-success" /> :
         status === 'error'    ? <AlertTriangle className="size-4 text-destructive" /> :
                                 <Loader2 className="size-4 animate-spin text-primary" />}
        <span className="text-sm font-medium">{angleName}</span>
      </div>
      {status === 'complete' && (
        <span className="text-xs text-muted-foreground">Done in {duration}</span>
      )}
    </div>

    {status === 'generating' && (
      <>
        <Progress value={progress} className="h-1.5" />
        <p className="text-xs text-muted-foreground">{statusMessage}</p>
      </>
    )}

    {status === 'complete' && (
      <div className="flex gap-2 mt-2">
        <Button variant="outline" size="sm">
          <Play className="size-4 mr-1" />Preview
        </Button>
        <Button variant="outline" size="sm">
          <Download className="size-4 mr-1" />Download
        </Button>
      </div>
    )}
  </CardContent>
</Card>
```

---

## Navigation Patterns

### Global keyboard shortcuts

| Shortcut | Action |
|---|---|
| `⌘K` | Open command palette (C-22 Command) |
| `⌘N` | New brief |
| `⌘E` | Export selected assets |
| `⌘/` | Toggle sidebar |
| `Esc` | Close modal / sheet |

### Command Palette (⌘K)

```
┌─────────────────────────────────────────────────────┐
│  [🔍  Search or jump to...]                         │
│  ────────────────────────────────────────────────── │
│  Recent                                             │
│  📄  Acme Protein Bar brief                         │
│  🎬  acme_conversion_ugc_desire_v1.mp4              │
│  ────────────────────────────────────────────────── │
│  Actions                                            │
│  ✨  New brief                          ⌘N          │
│  ▶   Generate videos                               │
│  📦  Go to Assets                                  │
│  ⚙   Brand Settings                               │
└─────────────────────────────────────────────────────┘
```

---

## Interaction Patterns

### Inline editing (brief fields)

1. User clicks on any brief text field
2. Field transitions: `border-b border-dashed` → `border-b border-primary` (150ms)
3. Text becomes editable input
4. Click outside OR press Enter → saves + transitions back
5. Auto-save indicator: `✓ Saved` appears for 1.5s top-right of card

### Drag-to-reorder (brief angles)

- `@dnd-kit/sortable` for accessible drag-and-drop
- Drag handle: `GripVertical` icon, cursor `grab`
- Active drag: card lifts (`shadow-lg scale(1.02)`) + placeholder shown
- Drop: spring animation snaps to position

### Optimistic updates

All mutations (approve brief, export, delete) update UI immediately before API confirms:
- Rollback on error + toast error message
- No loading spinners for fast actions (< 500ms)

---

## Accessibility Standards

| Requirement | Implementation |
|---|---|
| Focus management | All modals trap focus; return focus to trigger on close |
| Keyboard navigation | All interactive elements reachable via Tab; dropdowns via arrow keys |
| Screen reader | `aria-label` on all icon-only buttons; `aria-live` on generation progress |
| Colour contrast | All text meets WCAG AA (4.5:1 minimum) on dark background |
| Reduced motion | `@media (prefers-reduced-motion)` disables Framer Motion transitions |
| Touch targets | Minimum 44×44px for all interactive elements (mobile) |

---

*UI Spec v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources:**
- [ShadCN UI in 2026 — DEV Community](https://dev.to/whoffagents/shadcn-ui-in-2026-the-component-library-that-changed-how-we-build-uis-296o)
- [Build a Dashboard with shadcn/ui — DesignRevision](https://designrevision.com/blog/shadcn-dashboard-tutorial)
- [Empty States in UX — LogRocket](https://blog.logrocket.com/ux-design/empty-states-ux-examples/)
- [UI Design Trends 2026 — Midrocket](https://midrocket.com/en/guides/ui-design-trends-2026/)
