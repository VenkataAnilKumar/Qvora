# QVORA
## Feature Specification
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Draft

---

## Overview

This document specifies functional requirements, acceptance criteria, edge cases, and V1/V2 scope boundaries for each Qvora feature module. It is the engineering-handoff companion to the User Stories document.

---

## Module Index

| Module | Feature ID Prefix | V1 Scope |
|---|---|---|
| [FS-1] URL Ingestion & Extraction | EXT | ✅ |
| [FS-2] Creative Strategy Engine (Qvora Brief) | BRIEF | ✅ |
| [FS-3] Video Generation Engine (Qvora Studio) | GEN | ✅ |
| [FS-3a] Text → Video Generation | T2V | ✅ |
| [FS-3b] Image → Video Generation | I2V | ✅ |
| [FS-3c] Voice → Video Generation | V2V | ✅ |
| [FS-4] Voiceover & Caption Engine | VOICE | ✅ |
| [FS-5] Brand Kit System | BRAND | ✅ |
| [FS-6] Export & Naming Engine | EXPORT | ✅ |
| [FS-7] Asset Library | LIB | ✅ |
| [FS-8] Team & Collaboration | TEAM | ✅ |
| [FS-9] Ad Account Connector | CONN | V2 |
| [FS-10] Performance Learning Engine (Qvora Signal) | SIGNAL | V2 |
| [FS-11] Platform & Billing | PLAT | ✅ |

---

## [FS-1] URL Ingestion & Extraction

**Purpose:** Convert a product URL into structured product data that feeds the creative strategy engine.

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| EXT-01 | Accept HTTP/HTTPS URLs as primary input | P0 |
| EXT-02 | Support Shopify, WooCommerce, custom landing pages, App Store (iOS), Google Play | P0 |
| EXT-03 | Render JavaScript before extraction (headless browser — Playwright or Puppeteer) | P0 |
| EXT-04 | Extract fields: product name, category, price, key features (list), proof points, primary CTA, image URLs | P0 |
| EXT-05 | Extraction latency ≤ 10 seconds for standard pages | P0 |
| EXT-06 | Fallback to manual text input when URL fails | P0 |
| EXT-07 | Detect page language; extract in source language; translate output if UI language differs | P1 |
| EXT-08 | Handle PDF product sheets via file upload fallback | P2 |
| EXT-09 | Cache extraction results for 24 hours per URL (avoid repeat scraping) | P1 |
| EXT-10 | Log extraction confidence score (0–100) for each field; surface to user if < 60 | P1 |

### Edge Cases

| Scenario | Handling |
|---|---|
| 404 / dead URL | Show error: *"We couldn't reach that page. Try pasting your product description instead."* |
| Paywall / login required | Detect 401/403 redirect; show manual input fallback immediately |
| JavaScript SPA (no SSR) | Headless browser renders page; fallback to partial extraction if render > 15s |
| Shopify product variants | Extract primary variant by default; show variant selector if multiple |
| Non-English page | Extract in source language; translate to user UI language via LLM step |
| Very long product page (> 50K tokens) | Chunk and summarize; prioritize above-fold content |

### Out of Scope (V1)
- PDF upload extraction
- Bulk URL batch import
- Video/YouTube URL extraction

---

## [FS-2] Creative Strategy Engine (Qvora Brief)

