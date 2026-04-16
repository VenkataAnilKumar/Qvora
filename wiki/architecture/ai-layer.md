---
title: AI Layer — Models, APIs & Integrations
category: architecture
tags: [AI, FAL.AI, ElevenLabs, HeyGen, GPT-4o, Claude, Langfuse, Vercel-AI-SDK]
sources: [Qvora_Architecture-Stack, Qvora_Implementation-References]
updated: 2026-04-15
---

# AI Layer — Models, APIs & Integrations

## TL;DR
Vercel AI SDK v6 orchestrates all LLM calls. GPT-4o for structured brief extraction. Claude Sonnet 4.6 for creative regeneration. FAL.AI async queue for video. ElevenLabs for TTS. HeyGen v3 for avatar lip-sync. Langfuse for observability.

---

## SDK Foundation

**Vercel AI SDK v6** (`ai` package) is the orchestration layer for all LLM calls:

| Function | Use |
|---|---|
| `generateObject` | GPT-4o structured brief extraction (JSON strict mode) |
| `streamText` | Claude Sonnet 4.6 regen (streaming creative output) |
| `useObject` hook | Streaming UI — shows brief generation in real time |

---

## LLM Models

| Model | Role | Why |
|---|---|---|
| **GPT-4o** | Creative brief generation (angles, hooks, visual direction) | Strongest structured JSON output; reliable schema adherence |
| **Claude Sonnet 4.6** | Hook/angle regeneration, creative drafting | Better creative quality at lower cost; friendlier tone |

**Prompt management:** Langfuse — versioned prompts, A/B testing, cost attribution per workspace.

---

## Video Generation — FAL.AI

FAL.AI is the gateway to multiple T2V/I2V models.

**Rule:** Always `fal.queue.submit()`. Never `fal.subscribe()` (blocks the thread — breaks asynq workers).

### Available Models

| Model | Strength | Use case |
|---|---|---|
| **Veo 3.1** | Highest quality | Brand-quality product demo |
| **Kling 3.0** | Balanced quality/speed | Standard variant generation |
| **Runway Gen-4.5** | Strong motion quality | Dynamic UGC-style |
| **Sora 2** | Best product demo style | Clean, cinematic product showcase |

**I2V (Image → Video):** V2 only. Not in V1 scope.

---

## Text-to-Speech — ElevenLabs

| Model | Latency | Use case |
|---|---|---|
| `eleven_v3` | Standard | Quality voiceovers for final export |
| `eleven_flash_v2_5` | ~75ms | Low-latency preview/draft audio |

175+ languages supported. TTS is rendered first, then synced to generated video or avatar.

---

## Avatar Lip-Sync — HeyGen

**API:** HeyGen Avatar API **v3**  
**Platform:** `developers.heygen.com`  
**Capability:** V2V (Voice-to-Video) lip-sync — avatar speaks the ElevenLabs-generated voiceover

> ⚠️ **Critical constraint:** HeyGen v3 only. V2V lip-sync is a v3 capability. "v4" does not exist. Any docs reference to "v4" is incorrect.

**Flow:**
1. ElevenLabs generates audio (voiceover)
2. HeyGen v3 API takes the audio + avatar selection → returns lip-synced video
3. Rust postprocessor merges, watermarks, transcodes

---

## Observability — Langfuse

- Prompt versioning (compare v1 vs. v2 brief prompts)
- A/B prompt testing
- Cost attribution per workspace (critical for multi-tenant billing accuracy)
- Trace every LLM call with `org_id` + `project_id` tags

---

## Generation Progress — SSE

The generation pipeline (FAL.AI polling → ElevenLabs TTS → HeyGen render) is long-running (minutes). Progress is streamed to the frontend via:

**Standalone Route Handler:** `src/apps/web/app/api/generation/[jobId]/stream/route.ts`

> ⚠️ This is NOT a tRPC subscription. It is a standard Next.js Route Handler using the Web Streams API.

---

## Open Questions
- [ ] Is model selection exposed to the user in V1, or is a default model chosen automatically?
- [ ] What is the approximate cost per video generation (FAL.AI + ElevenLabs + HeyGen)?
- [ ] What are Langfuse's cost attribution groupings — per user, per workspace, per job?

## Related Pages
- [[stack-overview]] — full architecture context
- [[features]] — which modules use which AI components
- [[heygen]] — HeyGen as both vendor and competitor
- [[data-layer]] — where job state and results are stored
