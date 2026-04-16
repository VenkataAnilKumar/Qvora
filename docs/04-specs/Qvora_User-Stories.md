# QVORA
## User Stories
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Draft

---

## Personas

| ID | Persona | Role | Context |
|---|---|---|---|
| **P1** | **Agency Media Buyer** | Performance marketer at a 5–50 person agency | Manages $50K–$500K/mo ad spend across 3–15 brand clients. Needs high creative volume fast. Lives in Meta Ads Manager and TikTok Ads Manager. |
| **P2** | **Agency Creative Director** | Senior creative at the same agency | Owns brand voice, creative strategy, and quality bar. Historically briefs copywriters and video editors. Now needs AI to match their strategic thinking, not replace it. |
| **P3** | **Agency Account Manager** | Client-facing PM at the agency | Manages client expectations, delivery timelines, and reporting. Needs visibility into what's being produced and when without being in the weeds. **No generation stories — Reviewer role only (see US-19).** |
| **P4** | **DTC Brand Manager** *(Phase 2 ICP — not built for in V1)* | In-house growth marketer at a DTC brand | Runs paid social in-house, $10K–$100K/mo spend. Wears multiple hats. Needs speed and brand consistency without a creative team. Served by Growth tier after agency ICP is established. |

---

## Epic Structure

```
EPIC 1 — Onboarding & Activation
EPIC 2 — Qvora Brief (Strategy & Intelligence Layer)
EPIC 3 — Qvora Studio (Video Generation & Production Layer)
EPIC 4 — Export & Structured Testing
EPIC 5 — Brand Kit & Multi-Brand Management
EPIC 6 — Team & Collaboration
EPIC 7 — Qvora Signal (Performance Learning Layer — V2)
EPIC 8 — Platform & Administration
```

---

## EPIC 1 — Onboarding & Activation

> **Goal:** Get any agency user to their first complete ad set in under 15 minutes. Create a real artifact in the first 60 seconds — not a tutorial.

---

### US-01 — Signup and Role Selection

**As a** new user signing up for Qvora,
**I want to** select my role and use case during signup,
**so that** the product configures itself for my workflow from day one.

**Acceptance Criteria:**
- [ ] Signup form collects: name, work email, company name, role (Media Buyer / Creative Director / Brand Manager / Other), monthly ad spend range
- [ ] Role selection routes user to a role-appropriate empty state (not a generic dashboard)
- [ ] Signup completes in ≤ 3 fields before first value moment — remaining profile details collected progressively
- [ ] Email verification required before first generation
- [ ] Google OAuth supported as alternative to email/password

**Research signal:** Role-based onboarding paths lift 7-day retention by 35% (UserPilot, 2026). Arcads explicitly lacks an onboarding walkthrough — gap to exploit.

---

### US-02 — First Brand Setup

**As a** newly signed-up agency user,
**I want to** set up my first brand (or client brand) immediately after signup,
**so that** every ad I generate is on-brand from the first run.

**Acceptance Criteria:**
- [ ] Brand setup wizard is the first screen after email verification — not a dashboard
- [ ] Minimum viable brand kit: brand name, primary color (hex), logo upload (PNG/SVG)
- [ ] Optional extended inputs: secondary color, font, intro/outro bumper, tone of voice notes
- [ ] Brand kit auto-applied to all generated assets for that brand
- [ ] User can skip extended inputs and return later (progressive disclosure)
- [ ] Setup wizard completable in < 2 minutes

---

### US-03 — First Brief Generation (Activation Moment)

**As a** new user completing brand setup,
**I want to** paste a product URL and immediately see a generated creative brief,
**so that** I experience Qvora's core value within my first session.

**Acceptance Criteria:**
- [ ] URL input field is the first CTA after brand setup — no additional navigation required
- [ ] System generates a brief within 15 seconds of URL submission
- [ ] Brief preview shows: 3 creative angles, 3 hook variants per angle, recommended format, rationale summary
- [ ] User can approve brief and proceed to video generation in one click
- [ ] User can edit any angle or hook inline before proceeding
- [ ] Empty state copy: *"Paste your product URL. Get your brief in 15 seconds."*