**Purpose:** Transform extracted product data into a structured creative brief with performance-marketer-grade strategic intelligence.

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| BRIEF-01 | Generate 3–5 creative angles per product | P0 |
| BRIEF-02 | Each angle: name, rationale, target emotion, funnel stage (awareness / consideration / conversion / retention) | P0 |
| BRIEF-03 | Generate 3 hook variants per angle | P0 |
| BRIEF-04 | Each hook: hook type (problem / desire / social-proof / shock / curiosity), opening line, estimated positioning | P0 |
| BRIEF-05 | Recommend 2–3 formats per brief (UGC / Product Demo / Text-Motion) with platform rationale | P0 |
| BRIEF-06 | Recommend platform distribution (Meta Feed / Story / TikTok / YouTube Shorts) | P0 |
| BRIEF-07 | Brief generation latency ≤ 15 seconds post-extraction | P0 |
| BRIEF-08 | All brief fields editable inline; changes auto-saved | P0 |
| BRIEF-09 | "Regenerate section" available per angle and per hook | P0 |
| BRIEF-10 | Industry template bias: 7 templates at launch (see US-08) | P1 |
| BRIEF-11 | Brief version history: retain last 3 versions per brief | P1 |
| BRIEF-12 | Brief confidence score: derived from extraction quality and template match | P1 |
| BRIEF-13 | Brief stored with full metadata: `brand_id`, `product_url`, `template_used`, `created_by`, `created_at`, `last_edited_at`, `human_reviewed` flag | P0 |
| BRIEF-14 | Manual brief input: free-text up to 2,000 characters | P0 |
| BRIEF-15 | Brief shareable via read-only link (no Qvora account required) | P1 |

### Data Model (V1 Schema — feeds V2 Signal)

```sql
briefs
  id              UUID PRIMARY KEY
  brand_id        UUID FK → brands
  product_url     TEXT
  product_summary TEXT
  template_used   VARCHAR(50)
  human_reviewed  BOOLEAN DEFAULT FALSE
  created_by      UUID FK → users
  created_at      TIMESTAMPTZ
  updated_at      TIMESTAMPTZ

brief_angles
  id              UUID PRIMARY KEY
  brief_id        UUID FK → briefs
  angle_name      VARCHAR(100)
  rationale       TEXT
  emotion         VARCHAR(50)   -- aspiration/urgency/fear/humor/trust
  funnel_stage    VARCHAR(30)   -- awareness/consideration/conversion/retention
  sort_order      INTEGER

brief_hooks
  id              UUID PRIMARY KEY
  angle_id        UUID FK → brief_angles
  hook_type       VARCHAR(50)   -- problem/desire/social-proof/shock/curiosity
  opening_line    TEXT
  variant_index   INTEGER
```

### LLM Prompt Architecture

```
System: You are a senior performance marketing creative strategist...
        [Role definition + industry template bias if selected]

User:   Product data: {extracted_product_json}
        Industry: {template_name}
        Target platform: {platform_preference}
        Generate a creative brief following this structure: {schema}

Output: Structured JSON matching brief schema
```

- Model: GPT-4o class or equivalent (Claude Opus-class for quality; flash/haiku-class for iteration)
- Temperature: 0.7 for initial generation; 0.9 for regeneration variants
- Output format: JSON strict mode / structured outputs

### Quality Gate (pre-delivery)
- Validate: all required fields populated, hook types are valid enum values, funnel stages are valid enum values
- Reject and retry (up to 3 attempts) if validation fails
- Surface error to user only after 3 failed attempts

### Out of Scope (V1)
- Multi-product brief (single product per brief only)
- Competitor-comparative brief generation
- Brief collaboration (real-time multi-user editing)

---

## [FS-3] Video Generation Engine (Qvora Studio)

**Purpose:** Produce platform-ready video ads from a brief's angles, hooks, and format recommendations.

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| GEN-01 | Support 3 output formats: UGC-style, Product Demo, Text-Motion | P0 |
| GEN-02 | Support 3 aspect ratios: 9:16, 1:1, 16:9 | P0 |
| GEN-03 | Support 3 durations: 15s, 30s, 60s | P0 |
| GEN-04 | Resolution: 1080p standard; 4K Enterprise tier | P0 / P2 |
| GEN-05 | Generation latency: 60–180s per video at 1080p standard quality | P0 |
| GEN-06 | "Generate All" from approved brief: creates one video per angle | P0 |
| GEN-07 | Batch variant generation: Starter = max 3 variants/angle; Growth = max 10; Scale = unlimited | P0 |
| GEN-08 | Brand kit auto-applied: logo watermark, color overlay, intro/outro bumper | P0 |
| GEN-09 | Platform compliance auto-check: safe zone margins, text size, duration limit per platform | P0 |
| GEN-10 | Generation queue: async processing with in-app + email notification on completion | P0 |
| GEN-11 | Asset metadata tagged at generation: `angle_type`, `hook_type`, `format`, `emotion`, `platform`, `brief_id`, `variant_index` | P0 |
| GEN-12 | Light edits without re-generation: swap hook text, swap avatar/voice, caption toggle, CTA text, trim duration | P1 |
| GEN-13 | Full regeneration with modified parameters | P0 |
| GEN-14 | Generation cost per video must support profitable delivery at $99/mo Starter tier | P0 |

