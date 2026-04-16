# Changelog

All notable changes to this project are documented in this file.

## [0.1.0] - 2026-04-14

Initial repository foundation and documentation baseline for Qvora.

### Added
- Core project context document for fast onboarding: [CONTEXT.md](.github/CONTEXT.md)
- Decision log and long-memory reference: [MEMORY.md](.github/MEMORY.md)
- Agent runtime guidance for multi-agent tooling: [AGENTS.md](.github/AGENTS.md)
- GitHub Copilot workspace instructions: [.github/copilot-instructions.md](.github/copilot-instructions.md)
- Product-facing repository overview and setup guide: [README.md](README.md)

### Documentation
- Product definition and overview docs aligned for V1 scope and ICP in [docs/02-product](docs/02-product)
- Functional requirements and acceptance criteria baseline in [docs/04-specs/Qvora_Feature-Spec.md](docs/04-specs/Qvora_Feature-Spec.md)
- Epic and persona-based user stories baseline in [docs/04-specs/Qvora_User-Stories.md](docs/04-specs/Qvora_User-Stories.md)
- Technical architecture reference consolidated in [docs/06-technical/Qvora_Architecture-Stack.md](docs/06-technical/Qvora_Architecture-Stack.md)
- Repository structure decision and rationale in [docs/06-technical/Qvora_Repo-Structure.md](docs/06-technical/Qvora_Repo-Structure.md)
- Implementation references catalog in [docs/07-implementation/Qvora_Implementation-References.md](docs/07-implementation/Qvora_Implementation-References.md)
- 8-phase implementation tracker with delivery gates in [docs/07-implementation/Qvora_Implementation-Phases.md](docs/07-implementation/Qvora_Implementation-Phases.md)
- Task-level implementation checklist in [docs/07-implementation/IMPLEMENTATION_CHECKLIST.md](docs/07-implementation/IMPLEMENTATION_CHECKLIST.md)

### Architecture Decisions Locked
- Two-Redis model enforced: Upstash for HTTP cache/rate limiting, Railway Redis for asynq TCP queues
- SSE generation stream defined as standalone Next.js Route Handler (not tRPC subscription)
- Tailwind v4 CSS-only token strategy with `@theme {}` in globals
- HeyGen version locked to v3
- V1 scope locked to agency personas; DTC deferred to Phase 2

### Repository Status
- Repository remains private and internal-use only
- Documentation-first baseline complete; implementation scaffolding planned per phased execution