**Research signal:** Fastest activation = user creates real artifact within 60 seconds. Brief preview IS the artifact — it must feel intelligent, not generic.

---

### US-04 — Onboarding Checklist

**As a** new user in my first week,
**I want to** see a lightweight progress checklist of setup milestones,
**so that** I know what's next without reading documentation.

**Acceptance Criteria:**
- [ ] Checklist appears in sidebar or dashboard header for first 7 days only
- [ ] Milestones: Brand created → First brief generated → First video exported → Ad account connected (Growth+)
- [ ] Each milestone dismisses from checklist when completed
- [ ] Checklist disappears after all milestones complete or after day 7, whichever comes first
- [ ] Progress bar shows % complete — increases completion rate by 20–30% (SaaSUI, 2026)

---

### US-04b — Trial State & Conversion Prompt

**As a** user on a 7-day free trial,
**I want to** see a clear countdown of my trial status and a smooth path to upgrade,
**so that** I know when my trial ends and can continue without losing my work.

**Acceptance Criteria:**
- [ ] Trial badge shown in top navigation throughout the 7-day period: *"Trial — X days left"*
- [ ] Day 5 warning: in-app banner — *"2 days left on your trial. Keep your ads — upgrade before they're locked."*
- [ ] Day 7 final warning: in-app modal on login — *"Your trial ends today. Upgrade to keep access to your briefs and assets."* with upgrade CTA
- [ ] Day 8 (trial expired): account enters locked state — generation blocked; upgrade screen shown on login
- [ ] Locked state message: *"Your trial has ended. Upgrade to Starter ($99/mo) to keep generating. Your 3 ad sets are saved and ready."*
- [ ] All briefs, assets, and brand kit data retained for 30 days post-trial-expiry before deletion
- [ ] If user upgrades during trial: trial badge replaced with plan name; no interruption to generation
- [ ] Conversion email sequence: Day 3 (value recap), Day 6 (urgency), Day 8 (last chance + asset retention warning)
- [ ] Trial-to-paid conversion tracked as a funnel metric per cohort

---

## EPIC 2 — Qvora Brief (Strategy & Intelligence Layer)

> **Goal:** Generate a creative brief that a senior performance marketer would approve — not just bullets from a product page.

---

### US-05 — URL-Based Brief Generation

**As a** media buyer or creative director,
**I want to** paste a product URL and receive a structured creative brief with angles, hooks, and format recommendations,
**so that** I can skip the 3-day briefing process and go straight to production.

**Acceptance Criteria:**
- [ ] Supports URLs: Shopify PDP, WooCommerce PDP, custom landing pages, App Store listings, Google Play listings
- [ ] Extraction pipeline pulls: product name, category, key features, pricing, proof points, CTA, visual assets
- [ ] Brief output structure:
  - Product Summary (2–3 sentences)
  - 3–5 Creative Angles (each with: angle name, rationale, target emotion, funnel stage)
  - 3 Hook Variants per Angle (each tagged: hook type, opening line, estimated CTR positioning)
  - Recommended Formats (ranked by platform fit: UGC / Product Demo / Text-Motion)
  - Suggested Platform Distribution (Meta Feed / Story / TikTok / YouTube Shorts)
- [ ] Extraction latency: < 10 seconds
- [ ] Brief generation latency: < 15 seconds (post-extraction)
- [ ] Handles JS-heavy SPAs and Shopify storefronts

**Edge cases:**
- Paywalled or login-required pages → prompt user to paste brief text manually
- 404 or dead URL → clear error with manual brief fallback
- Non-English product pages → extract in source language, generate brief in user's UI language

---

### US-06 — Manual Brief Input

**As a** creative director whose client hasn't launched a public product page yet,
**I want to** paste a manual product brief or description instead of a URL,
**so that** I can use Qvora for pre-launch campaigns and internal briefs.