### Format Specifications

#### UGC-Style
- AI avatar (talking-head): selected from avatar library
- Script: hook + body (from brief) + CTA
- Voiceover: avatar lip-synced to voice selection
- B-roll overlay: optional product imagery behind / beside avatar
- Captions: auto-generated, styled to brand kit

#### Product Demo
- Opening: product hero shot / animation (3–5 seconds)
- Middle: feature showcase with voiceover (pulled from brief angle)
- CTA: animated text overlay + product shot
- B-roll: sourced from product page imagery + stock footage (licensed)

#### Text-Motion
- Animated typography: hook text animates in on screen
- Product imagery: background or split-screen
- Voiceover: optional (default off for text-motion)
- Music: background track (licensed library)

### Generation Pipeline Architecture

```
[Brief JSON]
     ↓
[Script Generator]
  — Hook + Body + CTA from brief angle/hook
  — Duration-mapped script (150 WPM for 30s = ~75 words)
     ↓
[Video Compositor]
  — Format router: UGC / Demo / Text-Motion
  — Asset assembler: avatar / product imagery / B-roll / text layers
  — Brand kit injector
     ↓
[Voiceover Renderer]
  — TTS generation (ElevenLabs API or equivalent)
  — Lip-sync for UGC format
     ↓
[Caption Generator]
  — Auto-transcription + styling
     ↓
[Platform Formatter]
  — Aspect ratio crop/pad
  — Safe zone compliance check
  — Duration trim if needed
     ↓
[Output: MP4 H.264 / H.265]
  — Asset stored in object storage (S3-compatible)
  — Metadata written to DB
  — Asset Library updated
```

### Generation Cost Model (V1 Constraint)

| Format | Est. GPU Cost/Video | Margin Target at $99/mo (20 ads) |
|---|---|---|
| Text-Motion | $0.05–$0.10 | ✅ Profitable |
| Product Demo | $0.15–$0.25 | ✅ Profitable |
| UGC-style (avatar) | $0.30–$0.60 | ⚠️ Monitor — must stay < $0.50 |
| Text → Video (15s) | $0.08–$0.20 | ✅ Profitable (via FAL.AI at ~$0.05–$0.40/sec) |
| Image → Video (5–10s) | $0.05–$0.15 | ✅ Profitable |
| Voice → Video (30s, lip-sync) | $0.40–$0.80 | ⚠️ Monitor — HeyGen API credits model |

- Starter tier: 20 ads/mo = max $10 COGS at $0.50/video
- Growth tier: 100 ads/mo = max $50 COGS at $0.50/video
- **Constraint:** Average video generation COGS must stay ≤ $0.40 at standard quality for unit economics to work
- **API layer:** FAL.AI unified API for T2V/I2V (access to Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2 at $0.05–$0.40/sec); HeyGen Avatar IV API for V2V lip-sync

### Edge Cases

| Scenario | Handling |
|---|---|
| Generation timeout (> 300s) | Auto-retry once; if second failure, notify user and offer credit |
| Avatar/voice combination incompatible | Surface warning at selection; disable incompatible combinations |
| Product page images too low resolution | Use brand kit assets as fallback; warn user |
| Script exceeds duration (too many words) | Auto-truncate at CTA; show word count warning in brief |
| Platform compliance failure (safe zone) | Auto-recompose with adjusted layout; flag to user if unresolvable |

