---
title: Pricing & Entitlements
category: product
tags: [pricing, tiers, trial, stripe, entitlements]
sources: [Qvora_Product-Definition]
updated: 2026-04-15
---

# Pricing & Entitlements

## TL;DR
Three tiers (Starter $99 / Growth $149 / Agency $399), 7-day free trial with no credit card, generation locked on Day 8. Tier limits enforced server-side in Go middleware — never client-side.

---

## Tiers

| Tier | Price | Variant Limit | Target Persona |
|---|---|---|---|
| **Starter** | $99/mo | Max 3 variants / angle | Solo media buyer or small agency |
| **Growth** | $149/mo | Max 10 variants / angle | Growth-stage agency, 3–8 clients |
| **Agency** | $399/mo | Unlimited | Full mid-market agency, 8–15+ clients |

---

## Trial Logic

| Day | State |
|---|---|
| Day 1–7 | Full access, no credit card required |
| Day 8 | **Generation locked.** Data and exports still accessible. |
| Day 8–37 | 30-day data retention window. |
| Day 38+ | Data purged if no conversion. |

### Conversion Email Schedule
- **Day 3:** "You've created X ads — here's what's possible with more variants"
- **Day 6:** "2 days left. Lock in your work."
- **Day 8:** "Generation paused. Don't lose your briefs."

---

## Entitlements

Enforced via **Stripe Entitlements API** + Go API middleware. Never enforced client-side only.

| Entitlement | Starter | Growth | Agency |
|---|---|---|---|
| Variants per angle | 3 | 10 | Unlimited |
| Workspaces per org | 1 | 3 | Unlimited |
| Brand kits | 1 | 3 | Unlimited |
| Team members | 2 | 5 | Unlimited |
| Export formats | Standard | Standard + HD | All |
| Priority queue | ❌ | ❌ | ✅ |

> 🔍 Gap: Confirm exact entitlement names used in Stripe Entitlements API (need to map to feature IDs in Go middleware).

---

## Stripe Integration

**Subscription statuses to handle:**

| Status | Behavior |
|---|---|
| `trialing` | Full access, no card on file |
| `active` | Full access per tier |
| `past_due` | Warn in UI; generation continues for grace period (7 days) |
| `canceled` | Generation locked; data retained for 30 days |

**Key webhooks:**
- `invoice.paid` — activate/upgrade
- `customer.subscription.updated` — tier change, downgrade enforcement
- `customer.subscription.deleted` — lock generation, start retention countdown

---

## Open Questions
- [ ] Should Agency tier have SLA / priority support, or just priority queue?
- [ ] What's the expected upgrade path: Starter → Growth → Agency or mostly direct to tier?

## Related Pages
- [[personas]] — which tier maps to which persona
- [[features]] — feature availability by tier
- [[qvora-overview]] — strategic objective
- [[stack-overview]] — Stripe integration in the Go API