**Acceptance Criteria:**
- [ ] Manual input mode available as alternative to URL input on brief creation screen
- [ ] Free-text input field: accepts up to 2,000 characters
- [ ] System generates same brief structure as URL-based flow
- [ ] Optional structured fields: product name, category, target audience, key differentiator, CTA — to improve output quality
- [ ] Brief quality indicator shown (confidence score or completeness signal based on input richness)

---

### US-07 — Brief Editing and Human Override

**As a** creative director reviewing an AI-generated brief,
**I want to** edit any angle, hook, or recommendation inline before generating videos,
**so that** I maintain creative control and brand voice while using AI as a starting point.

**Acceptance Criteria:**
- [ ] Every field in the brief is editable inline (click to edit)
- [ ] User can add a new angle (max 7 total)
- [ ] User can delete an angle or hook variant
- [ ] User can reorder angles via drag-and-drop
- [ ] "Regenerate this section" button available per angle and per hook variant
- [ ] All edits are saved automatically (no explicit save button)
- [ ] Brief version history: last 3 versions accessible via "History" link
- [ ] Edited brief marked as "Human-reviewed" in asset metadata (used downstream by Signal engine)

---

### US-08 — Brief Templates by Industry

**As a** media buyer managing clients across multiple verticals,
**I want to** select an industry template (e.g., DTC Fashion, Mobile App, SaaS, Health & Wellness) to bias the brief generation,
**so that** the angles and hooks are relevant to my client's category without manual editing.

**Acceptance Criteria:**
- [ ] Template selector available before URL submission (optional, default = auto-detect)
- [ ] Launch templates: DTC Fashion, DTC Beauty, Mobile App (iOS/Android), SaaS/Software, Health & Wellness, Food & Beverage, Financial Services
- [ ] Templates bias: angle archetypes, hook types, tone of voice, and format recommendations
- [ ] "Auto-detect" mode uses extracted product data to infer the closest template
- [ ] User can override auto-detected template at any time

---

### US-09 — Brief Library

**As a** creative director running multiple campaigns per month,
**I want to** save, search, and reuse past briefs,
**so that** I don't regenerate the same brief for seasonal or recurring campaigns.

**Acceptance Criteria:**
- [ ] All generated briefs saved to Brief Library automatically
- [ ] Library searchable by: brand, date, product name, angle type, status (draft / approved / in production)
- [ ] Brief can be duplicated and modified for a new campaign
- [ ] Brief can be archived (hidden from default view but not deleted)
- [ ] Growth+ tier: brief can be shared with team members

---

## EPIC 3 — Qvora Studio (Video Generation & Production Layer)

> **Goal:** Go from approved brief to a complete, export-ready ad set in one session. Quality bar: a media buyer should want to launch it, not just preview it.

---

### US-10 — Video Generation from Brief

**As a** media buyer with an approved brief,
**I want to** generate a full set of video ads — one per angle — in a single click,
**so that** I have a complete test set ready without individually managing each creative.

**Acceptance Criteria:**
- [ ] "Generate All" button on approved brief creates one video per angle (3–5 videos)
- [ ] Each video maps to its angle: uses the angle's hook variant, tone, emotion tag, and format recommendation
- [ ] Format options: UGC-style (AI avatar), Product Demo (product + B-roll + voiceover), Text-Motion (copy-driven)
- [ ] Generation latency: 60–180 seconds per video at standard quality (1080p)
- [ ] Progress indicator shown per video during generation queue
- [ ] User notified (in-app + email) when full set is ready
- [ ] All generated assets tagged with metadata: `angle_type`, `hook_type`, `format`, `emotion`, `platform`, `brief_id`, `variant_index`

**Research signal:** Creatify generates 5–10 variations per URL but without strategy layer. Qvora's per-angle mapping is the key UX differentiator.

---

### US-11 — Format Selection

**As a** media buyer preparing ads for multiple placements,
**I want to** select which formats and aspect ratios to generate for each angle,
**so that** I get platform-ready assets without manual resizing.