### Out of Scope (V1)
- Custom avatar upload (Enterprise roadmap)
- Real actor video (human UGC marketplace)
- Motion graphics / After Effects template rendering
- Audio music selection (auto-assigned from licensed library)

---

## [FS-3a] Text → Video Generation

**Purpose:** Generate video from a script or free-form text prompt — pure cinematic output without an avatar. Used for product B-roll, brand videos, abstract lifestyle shots.

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| T2V-01 | Accept text input: free-form prompt (max 500 chars) OR auto-generated from brief script | P0 |
| T2V-02 | Model selector: Auto / Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2 | P0 |
| T2V-03 | Auto mode: Qvora selects model based on prompt type (cinematic → Veo; long-form → Kling; control → Runway) | P1 |
| T2V-04 | Duration options: 5s, 10s, 15s, 30s (model-dependent ceiling) | P0 |
| T2V-05 | Resolution: 1080p standard; 4K for Veo 3.1 (Enterprise tier) | P0 |
| T2V-06 | Native audio generation: enabled for Veo 3.1 (ambient sound, SFX) | P1 |
| T2V-07 | Generation latency: 30–120s depending on model and duration | P0 |
| T2V-08 | Style modifiers: Cinematic / UGC / Product / Abstract / Lifestyle | P1 |
| T2V-09 | Camera controls (Runway): motion type (static / pan / zoom / dolly) | P1 |
| T2V-10 | All outputs tagged with `generation_mode: text_to_video`, `model_used`, `prompt_text` | P0 |
| T2V-11 | API layer: FAL.AI unified API for model routing; fallback model if primary unavailable | P0 |

### Model Routing Logic

```
User selects "Auto" →
  IF prompt contains product/object focus  → Kling 3.0
  IF prompt requires long clip (> 30s)     → Kling 3.0 (up to 2 min)
  IF prompt requires 4K / native audio     → Veo 3.1
  IF prompt requires camera control        → Runway Gen-4.5
  DEFAULT                                  → Kling 3.0 (best cost/quality ratio)
```

### Use Cases in Qvora Workflow
- **Brief → Script → T2V:** Brief generates a script; user sends script to T2V for a cinematic product ad (no avatar)
- **B-roll generator:** Creative director generates 5 × 5s clips from product descriptions to use as B-roll in composite ads
- **Brand video:** Full 30s brand story from a narrative prompt — no avatar, no voiceover overlay required

### Edge Cases

| Scenario | Handling |
|---|---|
| Model unavailable (FAL.AI outage) | Route to next preferred model; notify user of model used |
| Prompt violates content policy | Pre-screen via moderation layer before sending to model API; return error with guidance |
| Output quality below threshold | Allow one free regeneration per generation; deduct credit on second |
| Prompt too vague | Show prompt improvement hints (e.g., "Add product name, setting, and mood") |

---

## [FS-3b] Image → Video Generation

**Purpose:** Animate a static product image into a short video clip — subtle motion, zoom, parallax, or full scene animation. Converts product hero shots into scroll-stopping motion ads.

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| I2V-01 | Accept image input: PNG, JPG, WebP; minimum 512×512px; maximum 10MB | P0 |
| I2V-02 | Optional motion prompt: text describing desired motion (e.g., "gentle zoom in", "product rotating", "liquid splash") | P0 |
| I2V-03 | Model options: Kling 3.0 (default) / Runway Gen-4 / Stable Video Diffusion | P0 |
| I2V-04 | Duration: 3s, 5s, 8s, 10s | P0 |
| I2V-05 | Motion intensity slider: Subtle / Medium / Dynamic | P1 |
| I2V-06 | Output resolution: matches input resolution up to 1080p | P0 |
| I2V-07 | Generation latency: 20–60s per clip | P0 |
| I2V-08 | Multi-image input: upload up to 3 images for sequential animation (product showcase) | P1 |
| I2V-09 | Product page auto-source: use extracted product images from Qvora Brief directly (no re-upload) | P0 |
| I2V-10 | All outputs tagged with `generation_mode: image_to_video`, `source_image_url`, `motion_prompt` | P0 |

