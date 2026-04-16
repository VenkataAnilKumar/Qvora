# Qvora Wiki — Schema & Operating Manual

> This file is the operating manual for this wiki. It tells the LLM (GitHub Copilot) exactly how to maintain, ingest into, query, and lint this wiki. Read this file first in every session before touching any wiki content.

---

## What This Wiki Is

A persistent, compounding knowledge base for **Qvora** — an AI-powered performance creative SaaS. It spans three domains:

1. **Product** — decisions, features, personas, pricing, roadmap, brand
2. **Market** — competitive intelligence, CPM trends, customer research, industry signals
3. **Personal** — founder notes, goals, hypotheses, reflections, lessons learned

It is **not** a RAG index. The LLM writes and maintains the wiki. Knowledge is compiled once and kept current. Cross-references are already there. You never have to rediscover the same thing twice.

---

## Directory Structure

```
wiki/
  SCHEMA.md              ← this file — read first, every session
  index.md               ← content catalog (update on every ingest)
  log.md                 ← append-only chronological log

  product/               ← Qvora product knowledge
    qvora-overview.md    ← one-liner, tagline, elevator pitch, core flow
    features.md          ← feature modules, V1/V2 scope, acceptance criteria
    personas.md          ← ICP, user personas, jobs-to-be-done
    pricing.md           ← tiers, trial, entitlements, Stripe logic
    brand.md             ← brand identity, tokens, fonts, taglines
    roadmap.md           ← phases, current status, what's done/pending

  market/                ← competitive + market research
    market-context.md    ← CPM trends, creative velocity problem, why now
    competitive-landscape.md  ← comparison table, positioning map
    creatify.md          ← Creatify deep-dive (primary competitor)
    arcads.md            ← Arcads profile
    heygen.md            ← HeyGen profile (also used as vendor in stack)
    adcreative.md        ← AdCreative.ai profile

  architecture/          ← technical knowledge
    stack-overview.md    ← full stack summary, non-negotiables
    architecture-decisions.md  ← key ADRs: why Go, why Rust, why two Redis
    ai-layer.md          ← AI models, FAL.AI, ElevenLabs, HeyGen API v3
    data-layer.md        ← PostgreSQL, Redis (TWO instances), R2, Mux

  personal/              ← founder notes, goals, reflections
    (create files as needed, e.g. goals.md, weekly-notes.md, lessons.md)

  synthesis/             ← cross-cutting analysis, filed answers
    (create files as needed; every good answer should be filed here)
```

---

## Page Format

Every wiki page uses this structure:

```markdown
---
title: Page Title
category: product | market | architecture | personal | synthesis
tags: [comma, separated, tags]
sources: [source-file-name-1, source-file-name-2]
updated: YYYY-MM-DD
---

# Page Title

## TL;DR
One to three sentences. What this page says. No fluff.

## [Sections relevant to the topic]

## Open Questions
- [ ] Question that needs a source or research to answer

## Related Pages
- [[linked-page]] — one-line reason for the link
```

**Naming convention:** `kebab-case.md`. No spaces. No uppercase except README/SCHEMA.

---

## Ingest Workflow

When a new source arrives (doc, article, research note, meeting transcript):

1. **Read** the source fully.
2. **Discuss** key takeaways with the user if needed (especially for ambiguous or contradictory content).
3. **Write a summary page** (if the source is significant enough to warrant one) in the appropriate directory.
4. **Update existing pages** touched by the new source — revise claims, add data points, flag contradictions with `> ⚠️ Contradicts: [[page]] — detail`.
5. **Update `index.md`** — add the new page(s), update any existing entries whose summaries changed.
6. **Append to `log.md`** — one entry starting with `## [YYYY-MM-DD] ingest | Source Title`.

A single source may touch 5–15 wiki pages. That's expected. Thoroughness is the point.

### When to create a new page vs. update an existing one
- New entity (competitor, person, concept, product feature) → new page
- Additional data on an existing entity → update that page
- Cross-cutting analysis connecting two or more entities → new page in `synthesis/`

---

## Query Workflow

When the user asks a question:

1. Read `index.md` to find the most relevant pages.
2. Read those pages in full.
3. Synthesize an answer with citations: `([[page-name]])` or `([[page-name#section]])`.
4. **File the answer** if it's non-trivial: create a new page in `synthesis/` and append to `log.md`.
5. If answering required connections the wiki doesn't yet make explicit, update the relevant pages with those cross-references.

---

## Lint Workflow

Periodically (or when the user asks "lint the wiki"):

1. Scan all pages for:
   - **Contradictions** — claims on two pages that conflict (flag with `> ⚠️`)
   - **Stale claims** — data points that newer sources have superseded
   - **Orphan pages** — pages with no inbound links from other wiki pages
   - **Missing pages** — concepts mentioned in multiple pages but lacking their own page
   - **Missing cross-references** — two pages that clearly relate but don't link each other
   - **Data gaps** — open questions that could be answered with a web search

2. Report findings as a lint summary (can be filed in `synthesis/lint-YYYY-MM-DD.md`).
3. Fix issues where possible without a new source; flag the rest as open questions.

---

## Cross-Reference Syntax

Link to other wiki pages using: `[[filename-without-extension]]`  
Link to a section: `[[filename#section-heading]]`  
Flag a contradiction: `> ⚠️ Contradicts: [[page]] — explain why`  
Flag a gap: `> 🔍 Gap: describe what's missing`

VS Code renders `[[wiki-links]]` as plain text (not clickable), but they're grep-able and visually clear. If you install the "Foam" VS Code extension, they become navigable links.

---

## Non-negotiable Product Rules (from copilot-instructions.md)

These apply across all wiki content and any analysis touching the codebase:

1. **Two Redis** — Upstash = HTTP cache only. Railway = asynq TCP only. Never conflate.
2. **SSE ≠ tRPC** — Generation progress stream is a standalone Route Handler.
3. **Tailwind v4 = CSS-only** — No `tailwind.config.ts`. All tokens in `@theme {}`.
4. **HeyGen = v3** — `developers.heygen.com`. V2V lip-sync is v3-only. "v4" does not exist.
5. **ICP = Agency** — DTC (P4) is Phase 2. Never build V1 features for DTC.
6. **Go = I/O bound; Rust = CPU bound** — Don't expand Rust beyond the video postprocessor.
7. **FAL.AI = async queue** — Always `fal.queue.submit()`, never `fal.subscribe()`.

---

## Session Start Checklist

At the start of every wiki session:

1. Read this file (`SCHEMA.md`).
2. Read `index.md` to orient yourself.
3. Read `log.md` (last 10 entries) to understand what's been done recently.
4. Proceed with the user's request.

---

## Tips for This Setup (No Obsidian)

- **Navigation:** Use VS Code's built-in file explorer + `Ctrl+P` (quick open) + `Ctrl+Shift+F` (search across files) to navigate.
- **Graph view substitute:** Run `grep -r "\[\[" wiki/` in terminal to see all cross-references. Or ask the LLM to generate a Mermaid graph of the wiki structure.
- **Search:** VS Code full-text search across the `wiki/` folder is your primary search tool.
- **Foam extension (optional):** Install the "Foam" VS Code extension for `[[wikilink]]` navigation, backlinks panel, and a basic graph view — all without Obsidian.