**Acceptance Criteria:**
- [ ] Format options per video:
  - UGC-style (AI avatar, talking-head)
  - Product Demo (product shots + voiceover + B-roll)
  - Text-Motion (animated copy + product image + music)
- [ ] Aspect ratio options: 9:16 (TikTok/Stories), 1:1 (Feed), 16:9 (YouTube)
- [ ] Platform compliance auto-checked: safe zones, caption placement, duration limits (15s / 30s / 60s)
- [ ] Default selection based on brief's platform recommendation
- [ ] Multi-format generation: user can select multiple formats per angle — each generates as separate asset

---

### US-12 — Avatar and Voice Selection

**As a** creative director producing UGC-style ads,
**I want to** select an avatar and voice that match my client's target audience,
**so that** the ad feels authentic and relatable, not generic.

**Acceptance Criteria:**
- [ ] Avatar library: minimum 100 avatars at launch (varied: age 20–60, diverse ethnicity, casual/professional styles)
- [ ] Filterable by: gender, age range, ethnicity, style (casual / professional / active)
- [ ] Voice library: minimum 20 voices (varied: gender, age, accent, energy level — calm / upbeat / authoritative)
- [ ] Preview: 5-second avatar preview clip and voice sample before selection
- [ ] Avatar + voice combination saved per brand kit (default for future generations)
- [ ] Custom avatar upload: V2 roadmap (not V1)

---

### US-13 — Video Preview and Editing

**As a** creative director reviewing generated videos,
**I want to** preview each video in-platform and make light edits without re-generating,
**so that** I can approve or adjust before export without a full regeneration cycle.

**Acceptance Criteria:**
- [ ] Inline video player in asset library — no external download needed to preview
- [ ] Light edits available without regeneration:
  - Swap hook text (opening line only)
  - Swap avatar or voice
  - Toggle caption on/off
  - Adjust CTA text overlay
  - Trim video duration (clip start/end)
- [ ] "Regenerate" button available for full re-generation with modified parameters
- [ ] Edit history tracked per asset (last 3 versions)
- [ ] Side-by-side comparison: view two variants of the same angle together

---

### US-14 — Batch Variant Generation

**As a** media buyer running creative testing at scale,
**I want to** generate multiple hook variants for a single angle in one action,
**so that** I can test 5–10 hooks per angle without manually triggering each generation.

**Acceptance Criteria:**
- [ ] "Generate Variants" option per angle: specify number of variants (1–10)
- [ ] Each variant uses a different hook from the brief's hook library for that angle
- [ ] Variants auto-named with `variant_index` (v1, v2, ... v10)
- [ ] Batch generation queued and processed asynchronously
- [ ] User notified when batch complete
- [ ] Agency tier: unlimited variants; Starter tier: max 3 variants per angle

**Research signal:** Meta campaigns now require 1,000+ creative assets (Marketing Brew, 2026). Batch generation is table stakes for agencies.

---

## EPIC 3b — AI Video Generation Modes

> **Goal:** Beyond the core UGC/Demo/Text-Motion formats, give creative directors and media buyers direct access to the three fundamental AI generation primitives — so they can build any ad format, not just the pre-built templates.

> **Story numbering note:** These stories are numbered US-14b, US-14c, US-14d to reflect their position within Epic 3. The Story Summary table at the end of this document uses the corrected numbering.

---

### US-14b — Text → Video Generation

**As a** creative director,
**I want to** type a text prompt (or use a script from my brief) and generate a cinematic video clip,
**so that** I can produce product B-roll, brand videos, and lifestyle footage without a camera or editor.

