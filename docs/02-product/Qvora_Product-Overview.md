# QVORA
## Product Overview Document
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Foundation Summary  
**Compiled from:** Qvora_Product-Definition.md, Qvora_Brand-Identity.md, Qvora_Feature-Spec.md, Qvora_User-Stories.md

---

## 1. The One-Liner

> *"Paste a URL. Get 10 video ads. Know which ones win — automatically."*

---

## 2. What Qvora Is

**Qvora** is an AI-powered **performance creative system** that transforms a product URL or marketing brief into multiple high-converting short-form video ads — and then continuously learns from real ad performance data to generate better variants over time.

It is not a video editor. It is not an avatar tool. It is a **creative intelligence and production system** purpose-built for performance advertising on TikTok, Instagram Reels, YouTube Shorts, Meta, and Snapchat.

**The problem it solves:**
> Performance marketers are not failing because they lack creativity.  
> They're failing because the system around creativity is broken.  
> A brief that takes 3 days. An edit that takes a week. A result you can't read for another week after that.  
> By the time you know what worked, the moment is gone.

---

## 3. Brand Identity

### Name Meaning
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

### Hero Tagline
> ## *"Born to Convert."*

### Brand Personality
| Trait | In Practice |
|---|---|
| **Relentless** | Never stops optimizing. Always pushing for the next better version. |
| **Sharp** | Precise language. No fluff. Every word earns its place. |
| **Confident** | Makes bold claims — and backs them with data. |
| **Fast** | Speed is a core value. UI feels instant. Copy is punchy. |
| **Smart** | AI-native without being cold. Intelligent without being complex. |

### Visual Identity
| Token | Hex | Usage |
|---|---|---|
| **Qvora Black** | `#0A0A0F` | Primary background, hero sections, logo |
| **Qvora Volt** | `#7B2FFF` | Primary brand color, CTAs, highlights |
| **Qvora White** | `#F5F5F7` | Text, clean surfaces |
| **Convert Green** | `#00E87A` | Success states, performance wins |
| **Signal Red** | `#FF3D3D` | Fatigue alerts, warnings, urgency |
| **Data Blue** | `#2E9CFF` | Analytics, charts, performance data |

**Typography:**
- Display / Hero: Geometric sans-serif, heavy weight (e.g., Clash Display, Space Grotesk Bold)
- UI / Body: Clean grotesque, regular weight (e.g., Inter, DM Sans)
- Data / Metrics: Monospace (e.g., JetBrains Mono)

**Logo:** Q-as-reticle — the letter Q abstracted as a targeting crosshair / lens, communicating *precision* and *performance targeting*.

---

## 4. Product Architecture

Qvora is composed of four named product layers:

| Layer | Product Name | V1? | Description |
|---|---|---|---|
| Strategy engine | **Qvora Brief** | ✅ V1 | Generates creative angles, hooks, and briefs from a URL |
| Production engine | **Qvora Studio** | ✅ V1 | Converts briefs into multi-format video ads |
| Learning engine | **Qvora Signal** | V2 | Ingests ad performance data to improve future briefs |
| Developer access | **Qvora API** | V2 | Programmatic access for batch workflows and integrations |

---

## 5. Core Product Pillars

### Pillar 1 — Creative Strategy Engine (Qvora Brief)
**What it does:** Analyzes the product URL or brief → extracts offer, audience, differentiators → generates multiple creative angles with hooks tailored per angle.

**Key outputs:**
- 3–5 creative angles per product (emotional, functional, social proof, curiosity, problem-aware)
- 3 hook variants per angle (problem / desire / social-proof / shock / curiosity)
- Format and platform recommendation per brief
- Funnel stage labeling (awareness / consideration / conversion / retention)
- Fully editable brief with version history

**Why it matters:** Most tools start at video generation. Qvora starts at creative strategy — every generated video is grounded in a reason to believe, not just visually passable.

---

### Pillar 2 — Ad Generation Engine (Qvora Studio)

**Brief-driven formats (from approved angle + hook):**
- UGC-style (talking head, lifestyle, reaction)
- Spokesperson / AI avatar with natural delivery
- Product demo with voiceover
- Text-on-screen / motion caption

**AI Generation Modes (direct primitives):**

| Mode | Input | Output | Primary Model | V1? |
|---|---|---|---|---|
| **Text → Video** | Script or text prompt | Cinematic video clip | Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2 (via FAL.AI) | ✅ V1 |
| **Image → Video** | Product image + motion prompt | Animated clip from static image | Kling 3.0 / Runway Gen-4 | V2 |
| **Voice → Video** | Uploaded audio or cloned voice | Lip-synced avatar video | HeyGen Avatar v3 + ElevenLabs | ✅ V1 |