### Use Cases in Qvora Workflow
- **Product image → scroll-stopping ad:** Upload a product hero shot → animate with subtle zoom + product name text overlay → 5s TikTok/Reel opener
- **Static → dynamic:** Convert static product catalog images into motion for Stories and Reels without a full video shoot
- **B-roll from brand assets:** Agency uploads client brand assets → generates motion B-roll for composite UGC ads

### Motion Prompt Templates (V1)

```
Subtle Product Focus:  "Gentle zoom in on [product], soft bokeh background"
Product Rotation:      "[Product] slowly rotating on white surface, studio lighting"
Lifestyle Motion:      "Camera slowly pans across [setting], [product] in foreground"
Liquid / Texture:      "[Material] rippling, smooth motion, close-up"
Hero Entrance:         "[Product] slides in from left, comes to rest, subtle highlight"
```

### Edge Cases

| Scenario | Handling |
|---|---|
| Image resolution too low (< 512px) | Warn and upscale via AI upscaler before generation; flag to user |
| Image contains person's face | Apply face detection; prompt user to confirm consent before generation |
| Motion prompt conflicts with image content | Generate with best interpretation; show prompt + output side-by-side |
| Multi-image: images have inconsistent style | Warn about stylistic inconsistency; proceed on user confirmation |

---

## [FS-3c] Voice → Video Generation

**Purpose:** Sync a voice (uploaded audio or ElevenLabs-generated) to a chosen avatar for lip-synced video. Enables authentic-feeling UGC ads where the voice drives the video — not a generic AI voice paired to a generic avatar.

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| V2V-01 | Voice input mode A — Upload audio: accept MP3, WAV, M4A; max 5 minutes | P0 |
| V2V-02 | Voice input mode B — ElevenLabs generated: select from voice library or use brand kit voice | P0 |
| V2V-03 | Voice input mode C — Voice clone: upload 30s+ audio sample → clone voice for this brand | P1 (Growth+) |
| V2V-04 | Avatar selection: same library as UGC-style (100+ avatars at launch) | P0 |
| V2V-05 | Lip-sync engine: HeyGen Avatar IV API (most photorealistic; 175+ languages) | P0 |
| V2V-06 | Lip-sync accuracy: frame-accurate to uploaded audio; natural micro-expressions and head movement | P0 |
| V2V-07 | Language support: lip-sync works for all 175+ HeyGen-supported languages at launch | P0 |
| V2V-08 | Duration: matches audio length; max 60s at Starter/Growth; max 5 min at Scale/Enterprise | P0 |
| V2V-09 | Background options: solid color / blurred lifestyle / product-matched (from brand kit) | P1 |
| V2V-10 | Caption auto-generation: transcribed from uploaded audio, styled to brand kit | P0 |
| V2V-11 | Generation latency: 60–180s for 30s lip-synced video | P0 |
| V2V-12 | All outputs tagged with `generation_mode: voice_to_video`, `voice_source` (upload/elevenlabs/clone), `avatar_id`, `language` | P0 |

### Voice → Video Pipeline

```
[Voice Input]
  Upload audio (MP3/WAV)  OR  ElevenLabs TTS  OR  Voice Clone
         ↓
[Audio Processing]
  Normalize volume
  Trim silence at start/end
  Transcribe for captions (Whisper API or equivalent)
         ↓
[Avatar Lip-Sync — HeyGen Avatar IV API]
  Avatar selection
  Lip-sync rendering: audio → mouth + facial expressions + head motion
  Background composition
         ↓
[Post-Processing]
  Brand kit overlay (logo, color, captions)
  Platform format (9:16 / 1:1 / 16:9)
  Duration compliance check
         ↓
[Output: MP4 H.264]
```