**Acceptance Criteria:**
- [ ] Text input field: free-form prompt (max 500 chars) OR auto-populated from brief angle script
- [ ] Model selector: Auto (recommended) / Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2
- [ ] Auto mode selects model based on prompt intent (cinematic → Veo; long duration → Kling; camera control → Runway)
- [ ] Duration options: 5s, 10s, 15s, 30s
- [ ] Style modifiers: Cinematic / UGC / Product / Abstract / Lifestyle
- [ ] Runway camera control: motion type selector (static / pan / zoom / dolly) when Runway is selected
- [ ] Veo 3.1 native audio: ambient sound + SFX included when Veo is selected
- [ ] Generation latency: 30–120s (shown with model-specific estimate)
- [ ] Output tagged: `generation_mode: text_to_video`, `model_used`, `prompt_text`
- [ ] One free regeneration per prompt if user is unsatisfied; second regeneration deducts ad credit

**Research signal:** FAL.AI unified API provides access to Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2 at $0.05–$0.40/sec. Kling 3.0 generates clips up to 2 minutes — no other model at this price point (Pinggy, 2026).

---

### US-14c — Image → Video Generation

**As a** media buyer with a product hero shot,
**I want to** upload a static product image and turn it into a short animated video clip,
**so that** I can create scroll-stopping motion ads from existing brand assets without a video shoot.

**Acceptance Criteria:**
- [ ] Image upload: PNG, JPG, WebP; min 512×512px; max 10MB
- [ ] Optional motion prompt: text describing desired motion (e.g., "gentle zoom in", "product rotating slowly")
- [ ] Motion prompt templates provided: Subtle Focus / Product Rotation / Lifestyle Pan / Liquid Texture / Hero Entrance
- [ ] Model options: Kling 3.0 (default) / Runway Gen-4 / Stable Video Diffusion
- [ ] Duration: 3s, 5s, 8s, 10s
- [ ] Motion intensity: Subtle / Medium / Dynamic (slider)
- [ ] Auto-source from brief: product images extracted from product page URL available directly (no re-upload)
- [ ] Multi-image (up to 3): generate sequential animation across multiple product images (P1)
- [ ] Generation latency: 20–60s per clip
- [ ] Output tagged: `generation_mode: image_to_video`, `source_image_url`, `motion_prompt`

**Research signal:** Kling 3.0 is the leader in image-to-video quality at 1080p/30fps — upload image + text prompt → video. WaveSpeedAI provides unified API access (Atlas Cloud, 2026).

---

### US-14d — Voice → Video Generation

**As a** creative director who wants an ad that sounds like a real person,
**I want to** upload a voice recording (or use a cloned brand voice) and sync it to an avatar,
**so that** I produce a lip-synced video ad that feels human and authentic — not generic AI.

**Acceptance Criteria:**
- [ ] Voice input Mode A — Upload: MP3, WAV, M4A; max 5 minutes
- [ ] Voice input Mode B — ElevenLabs generated: select from voice library (20+ voices) or brand kit default voice
- [ ] Voice input Mode C — Voice clone (Growth+): upload 30s+ audio sample → clone voice stored per brand
- [ ] Voice clone consent: checkbox declaration required ("I confirm I have rights to this voice") — cannot proceed without
- [ ] Avatar selection: same 100+ avatar library; filter by gender, age, ethnicity, style
- [ ] Lip-sync engine: HeyGen Avatar v3 (frame-accurate, micro-expressions, natural head movement)
- [ ] Language support: lip-sync works for all 175+ HeyGen-supported languages
- [ ] Duration: matches audio length; max 60s (Starter/Growth), max 5 min (Agency)
- [ ] Background options: solid color / blurred lifestyle / product-matched (from brand kit)
- [ ] Captions auto-generated from audio transcription; styled to brand kit
- [ ] Generation latency: 60–180s for 30s lip-synced video
- [ ] Output tagged: `generation_mode: voice_to_video`, `voice_source`, `avatar_id`, `language`

**Research signal:** HeyGen Avatar v3 ranked #1 for photorealism in 2026 — full-body motion capture, timing-aware gestures, accurate lip-sync across 175+ languages. ElevenLabs provides the voice layer. Arcads uses this pattern but without voice cloning or brand voice persistence (Marketer Milk, 2026).

---

## EPIC 4 — Export & Structured Testing