- **Text → Video (V1):** Brief script → cinematic product shot, B-roll, or brand story; model selector (Auto / Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2); 9:16 aspect ratio enforced
- **Image → Video (V2):** Upload a product hero image → animate with motion (zoom, rotate, pan, splash); auto-sources from product page extraction
- **Voice → Video (V1):** Upload your own voice or use a cloned brand voice → lip-synced to a chosen avatar via HeyGen Avatar v3; 175+ languages

**Controls (all formats):**
- Tone selector (energetic, trustworthy, casual, premium)
- Pacing control (fast-cut, standard, storytelling)
- Brand kit integration (colors, logo, fonts auto-applied)
- 20+ AI voiceover voices (ElevenLabs); voice cloning for brand voice (Growth+)
- Dynamic caption / subtitle generation
- B-roll sourcing from licensed stock libraries

---

### Pillar 3 — Platform Adaptation Engine
**What it does:** Takes a single generated creative and produces platform-native exports optimized for each target channel.

| Format | Platforms |
|---|---|
| 9:16 vertical | TikTok, Instagram Reels, YouTube Shorts, Meta Stories |
| 1:1 square | Meta Feed, LinkedIn |
| 16:9 horizontal | YouTube pre-roll, Connected TV |

**Built-in compliance:**
- Platform-specific safe zone enforcement (e.g., no text in bottom 20% for TikTok)
- Auto-resizing and reframing
- Ad spec compliance check before export

---

### Pillar 4 — Experimentation Engine
**What it does:** Organizes generated ads into structured test sets designed for proper A/B and multivariate testing.

- Automatic grouping by angle, hook, format, and length
- Concept tagging: every ad labeled by creative variable
- Structured export naming: `brand_angle-type_format_hook-variant_date_vN.mp4`
- Bulk export with ad account metadata
- Direct integration with Meta Ads Manager and TikTok Ads Manager (V2)

---

### Pillar 5 — Performance Learning Engine (V2) — Qvora Signal
**What it does:** Ingests real ad performance data (CTR, CPA, ROAS, hold rate, completion rate) and maps outcomes back to specific creative variables. Uses this data to inform the next generation of briefs and hooks.

- Per-creative performance dashboard
- Creative fatigue detection (flags when CTR starts declining)
- "What's winning" summary: which angles, hooks, and formats are outperforming
- Next-gen recommendations: *"Based on your last 30 days, test these 3 new angles"*
- Automated weekly creative refresh suggestions

**Why it's the moat:** Raw generation is commoditizing. The learning loop — where every campaign makes the next campaign smarter — is what no purely generative tool can replicate.

---

## 6. MVP Scope

| Dimension | MVP Specification |
|---|---|
| **Input** | Product URL, text brief, product image, or audio file |
| **Output** | 5–10 ready-to-export video ads across 3–5 creative angles |
| **Formats** | UGC-style, spokesperson, product demo, voiceover-only, Text→Video, Voice→Video (Image→Video: V2) |
| **Platforms** | TikTok, Instagram Reels, Meta Feed/Stories, YouTube Shorts |
| **Time to first ad** | < 15 minutes from signup |
| **Generation latency** | Brief: < 15s; UGC/Demo: 60–180s; T2V: 30–120s; I2V: 20–60s; V2V: 60–180s |
| **Export** | MP4 (H.264/H.265), with structured naming convention + manifest.csv |

### Feature Priority Summary

| Feature | Phase | Priority |
|---|---|---|
| URL ingestion + product extraction | MVP | P0 |
| Creative angles + hook generation | MVP | P0 |
| Script generation per hook | MVP | P0 |
| AI voiceover (20+ voices) | MVP | P0 |
| UGC-style video generation | MVP | P0 |
| Text → Video (FAL.AI: Veo 3.1 / Kling 3.0 / Runway) | MVP | P0 |
| Image → Video (Kling 3.0) | V2 | P0 |
| Voice → Video lip-sync (HeyGen Avatar v3) | MVP | P0 |
| Voice cloning for brand voice (Growth+) | MVP | P1 |
| 9:16 / 1:1 / 16:9 export + platform compliance | MVP | P0 |
| Brand kit (logo, colors, fonts) | MVP | P1 |
| Structured test set export + manifest.csv | MVP | P1 |
| Batch generation (10+ at once) | V2 | P0 |
| Meta + TikTok Ads Manager integration | V2 | P0 |
| Performance data ingestion (Signal) | V2 | P1 |
| Creative fatigue detection | V2 | P1 |
| V2V multi-language translation pipeline | V2 | P1 |

---

## 7. Target Audience

### Launch ICP — Performance Marketing Agencies

| Factor | Detail |
|---|---|
| Budget | $199–$599/mo sweet spot; no procurement friction below $1K/mo |
| Team size | 5–50 person agencies managing 3–15 brand clients |
| Ad spend managed | $50K–$500K/mo across Meta, TikTok, YouTube |
| Data density | One agency account = 5–20 brand campaigns → accelerated V2 learning signal |
| Network effects | Agency tool adoption spreads to peer agencies and client-side teams |