### Voice Clone Spec (Growth+ tier)
- Upload requirement: 30 seconds clean audio (minimal background noise)
- Clone fidelity: voice timbre, pace, and accent preserved
- Clone stored per brand — reusable across all V2V generations for that brand
- Consent: user must check a consent declaration before cloning ("I confirm I have rights to this voice")
- GDPR: cloned voice data deleted on account cancellation; user can delete manually at any time

### Multi-Language Workflow (V2 expansion)
- Upload English audio → select target language → Qvora translates script + re-voices in target language + re-syncs lip-sync
- Powered by HeyGen video translation pipeline (175+ languages)
- V1 scope: English only with multi-language avatar delivery; full translation pipeline in V2

### Edge Cases

| Scenario | Handling |
|---|---|
| Audio quality too low (heavy noise) | Warn + apply noise reduction; proceed on user confirmation |
| Audio contains multiple speakers | Detect and warn: V2V works best with single speaker; user proceeds at own risk |
| Audio exceeds tier duration limit | Trim to tier maximum with warning; user can upgrade for longer |
| Voice clone quality insufficient (< 30s sample) | Reject with guidance: "Upload at least 30 seconds for best results" |
| Lip-sync desync on fast speech | Flag to user; offer "smooth pacing" re-render option |

---

## [FS-4] Voiceover & Caption Engine

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| VOICE-01 | TTS provider: ElevenLabs API or equivalent (multilingual V2) | P0 |
| VOICE-02 | Minimum 20 voices at launch: varied gender, age, accent, energy | P0 |
| VOICE-03 | Lip-sync for UGC avatar format | P0 |
| VOICE-04 | Voiceover latency: < 5 seconds per 30-second audio clip | P0 |
| VOICE-05 | Caption auto-generation: word-level timestamps from voiceover audio | P0 |
| VOICE-06 | Caption styling: font, color, and position pulled from brand kit | P0 |
| VOICE-07 | Caption toggle: on/off at preview and export stage | P0 |
| VOICE-08 | Multi-language voiceover: V2 (minimum 10 languages) | V2 |

---

## [FS-5] Brand Kit System

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| BRAND-01 | Brand kit fields: name, logo (PNG/SVG), primary hex, secondary hex, font (TTF/OTF), intro bumper (MP4 ≤ 5s), outro bumper (MP4 ≤ 5s), tone notes (free text) | P0 |
| BRAND-02 | Logo auto-positioned as watermark on all generated videos (bottom-right default; configurable) | P0 |
| BRAND-03 | Brand color applied to: caption background, CTA overlay, text-motion typography | P0 |
| BRAND-04 | Font applied to: caption text, CTA text overlay, title cards | P0 |
| BRAND-05 | Intro/outro bumper: pre-pended / appended to all generated videos for that brand | P0 |
| BRAND-06 | Brand kit versioned: changes create a new version; assets generated with prior version retain original branding | P1 |
| BRAND-07 | Multi-brand per workspace: Starter = 1, Growth = 3, Scale = 10, Enterprise = unlimited | P0 |
| BRAND-08 | Brand switcher: < 2 clicks to switch active brand context | P0 |

---

## [FS-6] Export & Naming Engine

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| EXPORT-01 | Export formats: MP4 H.264, MP4 H.265, MOV | P0 |
| EXPORT-02 | Resolution per platform: Meta Feed (1080×1080), Stories/TikTok (1080×1920), YouTube (1920×1080) | P0 |
| EXPORT-03 | Naming convention enforced: `{brand_slug}_{angle_type}_{format}_{hook_variant}_{YYYYMMDD}_v{n}.mp4` | P0 |
| EXPORT-04 | ZIP export: all selected assets + `manifest.csv` | P0 |
| EXPORT-05 | `manifest.csv` columns: `filename`, `brand`, `angle_type`, `hook_type`, `format`, `emotion`, `platform`, `brief_id`, `variant_index` | P0 |
| EXPORT-06 | CDN-accelerated delivery: download starts < 30 seconds after trigger | P0 |
| EXPORT-07 | Individual asset download via signed URL (48hr expiry) | P0 |
| EXPORT-08 | Export history: last 10 export bundles retained, re-downloadable for 30 days | P1 |