> **Goal:** Every export is test-ready. Named correctly, tagged correctly, structured for the media buyer to upload and interpret results.

---

### US-15 — Structured Export

**As a** media buyer preparing to upload ads to Meta or TikTok,
**I want to** export my full ad set as a structured ZIP with standardized file names,
**so that** I can upload to my ad account and track performance by creative variable without manual sorting.

**Acceptance Criteria:**
- [ ] Export formats: MP4 (H.264), MOV; resolution per platform
- [ ] Naming convention enforced: `{brand_slug}_{angle_type}_{format}_{hook_variant}_{YYYYMMDD}_v{n}.mp4`
  - Example: `acme_conversion_ugc_desire_20260501_v2.mp4`
- [ ] ZIP download contains: all selected videos + one `manifest.csv` with columns:
  - `filename`, `brand`, `angle_type`, `hook_type`, `format`, `emotion`, `platform`, `brief_id`, `variant_index`
- [ ] User can select subset of assets for export (not forced to export all)
- [ ] Individual asset download also available (right-click → download)
- [ ] Export delivery: CDN-accelerated, download starts < 30 seconds after trigger

---

### US-16 — Test Set Builder

**As a** media buyer designing a creative test,
**I want to** organize my generated ads into a structured A/B or multivariate test set before exporting,
**so that** the test is set up for clean statistical interpretation from the start.

**Acceptance Criteria:**
- [ ] Test Set Builder UI: drag-and-drop assets into a test grid
- [ ] Test dimensions: angle (rows) × format (columns) or angle × hook variant
- [ ] System validates test set structure: flags imbalances (e.g., 3 variants for angle A but 1 for angle B)
- [ ] Test set named and saved to project library
- [ ] Export test set as structured ZIP with manifest (same naming convention as US-15)
- [ ] Growth+ tier feature

---

## EPIC 5 — Brand Kit & Multi-Brand Management

> **Goal:** Agency users manage 5–20 client brands from one workspace. Switching between clients should be frictionless.

---

### US-17 — Multi-Brand Workspace

**As an** agency account manager,
**I want to** manage multiple client brand kits from a single Qvora workspace,
**so that** I don't need separate accounts for each client.

**Acceptance Criteria:**
- [ ] Workspace supports multiple brand kits: Starter = 1, Growth = 3, Agency = unlimited
- [ ] Brand switcher in top navigation: click → select brand → all context switches to that brand
- [ ] Each brand has isolated: briefs, assets, brand kit, avatar/voice defaults, performance data (V2)
- [ ] Brand kit fields: name, logo, primary/secondary hex colors, fonts, intro/outro bumper, tone notes
- [ ] Brand kit editable at any time; changes apply to future generations (not retroactive)
- [ ] Brand archived (not deleted) when client offboards — data retained for 90 days

---

### US-18 — Brand Kit Templates

**As a** media buyer onboarding a new client quickly,
**I want to** apply a brand kit template as a starting point,
**so that** I can get to first generation without fully specifying every brand element.

**Acceptance Criteria:**
- [ ] Templates available: Minimal (name + color only), Full (all fields), Agency Default (inherits workspace defaults)
- [ ] Template applies pre-filled defaults that user can override
- [ ] "Complete later" flag on incomplete brand kit fields — does not block generation, but shows reminder

---

## EPIC 6 — Team & Collaboration

> **Goal:** Agencies work in teams. The right people see the right things. No single-player bottlenecks.

---

### US-19 — Team Seats and Roles

**As an** agency owner setting up Qvora for my team,
**I want to** invite team members with defined roles,
**so that** media buyers, creative directors, and account managers each have appropriate access.

**Acceptance Criteria:**
- [ ] Roles: Admin (full access), Creator (generate + export, no billing/settings), Reviewer (view + comment, no generation)
- [ ] Invite by email; invite link expires in 48 hours
- [ ] Agency tier: unlimited seats; Growth tier: up to 5 seats; Starter: 1 seat
- [ ] Admin can remove or change role of any team member
- [ ] Activity log: shows who generated what and when (Admin view)

