---
title: Feature Modules
category: product
tags: [features, V1, V2, modules, acceptance-criteria]
sources: [Qvora_Feature-Spec, Qvora_User-Stories]
updated: 2026-04-15
---

# Feature Modules

## TL;DR
14 feature modules. 9 are V1. Connected Ad Accounts (CONN) and the Performance Learning Engine (SIGNAL) are V2. Image→Video (I2V) is also V2.

---

## Module Index

| Module | ID Prefix | V1 |
|---|---|---|
| URL Ingestion & Extraction | EXT | ✅ |
| Creative Strategy Engine (Qvora Brief) | BRIEF | ✅ |
| Video Generation Engine (Qvora Studio) | GEN | ✅ |
| Text → Video | T2V | ✅ |
| Image → Video | I2V | V2 |
| Voice → Video (avatar lip-sync) | V2V | ✅ |
| Voiceover & Caption Engine | VOICE | ✅ |
| Brand Kit System | BRAND | ✅ |
| Export & Naming Engine | EXPORT | ✅ |
| Asset Library | LIB | ✅ |
| Team & Collaboration | TEAM | ✅ |
| Ad Account Connector | CONN | V2 |
| Performance Learning Engine (Qvora Signal) | SIGNAL | V2 |
| Platform & Billing | PLAT | ✅ |

---

## [EXT] URL Ingestion & Extraction

**Purpose:** Convert a product URL into structured product data.

**Key requirements:**
- Accept HTTP/HTTPS URLs
- Support Shopify, WooCommerce, custom landing pages, App Store, Google Play
- Headless browser (Playwright on Modal) for JS rendering
- Extract: name, category, price, features, proof points, CTA, image URLs
- Latency ≤ 10s for standard pages
- Fallback to manual text input on failure
- Cache extraction results for 24 hours per URL

**Edge cases:**
- 404 / dead URL → show manual input fallback immediately
- Paywall/login required → detect 401/403, show fallback
- Shopify variants → extract primary variant, show selector if multiple

---

## [BRIEF] Creative Strategy Engine

**Purpose:** Generate creative angles and hooks from extracted product data.

**Key requirements:**
- Generate 3–5 creative angles per product
- Each angle: name, hypothesis, hook (text + spoken), visual direction, CTA
- Inline editing of angles and hooks
- Per-angle and per-hook regeneration without touching other angles
- Brief persisted to DB (fail-fast on insert errors — not silent)

**Current implementation status:** Partially complete. BRIEF-08 (inline edit persistence) and BRIEF-09 (per-angle/per-hook regenerate) are in progress as of April 15, 2026.

---

## [GEN / T2V / V2V] Video Generation Engine

**Purpose:** Turn brief angles into actual video variants.

**Key requirements:**
- T2V via FAL.AI queue (`fal.queue.submit()` — never `fal.subscribe()`)
- Models: Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2
- V2V (avatar lip-sync) via HeyGen Avatar API v3
- Real-time generation progress via SSE (standalone Route Handler, not tRPC)
- Tier-based variant limits enforced in Go middleware

**Model selection guidance:**
- Veo 3.1 — highest quality, slower
- Kling 3.0 — balanced quality/speed
- Runway Gen-4.5 — strong motion quality
- Sora 2 — best for product demo style

---

## [VOICE] Voiceover & Caption Engine

**Purpose:** Add voiceover and captions to generated videos.

**Key requirements:**
- ElevenLabs TTS: `eleven_v3` (quality) or `eleven_flash_v2_5` (~75ms latency)
- Auto-caption generation + timing sync
- Rust Axum + ffmpeg-sys handles transcode/watermark/captions (CPU-bound, Railway)

---

## [BRAND] Brand Kit System

**Purpose:** Store and apply brand voice, colors, and fonts across all ads.

**Key requirements:**
- Upload logo, primary/secondary colors, brand tone description
- Apply brand kit to new briefs automatically
- Tier limits: 1 kit (Starter), 3 kits (Growth), unlimited (Agency)

---

## [EXPORT] Export & Naming Engine

**Purpose:** Export final videos with platform-optimized naming.

**Key requirements:**
- 9:16 aspect ratio for all platforms
- Platform-specific export presets (TikTok, Reels, Meta, YouTube Shorts)
- Naming convention: `[brand]-[angle]-[variant-number]-[platform]`
- Presigned R2 PUT URLs for direct uploads (no server proxy)

---

## [SIGNAL] Performance Learning Engine (V2)

> ⚠️ V2 only. Do not build in V1.

- Connect ad accounts (Meta, TikTok, AppLovin)
- Ingest spend, ROAS, CTR per creative
- Identify winning angles; auto-suggest next-generation variants
- Requires Temporal for scheduling, pgvector for semantic similarity

---

## Open Questions
- [ ] What is the UX for variant limit enforcement? Hard block or warning?
- [ ] Does per-angle regeneration count against the variant limit?

## Related Pages
- [[qvora-overview]] — core flow
- [[personas]] — which features belong to which persona
- [[pricing]] — variant limits per tier
- [[ai-layer]] — AI model choices behind generation
- [[stack-overview]] — technical implementation