### Naming Convention Detail

```
{brand_slug}     — lowercase, hyphenated brand name (e.g., "acme-co")
{angle_type}     — enum: awareness / consideration / conversion / retention
{format}         — enum: ugc / demo / text-motion
{hook_variant}   — enum: problem / desire / social-proof / shock / curiosity
{YYYYMMDD}       — generation date
_v{n}            — variant index within same angle+format+hook combination

Example:
  acme-co_conversion_ugc_desire_20260501_v2.mp4
```

---

## [FS-7] Asset Library

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| LIB-01 | All generated assets stored and accessible in Asset Library | P0 |
| LIB-02 | Filter by: brand, brief, angle type, format, date, status (draft / exported / archived) | P0 |
| LIB-03 | Inline preview: video player within library (no download required to preview) | P0 |
| LIB-04 | Side-by-side comparison: select 2 assets → compare view | P1 |
| LIB-05 | Asset actions: preview, download, duplicate, archive, delete | P0 |
| LIB-06 | Bulk actions: select multiple → bulk export, bulk archive | P1 |
| LIB-07 | Storage limit: Starter = 50 assets retained; Growth = 500; Scale/Enterprise = unlimited | P0 |
| LIB-08 | Asset metadata visible on hover / detail panel: all tags, brief source, generation date | P0 |

---

## [FS-8] Team & Collaboration

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| TEAM-01 | Roles: Admin, Creator, Reviewer | P0 |
| TEAM-02 | Invite by email; invite link expires 48h | P0 |
| TEAM-03 | Seat limits per tier: Starter = 1, Growth = 5, Scale = unlimited | P0 |
| TEAM-04 | Admin: full access including billing, settings, team management | P0 |
| TEAM-05 | Creator: generate briefs and videos, export, manage own assets | P0 |
| TEAM-06 | Reviewer: view briefs and assets, comment only | P0 |
| TEAM-07 | Activity log: Admin view — who generated what, when | P1 |
| TEAM-08 | Brief sharing: read-only external link, no Qvora account required | P1 |
| TEAM-09 | Brief comments: Reviewer and Creator can comment on briefs | P1 |

---

## [FS-9] Ad Account Connector (V2)

> **V1 Requirement:** Data schema (FK columns on assets table) must be defined in V1 even though connector is not built until V2. See Product Definition §6 — V2 Learning Architecture V1 Design Constraint.

### V2 Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| CONN-01 | OAuth 2.0 connection: Meta Marketing API, TikTok Marketing API | V2-P0 |
| CONN-02 | Scoped permissions: read-only metrics; no budget write access without explicit consent | V2-P0 |
| CONN-03 | Multiple ad accounts per brand supported | V2-P0 |
| CONN-04 | Auto-match Qvora-exported assets to ad account creative IDs via filename or metadata | V2-P0 |
| CONN-05 | Sync frequency: every 6 hours for active campaigns; daily for inactive | V2-P1 |
| CONN-06 | Connection health monitoring: auto-refresh tokens; alert on failure | V2-P0 |

---

## [FS-10] Performance Learning Engine — Qvora Signal (V2)