---

### US-20 — Brief Sharing and Approval

**As a** creative director,
**I want to** share a brief with a client or internal stakeholder for approval before generating videos,
**so that** I don't waste generation credits on a brief the client hasn't signed off on.

**Acceptance Criteria:**
- [ ] "Share brief" generates a read-only link (no Qvora account required to view)
- [ ] Shareable brief shows: angles, hooks, rationale, format recommendations — no internal metadata
- [ ] Link expiry: 7 days (configurable by Admin)
- [ ] Reviewer can leave comments on the shared brief link
- [ ] Comments appear in-product for the brief owner
- [ ] Brief status updated to "Approved" when creator marks it approved (manual action, not auto)

---

## EPIC 7 — Qvora Signal (Performance Learning Layer — V2)

> **Goal:** Close the loop between what Qvora generates and what actually performs. Make every future brief smarter than the last.

---

### US-21 — Ad Account Connection

**As a** media buyer on Growth+ plan,
**I want to** connect my Meta and TikTok ad accounts to Qvora,
**so that** Qvora can pull performance data on the ads I've uploaded.

**Acceptance Criteria:**
- [ ] OAuth 2.0 connection for: Meta Marketing API, TikTok Marketing API
- [ ] Scoped permissions: read-only on performance metrics; no budget write access without explicit consent
- [ ] Connection wizard: < 3 steps from "Connect Account" to first data sync
- [ ] Multiple ad accounts per brand supported (e.g., one Meta account per client brand)
- [ ] Connection health status shown in Settings (connected / disconnected / token expired)
- [ ] Token auto-refresh handled silently; user notified only on failure

---

### US-22 — Performance Dashboard

**As a** media buyer or creative director,
**I want to** see performance metrics for my Qvora-generated ads broken down by creative variable,
**so that** I can identify which angles, hooks, and formats are driving results.

**Acceptance Criteria:**
- [ ] Metrics shown: CTR, CPA, ROAS, video hold rate (25% / 50% / 75% / 100%), completion rate, spend
- [ ] Breakdown dimensions: angle type, hook type, format, emotion, platform
- [ ] Date range selector: last 7 / 30 / 90 days or custom
- [ ] Default view: angle performance table sorted by CTR descending
- [ ] Chart view: bar chart comparing formats; line chart for trend over time
- [ ] Minimum data threshold: performance breakdown shown only after 1,000 impressions per variant
- [ ] Data attributed only to assets generated by Qvora and tagged with `brief_id`

---

### US-23 — Creative Fatigue Detection

**As a** media buyer running long-running campaigns,
**I want to** receive an alert when a Qvora-generated ad is showing signs of creative fatigue,
**so that** I can refresh the creative before CTR declines significantly.

**Acceptance Criteria:**
- [ ] Fatigue signal: CTR drop > 30% from 7-day peak, sustained over 3 consecutive days
- [ ] Alert: in-app notification + email summary (weekly digest or immediate, user-configured)
- [ ] Alert message includes: asset name, fatigue metric, suggested action (refresh hook / swap avatar / new angle)
- [ ] One-click "Refresh this ad" from alert → opens Studio with same angle pre-loaded, new hook variant suggested
- [ ] `fatigue_detected_at` timestamp stored on asset record

---

### US-24 — Next-Gen Brief Recommendations

**As a** creative director preparing a new campaign for a returning client,
**I want to** see Qvora's data-backed recommendations for which angles and hooks to prioritize,
**so that** my next brief starts from learned performance data, not a blank slate.

**Acceptance Criteria:**
- [ ] Recommendations shown on "New Brief" screen for brands with ≥ 30 days of Signal data
- [ ] Recommendation format: "Based on your last 3 campaigns, [angle type] outperformed on CTR by [X%]. Suggested starting angle: [name]."
- [ ] Confidence score shown per recommendation (based on impression volume)
- [ ] User can accept recommendation (pre-populates brief) or ignore (starts fresh)
- [ ] Recommendations are per-organization — no cross-account data sharing
- [ ] Recommendations generated weekly, cached; refresh on new data sync