### Primary Personas

#### Persona 2 — "The Performance Marketer" (Alex Chen)
- Head of Growth / Senior Paid Social Manager at a DTC brand (~$8M ARR)
- Manages $200K–$500K/mo in ad spend across Meta, TikTok, YouTube
- Current cycle: 3–5 days from brief to approved ad set
- Tests 20–40 variants/month; needs 80+
- **Decision trigger:** Sees competitor running 3x more ad variants on TikTok

#### Persona 1 — "The Agency Creative Lead" (Jordan Vasquez)
- Creative Director managing creative strategy across 15 brand clients
- Team at capacity: 2 video editors + 3 content strategists
- Clients expect 10–20 new variants/month each; team delivers 6–8
- **Decision trigger:** Competing agency delivers 3x creative output at the same retainer price

#### Persona 3 — "The App UA Manager" (Priya Kapoor)
- User Acquisition Manager at a mobile gaming studio with 2 live titles
- Tests 50–100 new creatives per month
- Current cycle: 5–7 days from "idea" to "live in ad account"
- **Decision trigger:** Competing titles running 3–5 creative refreshes per week; she runs 1

### ICP Expansion Roadmap
| Phase | Segment | Trigger |
|---|---|---|
| **Phase 1 — Launch** | Performance marketing agencies | Highest budget + data density |
| **Phase 2** | DTC brands ($5K–$50K/mo spend) | Follow agency adoption |
| **Phase 3** | Mobile App UA Managers | Requires MMP integration (V2+) |

---

## 8. User Journey

### Epic Structure
```
EPIC 1 — Onboarding & Activation
EPIC 2 — Qvora Brief (Strategy & Intelligence Layer)
EPIC 3 — Qvora Studio (Video Generation & Production Layer)
EPIC 4 — Export & Structured Testing
EPIC 5 — Brand Kit & Multi-Brand Management
EPIC 6 — Team & Collaboration
EPIC 7 — Qvora Signal (Performance Learning — V2)
EPIC 8 — Platform & Administration
```

### Activation Goal
> Get any agency user to their first complete ad set in **under 15 minutes**. Create a real artifact in the **first 60 seconds** — not a tutorial.

### First-Run Flow
1. Signup → role selection (Media Buyer / Creative Director / Brand Manager)
2. Brand setup wizard (name, color, logo — completable in < 2 minutes)
3. Paste product URL → brief generated in 15 seconds
4. Review / edit angles and hooks
5. Select formats → video generation begins
6. Preview, adjust, export ad set

---

## 9. Pricing Strategy

| Tier | Price | Target | Limits |
|---|---|---|---|
| **Starter** | $99/mo | Freelance media buyers, small agencies | 20 ads/mo, 1 brand kit, 1 seat, 3 variants/angle |
| **Growth** | $149/mo | DTC brands, growing agencies | 100 ads/mo, 3 brand kits, 5 seats, 10 variants/angle, voice cloning |
| **Agency** | $399/mo | Mid-market teams, agencies | Unlimited ads, unlimited brand kits, unlimited seats, 4K export, custom avatar (V2) |

> **Trial & Acquisition Motion:** Every major competitor offers a free tier. Qvora's documented acquisition motion is a **7-day full-access trial** (no credit card; limit 3 exported ad sets). Creates a real working ad set in the first session — Qvora's strongest conversion argument. See [Qvora_Product-Definition.md](Qvora_Product-Definition.md) §5 for full rationale and alternatives.

---

## 10. Technical Architecture

```
[Input Layer]
  URL / Brief / Product Image / Audio File
         ↓
[Intelligence Layer — QVORA BRIEF]
  Product Extraction Agent (Playwright/Puppeteer)
  Creative Strategy Agent (LLM — GPT-4o class)
  Script Generation Agent
         ↓
[Generation Layer — QVORA STUDIO]
  ┌── Brief-Driven Formats ──────────────────────────────┐
  │   UGC-style · Product Demo · Text-Motion             │
  └──────────────────────────────────────────────────────┘
  ┌── AI Generation Modes ──────────────────────────────┐
  │   Text → Video  (FAL.AI: Veo 3.1 / Kling 3.0 /     │
  │                  Runway Gen-4.5 / Sora 2)           │
  │   Image → Video (Kling 3.0 / Runway / SVD)          │
  │   Voice → Video (HeyGen Avatar v3 + ElevenLabs)     │
  └──────────────────────────────────────────────────────┘
  Voiceover Engine (ElevenLabs — 20+ voices; voice clone Growth+)
  Caption / Subtitle Engine (auto-transcription)
  B-Roll Sourcing (licensed stock)
         ↓
[Adaptation Layer]
  Platform Formatter (9:16 / 1:1 / 16:9)
  Brand Kit Injector
  Ad Spec Compliance Checker
         ↓
[Output Layer]
  Asset Library / Ad Manager
  Structured Export (naming convention enforced)
  [V2] Ad Account Connector (Meta, TikTok, Google)
         ↓
[Learning Layer — QVORA SIGNAL (V2)]
  Performance Data Ingestor
  Creative Analytics Engine (CTR/CPA → creative variable correlation)
  Next-Gen Brief Recommendation Engine
```