### V2 Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| SIGNAL-01 | Ingest per-asset metrics: CTR, CPA, ROAS, hold rate (25/50/75/100%), completion rate, spend | V2-P0 |
| SIGNAL-02 | Correlate metrics to asset metadata tags: `angle_type`, `hook_type`, `format`, `emotion` | V2-P0 |
| SIGNAL-03 | Performance breakdown dashboard: by angle, hook type, format, platform | V2-P0 |
| SIGNAL-04 | Minimum data threshold: 1,000 impressions per variant before surfacing breakdown | V2-P0 |
| SIGNAL-05 | Creative fatigue detection: CTR drop > 30% from 7-day peak, sustained 3 days | V2-P0 |
| SIGNAL-06 | Fatigue alerts: in-app + email; configurable frequency | V2-P0 |
| SIGNAL-07 | Next-gen recommendations: pre-populate new brief based on top-performing angle+hook combinations | V2-P1 |
| SIGNAL-08 | Per-organization learning only: zero cross-account data sharing | V2-P0 |
| SIGNAL-09 | GDPR-compliant data retention: performance data deleted on org cancellation after 90-day hold | V2-P0 |

---

## [FS-11] Platform & Billing

### Functional Requirements

| ID | Requirement | Priority |
|---|---|---|
| PLAT-01 | Authentication: email/password + Google OAuth | P0 |
| PLAT-02 | SSO via SAML 2.0: Enterprise tier | P2 |
| PLAT-03 | Billing: Stripe-hosted; monthly and annual plans | P0 |
| PLAT-04 | Plan tiers: Starter ($99), Growth ($149), Scale ($399), Enterprise (custom) | P0 |
| PLAT-05 | Usage metering: ads generated per month; resets on billing date | P0 |
| PLAT-06 | Usage alerts: 80% and 100% thresholds — in-app + email | P0 |
| PLAT-07 | Overage: generation blocked at limit (no auto-upgrade); upgrade CTA shown | P0 |
| PLAT-08 | Free trial: 7-day full-access trial on signup (no credit card). At trial end (Day 8): account enters locked state — generation blocked, all data retained. Upgrade CTA shown on every login. Briefs, assets, and brand kit retained for 30 days post-expiry before deletion. Conversion email sequence: Day 3 (value recap), Day 6 (urgency), Day 8 (last chance + retention warning). | P0 |
| PLAT-09 | Audit logs: full generation history per org; GDPR-compliant deletion on request | P0 |
| PLAT-10 | REST API + API key auth: Enterprise tier | P2 |
| PLAT-11 | Webhook events: `generation.completed`, `generation.failed`, `fatigue.detected` | P2 |

---

## Priority Key

| Priority | Meaning |
|---|---|
| **P0** | Launch blocker. Not shipped = product doesn't work. |
| **P1** | High value. Target for V1 but can cut if needed. |
| **P2** | Nice to have. Enterprise / post-launch. |
| **V2** | Explicitly post-V1 — but schema/architecture must be designed in V1. |

---

## V1 vs V2 Scope Summary

| Module | V1 Ships | V2 Adds |
|---|---|---|
| Extraction | Full | PDF upload, bulk URL |
| Brief Engine | Full | Competitor-comparative briefs, real-time co-editing |
| Studio | Full (3 formats, 3 ratios, 3 durations) | Custom avatars, real actor UGC |
| Voiceover | EN only | 10+ languages |
| Brand Kit | Full | |
| Export | Full | Direct-to-ad-account push |
| Asset Library | Full | |
| Team | Full (roles, invites) | SSO |
| Ad Account Connector | Schema only | Full OAuth + sync |
| Qvora Signal | Schema + tags only | Full dashboard, fatigue alerts, recommendations |
| API | | Full REST API + webhooks |

---

*Feature Specification v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources used in research:**
- [Creatify AI Review 2026 — Segwise](https://segwise.ai/blog/best-ai-advertising-creative-tools-2026)
- [Arcads AI Review 2026 — Dupple](https://dupple.com/tools/arcads-ai)
- [FAQ on AI Creative Optimization — eMarketer](https://www.emarketer.com/content/faq-on-ai-creative-optimization--what-automate--what-keep-human--how-compete)
- [IAB AI-Powered Video Outcomes — March 2026](https://www.iab.com/guidelines/ai-powered-video-outcomes-march-2026/)
