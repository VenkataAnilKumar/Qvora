# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| `main`  | ✅        |

## Reporting a Vulnerability

If you discover a security vulnerability in Qvora, please **do not** open a public GitHub issue.

Instead, report it privately:

1. **GitHub Security Advisories (preferred):**  
   Go to [Security → Report a vulnerability](../../security/advisories/new) and submit a private advisory.

2. **Email:**  
   Send details to the repository owner via GitHub profile contact.

Please include:

- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept (if applicable)
- Any suggested mitigations

## Response Timeline

| Action | Target |
|--------|--------|
| Acknowledgement | Within 48 hours |
| Initial assessment | Within 5 business days |
| Fix or mitigation | Depends on severity (critical: ≤7 days, high: ≤14 days) |

## Scope

The following are in scope:

- `src/apps/web` — Next.js frontend + tRPC BFF
- `src/services/api` — Go Echo REST API
- `src/services/worker` — Go asynq job workers
- `src/services/postprocess` — Rust Axum video postprocessor
- Authentication flows (Clerk integration)
- Data access controls (Supabase RLS)

Out of scope: third-party services (Clerk, Supabase, FAL.AI, etc.) — report those to their respective vendors.