### Key Infrastructure Requirements
| Component | Specification |
|---|---|
| Cloud | AWS or GCP (multi-region) |
| Video Generation | GPU-backed inference (A100/H100 class); auto-scale |
| Storage | S3-compatible object storage + CDN delivery |
| Database | PostgreSQL + Redis (job queue) + vector store |
| Auth | OAuth 2.0 + PKCE; SAML 2.0 (Enterprise) |
| Security | SOC 2 Type II (target within 12 months); GDPR + CCPA compliant |
| SLA | 99.5% (Growth/Scale); 99.9% (Enterprise) |

---

## 11. V1→V2 Data Architecture Note

> **Critical:** Qvora Signal (V2) is the primary competitive moat. The V1 schema must be built to support it now, or V2 requires a destructive migration.

Every generated asset must be tagged at creation time in V1:

| Tag | Values | Purpose |
|---|---|---|
| `angle_type` | awareness / consideration / conversion / retention | Funnel stage for performance correlation |
| `hook_type` | problem / desire / social-proof / shock / curiosity | Hook pattern → CTR prediction |
| `format` | ugc / avatar / product-demo / text-motion | Format-level performance breakdown |
| `emotion` | urgency / aspiration / fear / humor / trust | Emotional register for audience modeling |
| `platform` | meta-feed / meta-story / tiktok / youtube-short | Platform context |
| `brief_id` | UUID → source Qvora Brief | Traceability back to strategy |

---

## 12. Success Metrics

### North Star Metric
> **Ads launched from Qvora per active team per month**

### Key KPI Targets

| Metric | MVP Target (Month 3) | Month 12 Target |
|---|---|---|
| New signups/mo | 500 | — |
| Activation rate (brief → export) | > 60% | > 75% |
| Time to first ad | < 15 min | < 10 min |
| WAU/MAU ratio | > 40% | > 55% |
| D30 retention | > 45% | > 60% |
| MRR | — | $150K |
| Paid conversion rate | > 12% | > 18% |
| ARPA | — | > $180/mo |
| NRR | — | > 110% |
| NPS | — | > 50 |
| Creative cycle time reduction | -40% | -60% |
| CTR delta (AI-assisted vs. prior) | — | +20% |

---

## 13. Key Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| URL extraction quality insufficient | High | High | Hybrid pipeline + manual fallback + extraction quality scoring |
| AI video quality below performance threshold | Medium | High | Quality filter + human review in beta + real UGC hybrid option |
| Creatify closes the strategy gap | Medium | High | Accelerate Signal (data moat); lock customer relationships early |
| Video generation cost erodes margin | Medium | High | Cost-per-video benchmarking at Starter tier before launch |
| Data privacy violation (ad account ingestion) | Low | Critical | OAuth scoped permissions; no cross-account data sharing |

---

## 14. Roadmap Summary

| Horizon | Milestone | Key Deliverable |
|---|---|---|
| **V1 — MVP** | Launch | Brief engine + video generation + export + brand kit + team workspace |
| **V2 — +6 months** | Performance Loop | Ad account connectors + Qvora Signal + fatigue detection + batch gen |
| **V3 — +12 months** | Intelligence Moat | Autonomous weekly refresh + competitor ad inspiration + Qvora API GA |

---

## 15. References

| Document | Description |
|---|---|
| [Qvora_Product-Definition.md](Qvora_Product-Definition.md) | Full product definition, technical specs, KPIs, risk register |
| [Qvora_Brand-Identity.md](../01-brand/Qvora_Brand-Identity.md) | Brand essence, visual identity, voice and tone guidelines |
| [Qvora_Feature-Spec.md](../04-specs/Qvora_Feature-Spec.md) | Engineering-handoff feature specifications, acceptance criteria |
| [Qvora_User-Stories.md](../04-specs/Qvora_User-Stories.md) | Full epic and user story backlog with acceptance criteria |
| [Qvora_User-Journey.md](../04-specs/Qvora_User-Journey.md) | End-to-end user journey, onboarding flow, activation milestones |
| [Qvora_Competitive-Analysis.md](../03-market/Qvora_Competitive-Analysis.md) | Market research, live competitor data, pricing landscape |
