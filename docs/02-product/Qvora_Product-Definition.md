# Product Definition Document
## Qvora — AI Ad Creative Agent
**Version:** 2.1 | **Date:** April 16, 2026 | **Status:** Active (V1 + Post-Launch Roadmap)
**Owner:** Product Management | **Classification:** Confidential

> *"Born to Convert."*  — Qvora Brand Tagline

---

## Table of Contents
1. [Executive Summary](#1-executive-summary)
2. [Brand Identity](#2-brand-identity)
3. [Target Audience & User Personas](#3-target-audience--user-personas)
4. [Core Features & Value Proposition](#4-core-features--value-proposition)
5. [Market Positioning & Competitive Landscape](#5-market-positioning--competitive-landscape)
6. [Technical & Operational Requirements](#6-technical--operational-requirements)
7. [Success Metrics (KPIs)](#7-success-metrics-kpis)
8. [Risks, Dependencies & Mitigation Strategies](#8-risks-dependencies--mitigation-strategies)

---

## 1. Executive Summary

### Product Overview
**Qvora** is an AI-powered performance creative system that transforms a product URL or marketing brief into multiple high-converting short-form video ads — and then continuously learns from real ad performance data to generate better variants over time.

It is not a video editor. It is not an avatar tool. It is a **creative intelligence and production system** purpose-built for performance advertising on TikTok, Instagram Reels, YouTube Shorts, Meta, and Snapchat.

### The One-Liner
> *"Paste a URL. Get 10 video ads. Know which ones win — automatically."*

### Why This Product, Why Now
- Social media CPMs are rising across every major platform (Instagram avg. $9.46, TikTok $4–9, YouTube $4–5 as of 2026). [[Source: Hootsuite, 2026]](https://blog.hootsuite.com/social-media-advertising/)
- Performance teams are running 50–200+ creative variants per month but producing them takes a full creative team.
- No existing product fully closes the **brief → production → performance feedback → next iteration** loop autonomously.
- The fastest-growing unit in performance marketing is **video creative velocity** — the ability to produce, test, and learn from ad variants faster than competitors.
- Creatify (closest existing competitor) manages $650M+ in ad spend and has served 15,000+ brands, proving significant enterprise demand for this category. [[Source: Creatify.ai, April 2026]](https://creatify.ai)

### Strategic Objective
Make **Qvora** the default creative production system for mid-market performance marketing teams — the product they open before launching any paid social campaign.

### MVP Scope Summary
- Input: Product URL or text brief
- Output: 5–10 ready-to-export 9:16 video ads across 3–5 creative angles
- Formats: UGC-style, spokesperson, product demo, voiceover-only
- Platforms: TikTok, Instagram Reels, Meta Feed/Stories, YouTube Shorts

---

---

## 2. Brand Identity

### Name
```
Q  V  O  R  A
│  │  │  │  │
│  │  │  │  └─ Latin suffix — alive, flowing, strength
│  │  │  └──── Latin "vorare" — to devour, dominate
│  │  └─────── Open vowel — confidence, approachability
│  └────────── Velocity, Vision, Victory
└───────────── The rarest letter — unknown, unstoppable
```
> **QVORA** = *"The force that devours mediocre ads"*

### Taglines

| # | Tagline | Tone |
|---|---|---|
| **HERO** | **"Born to Convert."** | Bold, category-defining |
| 2 | "Create bold. Convert faster." | Action, performance |
| 3 | "Feed the algorithm. Starve the competition." | Aggressive, viral |
| 4 | "Where ideas become ads that win." | Transformational |
| 5 | "The last creative tool you'll need." | Category-ending |

### Brand Personality
```
┌──────────────────────────────────────────────────────┐
│  RELENTLESS   SHARP   CONFIDENT   FAST   SMART       │
└──────────────────────────────────────────────────────┘
```

| Trait | In Practice |
|---|---|
| **Relentless** | Never stops optimizing. Always pushing for the next better version. |
| **Sharp** | Precise language. No fluff. Every word earns its place. |
| **Confident** | Makes bold claims — and backs them with data. |
| **Fast** | Speed is a core value. UI feels instant. Copy is punchy. |
| **Smart** | AI-native without being cold. Intelligent without being complex. |

### Brand Voice

| Voice Pillar | Rule | Example |
|---|---|---|
| **Direct** | Say what you mean. Cut the padding. | *"Paste a URL. Get 10 ads."* |
| **Confident** | Own the category. Don't hedge. | *"Your best-performing ad hasn't been made yet. Qvora makes it today."* |
| **Provocative** | Challenge the status quo. | *"You spent 3 days on that ad. It got 0.4% CTR. There's a better way."* |
| **Human** | Smart but never cold. | *"Your creative team is brilliant. Stop making them do boring work."* |

### Visual Identity

**Color Palette**
| Token | Hex | Usage |
|---|---|---|
| **Qvora Black** | `#0A0A0F` | Primary background, hero sections, logo |
| **Qvora Volt** | `#7B2FFF` | Primary brand color, CTAs, highlights |
| **Qvora White** | `#F5F5F7` | Text, clean surfaces |
| **Convert Green** | `#00E87A` | Success states, performance wins |
| **Signal Red** | `#FF3D3D` | Fatigue alerts, warnings, urgency |
| **Data Blue** | `#2E9CFF` | Analytics, charts, performance data |

**Typography Direction**
- Display / Hero: Geometric sans-serif, heavy weight (e.g., Clash Display, Space Grotesk Bold)
- UI / Body: Clean grotesque, regular weight (e.g., Inter, DM Sans)
- Data / Metrics: Monospace / tabular figures (e.g., JetBrains Mono)

**Logo Concept:** Q-as-reticle — the letter Q abstracted as a targeting crosshair / lens, communicating *precision* and *performance targeting*. Functions as app icon, favicon, and watermark.

### Product Naming System
| Product Area | Name | Layer |
|---|---|---|
| Core platform | **Qvora** | Full system |
| Strategy engine | **Qvora Brief** | Intelligence layer |
| Generation engine | **Qvora Studio** | Production layer |
| Analytics engine | **Qvora Signal** | Performance layer |
| API product | **Qvora API** | Developer access |
| Post-launch enterprise package | **Qvora Pro** | Full platform |

### Brand Story
> Performance marketers are not failing because they lack creativity.
> They're failing because the system around creativity is broken.
>
> A brief that takes 3 days. An edit that takes a week. A result you can't read for another week after that.
> By the time you know what worked, the moment is gone.
>
> Qvora was built to devour that delay.
>
> Paste your URL. Get your strategy. Get your ads. Know what wins.
> In minutes, not months.
>
> **Born to Convert.**

---

## 3. Target Audience & User Personas

### Launch ICP — Single Wedge

> **Validation finding (April 2026):** Launching across 4 ICPs simultaneously fragments GTM, messaging, and channel strategy. Performance agencies are the sharpest wedge based on budget, multi-brand data density for V2 learning, and network effects. All launch resources — onboarding, case studies, pricing design, and sales motion — are optimised for this segment first.

**Launch ICP: Performance Marketing Agencies**

| Factor | Evidence |
|---|---|
| Budget | $199–$599/mo sweet spot; no procurement friction below $1K/mo |
| Data density | One agency account = 5–20 brand campaigns = accelerated V2 learning signal |
| Strategy-layer fit | Agencies sell strategy to clients; AI brief generation makes their team faster without reducing perceived value |
| Network effects | Agency tool adoption spreads to peer agencies and client-side teams via word of mouth |
| Willingness to pay | Agencies spend $500–$3,000/mo on creative tools; Qvora at $249–$399 is mid-stack |

**Phase 2 ICP: DTC Brands ($5K–$50K/mo ad spend)**
- Trained buyers: 42% already use AI creative tools
- Follow agencies naturally — often discover Qvora through agency recommendation
- Budget: $100–$500/mo; Growth tier is the natural fit

**Phase 3 ICP: Mobile App UA Managers**
- Requires AppsFlyer/Adjust MMP integration before this segment is fully served
- Defer until V2 (Performance Learning Engine) is live

**Deprioritised at launch:**
- Ecommerce operators overlap entirely with DTC brands — not a distinct segment
- Solo founders: high volume, low ACV, low data density for learning loop

### Primary Segments (Post-Launch Horizon)

| Segment | Launch Phase | Rationale |
|---|---|---|
| **Performance Marketing Agencies** | ✅ Phase 1 — Launch wedge | Highest budget, multi-brand data, network effects |
| **DTC Brands ($5K–$50K/mo spend)** | Phase 2 | Follow agency adoption, trained buyers |
| **Mobile App UA Managers** | Phase 3 | Requires MMP integrations; high volume once unblocked |

---

### Persona 2 — "The Performance Marketer"

**Name:** Alex Chen  
**Role:** Head of Growth / Senior Paid Social Manager  
**Company:** DTC wellness brand, ~$8M ARR  

**Background:**
- Manages $200K–$500K/mo in ad spend across Meta, TikTok, and YouTube
- Currently relies on a freelance video editor + UGC creators for ad content
- Spends 3–5 days per creative cycle from brief to approved ad set
- Tests 20–40 new creative variants per month; needs 80+

**Goals:**
- Get more testable ad variants without hiring more people
- Know faster which angle, hook, or format is winning
- Stop writing the same creative brief for the 50th time

**Frustrations:**
- Tools like Creatify generate video but don't help pick angles or understand the product at a strategic level
- Current workflow: Notion brief → Slack freelancer → revision loop → upload → wait for data → repeat
- Fatigue detection is manual; teams often keep running dead creatives

**Decision Trigger:** Sees a competitor running 3x more ad variants on TikTok with consistent visual quality.

**Willingness to Pay:** $200–$800/mo for a tool that meaningfully accelerates their creative pipeline.

---

### Persona 1 — "The Agency Creative Lead"

**Name:** Jordan Vasquez  
**Role:** Creative Director / Paid Social Lead  
**Company:** Performance marketing agency, 15 brand clients  

**Background:**
- Responsible for creative strategy across all client accounts
- Currently has 2 video editors and 3 content strategists; team is at capacity
- Clients expect 10–20 new ad variants per month each; team can realistically deliver 6–8

**Goals:**
- Scale creative output without scaling headcount
- Make client reporting on creative performance faster and cleaner
- Reduce time from client brief to first usable creative asset

**Frustrations:**
- Generic AI video tools produce content that looks AI-generated and performs poorly
- No tool understands the client's brand voice, product positioning, or competitor landscape
- Clients don't understand why creative strategy takes time; they just want more videos

**Decision Trigger:** A competing agency delivers 3x creative output using AI at the same retainer price.

**Willingness to Pay:** $500–$2,000/mo per agency seat; would pay per-seat or per-output.

---

### Persona 3 — "The App UA Manager"

**Name:** Priya Kapoor  
**Role:** User Acquisition Manager  
**Company:** Mobile gaming studio, 2 live titles  

**Background:**
- Runs Meta, TikTok, and YouTube campaigns for 2 games simultaneously
- Creative testing is core to her job; she tests 50–100 new creatives per month
- Cycle time from "idea" to "live in ad account" is currently 5–7 days minimum

**Goals:**
- Compress creative cycle to 24 hours or less
- Identify which creative frameworks (gameplay hooks, reaction, tutorial) are generating lowest CPI
- Generate batch creative sets at varying lengths (6s, 15s, 30s)

**Frustrations:**
- Creative agencies are slow and expensive
- Internal teams can't keep up with the volume required for meaningful A/B testing
- Performance data is in one tool, creative is in another — no connection between them

**Decision Trigger:** Competing titles are running 3–5 creative refreshes per week; she's running 1.

**Willingness to Pay:** $300–$600/mo; values API access for programmatic creative production.

---

### Assumption Flags
> **[ASSUMPTION]** Willingness-to-pay ranges are estimated from comparable tooling (Creatify: $33–$166/mo; HeyGen: up to enterprise tiers) and adjusted for target segment size. Validation via discovery interviews required before pricing finalization.

---

## 4. Core Features & Value Proposition

### Value Proposition Statement
> **Qvora** replaces the 3-day brief-to-video cycle with a 15-minute automated workflow — delivering strategically sound, platform-native, performance-tested video ad sets that get smarter with every campaign.

### The Five Product Pillars

---

#### Pillar 1: Creative Strategy Engine
**What it does:** Analyzes the product URL or brief to extract offer, audience, and differentiators. Generates multiple creative angles (emotional, functional, social proof, curiosity, problem-aware) with hooks tailored per angle.

**Key capabilities:**
- URL ingestion and automatic product/offer extraction
- AI-generated creative brief (audience, message, angles, hooks)
- Hook library generation: 3–5 variants per angle
- Creative concept labeling (angle type, emotional driver, format recommendation)
- Brief editing interface for human override

**Why it matters:** Most tools start at video generation. We start at creative strategy — ensuring every generated video is grounded in a reason to believe, not just visually passable.

---

#### Pillar 2: Ad Generation Engine
**What it does:** Converts approved briefs and hooks into complete, export-ready video ads across multiple formats.

**Key capabilities:**
- UGC-style (talking head, lifestyle, reaction)
- Spokesperson / AI avatar with natural delivery
- Product demo with voiceover
- Text-on-screen / motion caption formats
- Automatic script generation per hook
- AI voiceover with 20+ voices and tones
- Dynamic caption/subtitle generation
- B-roll sourcing from licensed stock libraries

**Generation controls:**
- Tone selector (energetic, trustworthy, casual, premium)
- Pacing control (fast-cut, standard, storytelling)
- Brand kit integration (colors, logo, fonts)

---

#### Pillar 3: Platform Adaptation Engine
**What it does:** Takes a single generated creative and produces platform-native exports optimized for each target channel.

**Key capabilities:**
- 9:16 vertical (TikTok, Reels, Shorts, Meta Stories)
- 1:1 square (Meta Feed, LinkedIn)
- 16:9 horizontal (YouTube pre-roll, Connected TV)
- Platform-specific safe zone compliance (no text in bottom 20% for TikTok, etc.)
- Auto-resizing and reframing
- Platform-specific caption positioning and sizing
- Ad spec compliance check before export

**Why it matters:** Platform-native formatting is a measurable performance driver. Lo-fi Reels ads drove a 5.6x increase in Thruplays for brands that matched native content style. [[Source: Hootsuite / PureGym case study, 2026]](https://blog.hootsuite.com/social-media-advertising/)

---

#### Pillar 4: Experimentation Engine
**What it does:** Organizes generated ads into structured test sets designed for proper A/B and multivariate testing.

**Key capabilities:**
- Automatic grouping of ads by angle, hook, format, and length
- Test set builder: select control and variant groups
- Concept tagging: every ad labeled by creative variable (hook type, format, emotion)
- Export as structured naming convention (e.g., `brand_angle-social-proof_format-UGC_hook-v1`)
- Bulk export with ad account metadata
- Direct integration with Meta Ads Manager and TikTok Ads Manager (V2)

**Why it matters:** Without structured test sets, performance data is noise. This pillar ensures every export is designed for learning, not just launch.

---

#### Pillar 5: Performance Learning Engine *(V2+)*
**What it does:** Ingests real ad performance data (CTR, CPA, ROAS, hold rate, completion rate) and maps outcomes back to specific creative variables. Uses this data to inform the next generation of briefs and hooks.

**Key capabilities:**
- Ad account connector (Meta, TikTok, Google)
- Per-creative performance dashboard (CTR, CPA, ROAS by ad, angle, format)
- Creative fatigue detection (flagging when CTR starts declining)
- "What's winning" summary report: which angles, hooks, and formats are outperforming
- Next-generation recommendations: "Based on your last 30 days, test these 3 new angles"
- Automated weekly creative refresh suggestions

**Why it matters:** This is the moat. Raw generation is commoditizing. The learning loop — where every campaign makes the next campaign smarter — is what no purely generative tool can replicate.

---

### Feature Prioritization Matrix

| Feature | Phase | Priority | Complexity |
|---|---|---|---|
| URL ingestion + offer extraction | MVP | P0 | High |
| Creative angles + hook generation | MVP | P0 | Medium |
| Script generation per hook | MVP | P0 | Low |
| AI voiceover (20+ voices) | MVP | P0 | Low |
| UGC-style video generation | MVP | P0 | High |
| Text → Video (FAL.AI: Veo 3.1 / Kling 3.0 / Runway Gen-4.5) | MVP | P0 | High |
| Image → Video (Kling 3.0 / Runway / SVD) | MVP | P0 | Medium |
| Voice → Video lip-sync (HeyGen Avatar v3 + ElevenLabs) | MVP | P0 | High |
| Spokesperson/avatar generation | MVP | P0 | High |
| Auto captions/subtitles | MVP | P0 | Low |
| 9:16 export + platform compliance | MVP | P0 | Medium |
| Brand kit (logo, colors, fonts) | MVP | P1 | Medium |
| Manual brief override | MVP | P1 | Low |
| Tone + pacing controls | MVP | P1 | Medium |
| Asset library / ad manager | MVP | P1 | Medium |
| Structured test set export | MVP | P1 | Low |
| Batch generation (10+ at once) | V2 | P0 | Medium |
| Meta + TikTok Ads Manager integration | V2 | P0 | High |
| Performance data ingestion | V2 | P1 | High |
| Creative fatigue detection | V2 | P1 | Medium |
| Localization / multilingual | V2 | P2 | High |
| Competitor ad inspiration library | V3 | P1 | High |
| Autonomous weekly refresh | V3 | P2 | High |

---

## 5. Market Positioning & Competitive Landscape

### Market Context

The global social media advertising market is one of the fastest-growing segments in digital media, with CPMs rising across all major platforms in 2026. Competition for audience attention is intensifying, making **creative quality and velocity** the primary lever for performance marketers.

Key market dynamics:
- Video ads consistently outperform static: **2.7x more leads** vs. image ads [[Source: Meta/Creatify benchmarks, 2026]](https://creatify.ai)
- Short-form video (under 60s) is the dominant ad format on TikTok, Instagram Reels, and YouTube Shorts
- Performance teams testing more variants = lower CPA; creative testing velocity is a competitive moat
- AI-generated content is mainstream — the bar has shifted from "does it look AI?" to "does it convert?"

---

### Competitive Landscape

#### Tier 1: Direct Competitors (Performance-focused AI Ad Tools)

| Product | Core Strength | Core Gap | Funding |
|---|---|---|---|
| **Creatify** | URL-to-video, batch gen, ad analytics, A/B testing, Meta/TikTok launch | Strategy layer is thin; no autonomous learning loop | $24M |
| **AdCreative.ai** | Image ad generation with performance scoring | Weak on video; no creative strategy engine | ~$15M est. |

#### Tier 2: Adjacent Competitors (General AI Video with Ad Use Cases)

| Product | Core Strength | Core Gap |
|---|---|---|
| **HeyGen** | Avatar quality, 175+ language support, enterprise compliance | Not performance-first; no ad analytics, no brief-to-launch workflow |
| **InVideo** | Long-form video AI (30-min from one prompt), bundled Sora/Veo/Kling | General-purpose; no ad creative strategy or performance learning |
| **Chraft** | Viral video specialists, Ad Specialist agent, platform-native angles | Early stage; limited performance data integration |

#### Tier 3: Workflow-Adjacent Tools

| Product | Role | Gap |
|---|---|---|
| **Runway** | World model for cinematic video gen; Characters API | Production tool, not ad workflow tool |
| **Luma AI** | Creative agent for brand campaigns | Enterprise/brand-focused; not performance marketing |
| **Captions** | Editing-first, talking-head optimization | Editing workflow only; no brief-to-creative-to-insight loop |

---

### Positioning Map

```
                    HIGH CREATIVE STRATEGY
                           |
          Luma AI          |        ◉ QVORA  ← us
                           |
GENERAL VIDEO ─────────────┼──────────────── PERFORMANCE AD
                           |
     HeyGen, InVideo       |    Creatify, AdCreative.ai
                           |
                    LOW CREATIVE STRATEGY
```

**Our positioning claim:** The only AI ad creative system that combines **strategic brief generation**, **multi-format video production**, and **performance learning** in a single workflow.

---

### Differentiation Summary

| Capability | **Qvora** | Creatify | HeyGen | InVideo |
|---|:---:|:---:|:---:|:---:|
| Creative strategy / angle engine | ✅ **Qvora Brief** | ⚠️ Partial | ❌ | ❌ |
| URL-to-ad workflow | ✅ | ✅ | ❌ | ❌ |
| Multi-format video generation | ✅ **Qvora Studio** | ✅ | ✅ | ✅ |
| Performance data ingestion | ✅ **Qvora Signal** (V2) | ✅ | ❌ | ❌ |
| Autonomous creative refresh | ✅ (V3) | ❌ | ❌ | ❌ |
| Creative fatigue detection | ✅ (V2) | ⚠️ Basic | ❌ | ❌ |
| Brief editing / human override | ✅ | ⚠️ Limited | ✅ | ✅ |
| Platform-native format compliance | ✅ | ✅ | ⚠️ | ⚠️ |
| Structured test set export | ✅ | ⚠️ | ❌ | ❌ |

> **[ASSUMPTION]** Competitor capability assessments are based on publicly available product pages and self-reported features as of April 2026. Direct product testing should be conducted before finalizing competitive positioning claims.

---

### Pricing Strategy

| Tier | Name | Target | Price | Limits |
|---|---|---|---|---|
| **Starter** | Qvora Starter | Freelance media buyers, small agencies | $99/mo | 20 ads/mo, 1 brand kit, basic formats |
| **Growth** | Qvora Growth | DTC brands, growing agencies | $149/mo | 100 ads/mo, 3 brand kits, all formats, ad account sync |
| **Agency** | Qvora Agency | Mid-market teams, agencies | $399/mo | Unlimited ads, 10 brand kits, Qvora Signal (V2), batch gen, team seats |

> **[VALIDATION UPDATE — April 2026]** Starter repriced from $49 → $99/mo based on competitive benchmarking. Arcads (agency-focused) charges $199–$599/mo; Creatify charges $33–$166/mo but targets SMBs without strategy layer. Qvora's strategy-first differentiation positions it above Creatify parity. $99 is the minimum defensible floor for an agency-first ICP without triggering procurement friction. Requires price sensitivity testing before launch — but $49 undersells the intelligence layer and attracts the wrong ICP (solo operators vs. media buyers).

### Trial & Acquisition Motion

> **Decision (locked):** Qvora runs a **14-day full-access trial** (no credit card required), with generation lock on Day 15 and 30-day data retention after expiry. Conversion lifecycle emails run on Day 5 / Day 10 / Day 15.

---

## 6. Technical & Operational Requirements

### System Architecture Overview

```
[Input Layer]
  URL / Brief / Product Assets
         ↓
[Intelligence Layer — QVORA BRIEF]
  Product Extraction Agent
  Creative Strategy Agent (angles, hooks, briefs)
  Script Generation Agent
         ↓
[Generation Layer — QVORA STUDIO]
  Video Generation (UGC, Avatar, Demo, Motion)
  Voiceover Engine
  Caption/Subtitle Engine
  B-Roll Sourcing
         ↓
[Adaptation Layer — QVORA STUDIO]
  Platform Formatter (9:16 / 1:1 / 16:9)
  Brand Kit Injector
  Compliance Checker
         ↓
[Output Layer]
  Asset Library / Ad Manager
  Structured Export (with naming convention)
  [V2] Ad Account Connector (Meta, TikTok)
         ↓
[Learning Layer — QVORA SIGNAL (V2)]
  Performance Data Ingestor
  Creative Analytics Engine
  Next-Gen Recommendation Engine
```

---

### Core Technical Components

#### 5.1 URL Ingestion & Product Extraction
- **Technology:** Web scraper + LLM-based extraction pipeline
- **Function:** Parse product page → extract product name, category, features, pricing, proof points, CTA
- **Requirement:** Must work on Shopify, WooCommerce, custom landing pages, and app store listings
- **Edge cases:** Handle paywalled pages, JavaScript-heavy SPAs, PDFs (via upload fallback)
- **Latency target:** < 10 seconds for standard product page extraction

#### 5.2 Creative Strategy Engine
- **Technology:** Fine-tuned LLM (GPT-4o class or equivalent) with performance marketing RLHF
- **Function:** Generate creative angles, hook variants, concept rationale
- **Input:** Extracted product data + optional brand brief
- **Output:** Structured creative brief (JSON) with 3–5 angles, 3 hooks per angle, format recommendation
- **Latency target:** < 15 seconds per brief generation

#### 5.3 Video Generation Engine
- **Technology:** Composited pipeline using video generation models (Sora, Veo, Kling, or equivalent) + avatar engine + motion graphics layer
- **Format outputs:** UGC-style (AI avatar or stock talent), product demo (product + voiceover + B-roll), text-motion (copy-focused)
- **Resolution:** 1080p minimum; 4K for Agency tier
- **Latency target:** 60–180 seconds per video at standard quality
- **Cost model:** Generation cost per video must support profitable delivery at Starter tier pricing

#### 5.4 Voiceover Engine
- **Technology:** ElevenLabs API or equivalent TTS provider (multilingual in V2)
- **Voices:** Minimum 20 voices at launch (varied age, gender, accent, tone)
- **Sync:** Accurate lip-sync for avatar-based formats
- **Latency target:** < 5 seconds per 30-second audio generation

#### 5.5 Brand Kit System
- **Inputs:** Logo (SVG/PNG), hex colors (primary/secondary), font files (TTF/OTF), intro/outro bumper video
- **Application:** Auto-applied to all generated videos for that brand
- **Storage:** Per-organization, versioned; supports multiple brand kits per account (Growth+)

#### 5.6 Export & Naming Engine
- **Export formats:** MP4 (H.264 / H.265), MOV; resolution per platform
- **Naming convention:** `{brand}_{angle}_{format}_{hook-variant}_{date}`
- **Metadata:** EXIF-level tagging for platform, angle, format, hook type — enables downstream filtering

#### 5.7 Ad Account Connector (V2)
- **Platforms:** Meta Marketing API, TikTok Marketing API, Google Ads API
- **Function:** Push generated ads directly to ad account; launch campaigns; pull performance metrics
- **Auth:** OAuth 2.0 per platform; token refresh handling; scoped permissions (no budget write access without explicit consent)
- **Data retention:** Performance metrics stored in-product for learning engine training

#### 5.8 Performance Learning Engine (V2)
- **Input:** CTR, CPA, ROAS, video hold rate, completion rate per ad, mapped to creative metadata tags
- **Output:** Per-account statistical model of what creative variables correlate with performance
- **Recommendations:** Generated weekly; confidence-scored; show minimum data threshold before activating
- **Privacy:** No cross-account data sharing; learning is per-organization

---

### Infrastructure Requirements

| Component | Requirement |
|---|---|
| **Cloud Platform** | AWS or GCP (multi-region for latency) |
| **Video Generation** | GPU-backed inference (A100 or H100 class); auto-scale on demand spikes |
| **Storage** | S3-compatible object storage; CDN delivery for exports |
| **Database** | PostgreSQL (relational); Redis (cache/rate limit); vector store (creative metadata search) |
| **Auth** | OAuth 2.0 + PKCE; SSO via SAML 2.0 (post-launch enterprise package) |
| **Security** | SOC 2 Type II (target within 12 months of launch); GDPR compliant; CCPA compliant |
| **Uptime** | 99.5% SLA (Growth/Agency); 99.9% SLA (post-launch enterprise package) |
| **Video Delivery** | CDN-accelerated export delivery; average export download < 30 seconds |

---

### V2 Learning Architecture — V1 Design Constraint

> **[VALIDATION FINDING — April 2026]** Qvora Signal (V2) is the primary competitive moat. If V1 is built without the data schema to support it, V2 becomes a schema rebuild (3–6 month delay) rather than a feature extension. The following data architecture decisions must be locked in V1 even though the learning loop does not activate until V2.

#### Creative Asset Tagging (must be stored in V1 DB schema)
Every generated asset must be tagged at creation time with:

| Tag Field | Values | Purpose |
|---|---|---|
| `angle_type` | awareness / consideration / conversion / retention | Maps to funnel stage for performance correlation |
| `hook_type` | problem / desire / social-proof / shock / curiosity | Hook pattern tracking for CTR prediction |
| `format` | ugc / avatar / product-demo / text-motion | Format-level performance breakdown |
| `emotion` | urgency / aspiration / fear / humor / trust | Emotional register for audience modeling |
| `platform` | meta-feed / meta-story / tiktok / youtube-short | Platform-native format context |
| `brief_id` | UUID of source Qvora Brief | Traces asset back to the strategy that generated it |
| `variant_index` | integer 1–N | Distinguishes hook/copy variants within same angle |

Schema note: tags must be stored as queryable columns (not JSON blob) to support V2 aggregate queries across accounts.

#### Structured Export Naming Convention (enforced in V1)
```
{brand_slug}_{angle_type}_{format}_{hook_variant}_{YYYYMMDD}_v{n}.mp4
```
Example: `acme_conversion_ugc_desire_20260501_v2.mp4`

This naming convention must be applied to all exports at download time. It is the human-readable mirror of the DB tags and enables ad platform upload tracking without API integration.

#### Ad Account Integration Schema (defined in V1, built in V2)
Define the data model in V1 for these fields — even if the Meta/TikTok connector is not built until V2:

| Field | Source (V2) | V1 Requirement |
|---|---|---|
| `ad_id` | Meta / TikTok API | Foreign key column on `assets` table |
| `campaign_id` | Ad platform | Nullable FK; populated when ad account connected |
| `impressions`, `clicks`, `spend` | Ad platform | Nullable columns; schema must exist at launch |
| `ctr`, `cpa`, `roas` | Derived | Computed columns or materialized view |
| `fatigue_detected_at` | V2 Signal engine | Nullable timestamp; column reserved in V1 |

**Why this matters:** The V2 learning engine correlates `angle_type + hook_type + format` tags against `ctr + cpa` signals. If the V1 asset schema does not include tag fields or leaves no FK anchor for ad platform data, V2 requires a destructive migration that risks data loss on existing accounts.

---

### Operational Requirements

- **Content Moderation:** AI-generated content review pipeline to prevent policy-violating ad creative (per platform ToS)
- **Human Review Queue:** Flag edge cases (sensitive categories: health, finance, legal) for human review before delivery
- **Rate Limits:** Per-account generation limits to manage infrastructure cost predictability
- **Audit Logs:** Full generation history per organization; GDPR-compliant deletion on request
- **API Access:** REST API with API key auth for post-launch enterprise package (batch generation, programmatic workflows)

---

## 7. Success Metrics (KPIs)

### North Star Metric
> **Ads launched from Qvora per active team per month**

This metric captures both product utility (teams use it) and business value (teams complete the workflow to launch). It is a leading indicator of paid conversion, retention, and word-of-mouth.

---

### Product KPIs

#### Acquisition & Activation
| Metric | Definition | MVP Target (Month 3) |
|---|---|---|
| Signups | New accounts created | 500/mo |
| Activation Rate | % of signups who generate ≥ 1 ad set | > 60% |
| Time to First Ad | Minutes from signup to first generated ad | < 15 min |
| Brief-to-Video Completion Rate | % of briefs that result in a completed video export | > 75% |

#### Engagement & Retention
| Metric | Definition | Target (Month 6) |
|---|---|---|
| WAU / MAU Ratio | Weekly active users as % of monthly | > 40% |
| Ads Generated per Active Team/Mo | Creative output velocity per account | > 20 ads/mo |
| Export Rate | % of generated ads that are exported/downloaded | > 50% |
| Ad Account Connection Rate (V2) | % of Growth+ accounts that connect an ad account | > 30% |
| D30 Retention | % of activated users still active at day 30 | > 45% |

#### Revenue
| Metric | Definition | Target (Month 12) |
|---|---|---|
| MRR | Monthly Recurring Revenue | $150K |
| Paid Conversion Rate | % of free trialists converting to paid | > 12% |
| Average Revenue Per Account (ARPA) | MRR / paying accounts | > $180/mo |
| Net Revenue Retention (NRR) | Revenue retention including expansion | > 110% |
| Payback Period | Months to recover CAC | < 9 months |

#### Performance Outcome KPIs *(reported via customer surveys and connected ad accounts — V2)*
| Metric | Definition | Target Signal |
|---|---|---|
| Creative Cycle Time Reduction | Days from brief to first launched ad | -60% vs. baseline |
| CTR Delta | CTR of AI-assisted ads vs. prior creative | +20% improvement |
| CPA Delta | CPA of AI-assisted campaigns vs. prior | -15% improvement |
| Creative Testing Velocity | New variants tested per team per month | 3x increase |

---

### Business Health KPIs
| Metric | Target |
|---|---|
| CAC (blended) | < $400 |
| LTV (12-month) | > $2,000 |
| LTV:CAC Ratio | > 4:1 |
| Gross Margin (platform) | > 65% at scale |
| NPS | > 50 by Month 12 |

---

### Leading Indicator Dashboard (Weekly Review)
1. New signups (by channel)
2. Activation rate (7-day cohort)
3. Ads generated (total and per active team)
4. Export rate
5. Paid conversions
6. Churn (trial and paid)
7. Support ticket volume by category (signals friction points)

---

## 8. Risks, Dependencies & Mitigation Strategies

### Risk Register

---

#### Risk 1: URL Extraction Quality is Insufficient
**Category:** Technical  
**Likelihood:** High  
**Impact:** High — the entire workflow depends on accurate product extraction  
**Description:** Many product pages are JavaScript-heavy, dynamically loaded, or obfuscated. Extraction failures produce a bad or empty creative brief, breaking the core value proposition.

**Mitigation:**
- Build a hybrid extraction pipeline: web scraper + structured LLM extraction + fallback manual brief entry
- Support PDF, text paste, and image upload as alternative inputs at launch
- Monitor extraction quality score (completeness of extracted fields) as an operational KPI
- Maintain a test set of 200+ real product URLs across categories for regression testing

---

#### Risk 2: AI-Generated Video Quality Below Performance Threshold
**Category:** Product  
**Likelihood:** Medium  
**Impact:** High — if generated ads don't perform, customers churn immediately  
**Description:** Current open-source and commercial video generation models produce inconsistent quality. UGC-style AI videos may look obviously synthetic, reducing ad performance.

**Mitigation:**
- Invest in a curated quality filter that prevents low-quality frames from being exported
- Use human-in-the-loop review for beta customers to identify and fix quality failure modes
- Offer a hybrid model: AI script/strategy + real UGC creator fulfillment (partnership with UGC marketplace)
- Track "export-to-launch rate" as a proxy for quality satisfaction — if users are generating but not exporting, quality is the blocker

> **[ASSUMPTION]** Video generation quality from current providers (Sora, Veo, Kling, Runway) is sufficient for UGC-style ad formats but may require additional compositing and quality filtering for polished brand-use cases. Requires hands-on testing with target ad formats.

---

#### Risk 3: Creatify Closes the Strategy Gap
**Category:** Competitive  
**Likelihood:** Medium  
**Impact:** High — Creatify has $24M, 15,000+ customers, and strong distribution  
**Description:** Creatify already performs URL-to-video and ad analytics. If they add a meaningful AI creative strategy layer, Qvora's primary differentiation narrows.

**Mitigation:**
- Accelerate **Qvora Signal** (performance learning engine) as the true moat — data network effects take time to build
- Establish direct customer relationships and brand loyalty before Creatify improves
- Differentiate on creative intelligence quality (Qvora Brief), not just video output volume
- Monitor Creatify product releases monthly; maintain a competitive feature parity tracker

---

#### Risk 4: Ad Platform API Access is Restricted or Revoked
**Category:** Technical / Regulatory  
**Likelihood:** Low–Medium  
**Impact:** High for V2 performance learning features  
**Description:** Meta, TikTok, and Google all have terms of service governing third-party ad tool API access. Policy changes, API deprecations, or access revocations could break core V2 features.

**Mitigation:**
- Design performance learning engine to also work via manual CSV data import (platform-export fallback)
- Maintain compliance with all platform partner program terms from day one
- Pursue official Meta Marketing Partner and TikTok Marketing Partner certification
- Do not make the performance learning engine a blocking MVP feature — it is V2, allowing time to build compliant integrations

---

#### Risk 5: Regulators or Platforms Restrict AI-Generated Ad Content
**Category:** Regulatory  
**Likelihood:** Medium (rising)  
**Impact:** Medium — primarily affects certain ad categories (health, finance, political)  
**Description:** The EU AI Act, FTC guidance, and platform policies are increasingly requiring disclosure of AI-generated ad content. New restrictions on AI-generated deepfakes in advertising are emerging.

**Mitigation:**
- Build AI disclosure labeling into every generated export from day one (visible and metadata-level)
- Restrict generation of content in high-risk categories (health claims, financial advice) by default
- Maintain a legal review process for platform policy changes
- Position compliance as a feature, not a burden — "platform-compliant by default"

---

#### Risk 6: Customer Data and Ad Account Security Breach
**Category:** Security  
**Likelihood:** Low  
**Impact:** Critical — ad account access represents real financial exposure for customers  
**Description:** Connecting ad accounts means handling OAuth tokens that have access to customer ad spend. A breach could allow unauthorized spend or data theft.

**Mitigation:**
- Request minimum necessary permissions (read-only data access at V2 launch; write access only after explicit enterprise opt-in)
- OAuth tokens encrypted at rest and in transit; stored in a secrets manager (AWS Secrets Manager or equivalent)
- SOC 2 Type II audit as a 12-month target
- Implement anomaly detection on ad account API calls
- GDPR-compliant data handling; clear data deletion and portability policies

---

#### Risk 7: Generation Cost Exceeds Revenue at Lower Tiers
**Category:** Financial  
**Likelihood:** Medium  
**Impact:** Medium — affects gross margin; could require pricing adjustments  
**Description:** Video generation is compute-intensive. If Starter-tier customers generate at maximum volume, per-unit cost may exceed per-unit revenue.

**Mitigation:**
- Model cost per generated video at each generation quality level before pricing finalization
- Implement hard monthly generation limits per tier enforced at the infrastructure level
- Use queued/async generation (not real-time) for Starter tier to optimize infrastructure cost
- Monitor cost per generated video weekly; raise limits only when unit economics are validated

> **[ASSUMPTION]** Generation cost per 30-second video is estimated at $0.10–$0.50 based on current provider pricing trends. This assumption requires actual integration testing with target video generation providers to confirm before pricing is finalized.

---

### Dependency Map

| Dependency | Type | Owner | Risk Level |
|---|---|---|---|
| Video generation API (Sora/Veo/Kling/Runway) | External | Vendor | High |
| TTS/Voiceover provider (ElevenLabs or equiv.) | External | Vendor | Medium |
| Meta Marketing API | External | Meta | High (V2) |
| TikTok Marketing API | External | TikTok | High (V2) |
| Web scraping infrastructure | Internal | Engineering | Medium |
| LLM provider (OpenAI/Anthropic/equiv.) | External | Vendor | Medium |
| Stock video/image library license | External | Vendor | Low |
| SOC 2 auditor | External | Legal/Ops | Medium |

---

## Appendix

### A. Assumptions Log

| # | Assumption | Section | Validation Method |
|---|---|---|---|
| A-1 | Willingness-to-pay ranges ($200–$800/mo for primary persona) | §2 | User interviews (n=20 minimum) |
| A-2 | Competitor capability assessments based on public product pages | §4 | Direct product testing |
| A-3 | Pricing benchmarked against comparable tools; premium justified | §4 | Price sensitivity survey (Van Westendorp) |
| A-4 | Video generation quality sufficient for UGC-style formats | §7, Risk 2 | Model evaluation against real ad performance |
| A-5 | Generation cost $0.10–$0.50 per 30s video | §7, Risk 7 | Provider API testing + cost modeling |

---

### B. Out of Scope (V1)

The following are explicitly excluded from the MVP to maintain focus:
- Long-form video (> 60 seconds)
- Podcast-to-ad conversion
- Connected TV / OTT ad formats
- Influencer or real-talent UGC marketplace
- Organic social content (non-paid)
- Email or display ad generation
- Analytics dashboards for brand awareness metrics
- Multi-language generation (V2 feature)

---

### C. Glossary

| Term | Definition |
|---|---|
| **Creative Angle** | The strategic frame or emotional driver of an ad (e.g., social proof, problem-aware, aspirational) |
| **Hook** | The first 3 seconds of a video ad; the line or visual designed to stop the scroll |
| **UGC** | User-Generated Content; a video style designed to look like organic content from a real person |
| **CPA** | Cost Per Acquisition; what a brand pays for each customer or conversion |
| **ROAS** | Return On Ad Spend; revenue generated per dollar spent on advertising |
| **CTR** | Click-Through Rate; % of ad impressions that result in a click |
| **Creative Fatigue** | The performance decline that occurs when an audience has seen the same ad too many times |
| **Brief** | A creative strategy document that defines audience, message, angle, and format for an ad |
| **Test Set** | A structured group of ad variants designed for systematic A/B or multivariate testing |

---

### D. References

- Creatify product page and case studies (creatify.ai, April 2026)
- HeyGen product page (heygen.com, April 2026)
- Hootsuite Social Media Advertising Guide 2026 (blog.hootsuite.com)
- Pew Research Center — Social Media Demographics (referenced via Hootsuite)
- Meta / Wistia / HubSpot benchmark data (cited via Creatify case studies)

---

---

### E. Launch Messaging by Segment

| Audience | Launch Line |
|---|---|
| DTC Brand / Growth Marketer | *"Stop briefing freelancers. Start briefing Qvora. Your next ad set is 15 minutes away."* |
| Performance Marketing Agency | *"Deliver 3x the creative output at the same retainer. Your clients will notice. Your competitors will wonder."* |
| Mobile App / UA Manager | *"50 creative variants. One afternoon. Qvora tests while you sleep."* |
| Solo Founder / Ecommerce | *"You don't need a creative team. You need Qvora."* |

---

### F. Competitive Positioning Statement

> **For performance marketers who need video ads that actually convert,**
> **Qvora is the AI creative agent**
> **that goes from product URL to launched ad set in minutes —**
> **and gets smarter with every campaign.**
>
> *Unlike Creatify, which generates video without creative strategy,*
> *Qvora starts with intelligence (Qvora Brief) and ends with insight (Qvora Signal).*

---

*Document prepared by: Product Management Team*
*Product: Qvora — "Born to Convert."*  
*Last updated: April 14, 2026*  
*Next review date: May 14, 2026*  
*Classification: Confidential — Internal Use Only*