---

## EPIC 8 — Platform & Administration

---

### US-25 — Usage Dashboard

**As an** agency admin,
**I want to** see my team's generation usage against plan limits,
**so that** I can manage credits, anticipate overages, and right-size my plan.

**Acceptance Criteria:**
- [ ] Dashboard shows: ads generated this month, ads remaining, resets on billing date
- [ ] Usage broken down by: brand, team member, format type
- [ ] Alert at 80% usage: in-app banner + email notification
- [ ] Overage handling: generation blocked at limit (not auto-upgraded) with upgrade CTA
- [ ] Usage history: last 3 months visible

---

### US-26 — Billing and Plan Management

**As an** agency admin,
**I want to** upgrade, downgrade, or cancel my plan from within the product,
**so that** I don't need to contact support for routine plan changes.

**Acceptance Criteria:**
- [ ] Plan comparison table accessible from Settings > Billing
- [ ] Upgrade: immediate access to new tier features; prorated billing
- [ ] Downgrade: takes effect at next billing cycle; data retained
- [ ] Cancellation: takes effect at end of billing period; data retained for 90 days post-cancellation
- [ ] Invoice history: last 12 invoices downloadable as PDF
- [ ] Payment: Stripe-hosted; accepts major cards + SEPA (Enterprise)

---

### US-27 — API Access (Enterprise)

**As a** technical lead at a large agency,
**I want to** access Qvora's generation pipeline via API,
**so that** I can integrate creative generation into our internal tooling and programmatic campaign workflows.

**Acceptance Criteria:**
- [ ] REST API with API key authentication (generated in Settings > Developer)
- [ ] Endpoints: `POST /briefs` (create brief from URL or text), `POST /generations` (trigger video generation), `GET /assets` (list + filter), `GET /assets/{id}/download` (signed URL)
- [ ] Rate limits: Enterprise tier = 100 req/min; documented in API reference
- [ ] Webhook support: `generation.completed`, `generation.failed`, `fatigue.detected` events
- [ ] OpenAPI 3.0 spec published; SDKs: Python, Node.js (V2)
- [ ] API access: Enterprise tier only

---

## Story Summary

| Epic | Stories | V1 | V2 |
|---|---|---|---|
| Onboarding & Activation | US-01 to US-04, US-04b | ✅ All | — |
| Qvora Brief | US-05 to US-09 | ✅ All | — |
| Qvora Studio (Core) | US-10 to US-14 | ✅ All | — |
| AI Video Generation Modes | US-14b to US-14d | ✅ All | US-14d multi-language translation |
| Export & Testing | US-15 to US-16 | ✅ All | — |
| Brand Kit & Multi-Brand | US-17 to US-18 | ✅ All | — |
| Team & Collaboration | US-19 to US-20 | ✅ All | — |
| Qvora Signal | US-21 to US-24 | ❌ Schema only | ✅ Full |
| Platform & Admin | US-25 to US-27 | ✅ US-25, US-26 | ✅ US-27 (Enterprise) |

---

*User Stories v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources used in research:**
- [Creatify Review 2026 — Superscale AI](https://superscale.ai/alternatives/creatify/review)
- [Arcads AI Review 2026 — EzUGC](https://www.ezugc.ai/blog/arcads-ai)
- [How Meta's AI push is changing ad creation — Marketing Brew](https://www.marketingbrew.com/stories/2026/04/07/meta-ai-ad-creation)
- [SaaS Onboarding Flows That Actually Convert in 2026 — SaaSUI](https://www.saasui.design/blog/saas-onboarding-flows-that-actually-convert-2026)
- [AI User Onboarding: 8 Real Ways to Optimize — Userpilot](https://userpilot.com/blog/ai-user-onboarding/)
- [Performance Marketing Agency Workflow — Smartsheet](https://www.smartsheet.com/content/creative-agency-process-workflows)
