# CLAUDE.md — AI operating rules for the Gigmann Executive Cockpit

This file defines how AI assistants (and humans) work in this repository. It is mandated by the company
AI-Native Engineering manuals. Read it fully before making changes.

## 1. What this project is
The **Gigmann Executive Cockpit ("Ahenfie")** — an AI-native executive "chief of staff" for a 12-facility
healthcare network in Ghana. The hero feature is the **Daily Brief**. A **deterministic signal engine computes
all numbers**; Claude only **narrates and prioritises** them. **The AI never invents a figure.**

## 2. Source of truth & traceability
- **`agent_plan.md` replaces Jira.** Every change maps to a story `GEC-###`.
- Read `agent_plan.md` §0 (conventions), §2 (stack), §3 (Definition of Done), §4 (cross-cutting baselines)
  before coding.
- Update the story status and the progress dashboard in the **same PR** that does the work.

## 3. Architecture rules (hexagonal / ports & adapters) — non-negotiable
- `internal/core/**` = **domain**. Pure Go. **No** imports of frameworks, DB drivers, HTTP, or
  `internal/adapters`. Enforced by `internal/architecture/arch_test.go`.
- `internal/ports/**` = interfaces the application depends on.
- `internal/app/**` = **use cases**. Orchestrates ports. Holds authorization decisions. No concrete adapters.
- `internal/adapters/inbound/**` = HTTP (Chi) and other entrypoints; thin, translate to/from app.
- `internal/adapters/outbound/**` = Postgres (pgx+sqlc), Redis, Anthropic, etc. Implement ports.
- `cmd/**` = composition root (wiring only).
- Dependencies point **inward**: adapters → app → ports → core. Never the reverse.

## 4. Coding standards
- Go: `gofmt`/`goimports` clean; pass `go vet` and `golangci-lint`. Wrap errors with `%w` and context.
  Money in **minor units** (never float). Use `log/slog` (structured); **never log PII or secrets**.
- Frontend: TypeScript strict; React function components + hooks; MUI v9 theming; **skeleton** loaders for
  content, **animated-dot** loaders for buttons; **paginate** growable lists. Honour `prefers-reduced-motion`.
- Parameterised SQL only. Validate all inputs at the app boundary (allow-list).

## 5. Testing & quality gates (every push)
- **Coverage must stay > 80%** (`make backend-cover-gate`; frontend `npm run test:coverage`).
- **SonarQube quality gate must pass.**
- Domain + signal engine should approach ~100% coverage.
- Write tests with the change, not after.

## 6. GitHub workflow
- Branch: `feature/GEC-123-short-description` (also `fix/`, `chore/`, `docs/`).
- Commit: `GEC-123 implement X`. PR title: `GEC-123 Short summary`.
- A merged PR is the only thing that flips a story to Done; the PR must update `agent_plan.md`.

## 7. Security (see `agent_plan.md` E9 + §4.1)
- Secrets only via environment / Render env groups. Never commit secrets; secret-scanning runs in CI.
- Enforce authz in the app layer; managers are scoped to their facility (no IDOR).
- Treat all user/NL input as untrusted; AI output must never trigger side-effects without explicit user confirm.

## 8. AI operating procedures
- Break work into stories/subtasks; keep changes aligned to the referenced `GEC-###`.
- **Do not modify unrelated code.** Keep diffs minimal and reviewable.
- Update docs (`agent_plan.md`, ADRs, this file, `AGENTS.md`) when behaviour or workflow changes.
- When unsure about an external library's current API, check its docs before writing code — do not guess.
- Prefer the deterministic path: compute in code, narrate with the model.

## 9. Useful commands
```
make help                 # list tasks
make backend-test         # tests + coverage
make backend-cover-gate   # enforce >80%
make backend-run          # run API
make dev-up               # local Postgres 16 + pgvector + Redis
```
