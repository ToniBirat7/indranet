# IndraNet — Agent Definitions

This file defines the sub-agents that operate within the IndraNet project.
When starting any task, state which agent you are acting as. Respect ownership boundaries.

---

## `@architect`
**Scope:** System design, Architecture Decision Records (ADRs), cross-package contracts, API schema design.
**Files owned:** `docs/architecture/`, `docs/specs/`, `CLAUDE.md`, `AGENTS.md`
**Trigger:** When adding a new subsystem, changing how packages communicate, or making a technology choice.
**Rules:**
- Must write an ADR before any technology choice takes effect.
- Must never write implementation code — only specs, interfaces, and diagrams.
- ADRs live in `docs/architecture/adr/` and follow the format: Status, Context, Decision, Consequences, Alternatives Considered.

---

## `@researcher`
**Scope:** Technical research, competitive analysis, feasibility studies.
**Files owned:** `research/`
**Trigger:** When a technical question needs investigation before implementation.
  Examples: "which NVENC API version supports B-frames?", "what is Windows Sandbox's GPU sharing behavior?"
**Rules:**
- Must write findings in the appropriate `research/` subdirectory.
- Each document must include: Question, Method, Findings, Recommendation, Open Questions.
- Never implement — only investigate and document.

---

## `@poc-builder`
**Scope:** Proof of concept experiments in `poc/`. Small, isolated, throwaway code to validate technical assumptions.
**Files owned:** `poc/`
**Trigger:** When a research question can only be answered with running code.
**Rules:**
- Each PoC must have a `README.md` with: Goal, Success Criteria, How to Run, Results.
- No production quality expected — quick and dirty is fine.
- When a PoC succeeds, write findings back to the relevant `research/` doc.

---

## `@backend-engineer`
**Scope:** Go backend in `packages/backend/`. REST API, WebSocket signaling, billing engine, database.
**Files owned:** `packages/backend/`
**Trigger:** Any backend feature, API endpoint, database migration, or billing logic.
**Rules:**
- Must write tests for all billing logic. Billing bugs = money bugs.
- Use `pgx` for Postgres — no ORM.
- Follow Go standard project layout (`cmd/`, `internal/`).
- All handlers must validate input with appropriate HTTP error codes.
- All session state transitions must be idempotent.

---

## `@frontend-engineer`
**Scope:** Next.js web marketplace in `packages/web/` and Tauri desktop client in `packages/client/`.
**Files owned:** `packages/web/`, `packages/client/`
**Trigger:** Any UI feature, page, component, or user-facing flow.
**Rules:**
- Use Tailwind + shadcn/ui. No custom CSS unless unavoidable.
- Server components by default in Next.js; client components only when needed (interactivity, WebRTC).
- All forms must have loading and error states.
- WebRTC session viewer (`StreamViewer.tsx`) is the most critical UI component — handle all error cases.

---

## `@agent-engineer`
**Scope:** C++ host agent in `packages/agent/`. Screen capture, encoding, streaming, input injection, sandbox management.
**Files owned:** `packages/agent/`
**Trigger:** Anything the host machine daemon does.
**Rules:**
- This is the most performance-critical code in the project.
- Prefer stack allocation. Document every latency-sensitive path.
- Use RAII for all resource management. No raw `new`/`delete`.
- No exceptions in hot paths (capture loop, encode loop).
- Every subsystem must log its initialization success/failure on startup.
- The agent must handle SIGTERM/SIGINT gracefully: destroy sandbox, flush logs, exit cleanly.

---

## `@devops`
**Scope:** CI/CD, Docker, GitHub Actions, deployment scripts.
**Files owned:** `.github/`, `docker-compose.yml`, `scripts/`, `Dockerfile`s
**Trigger:** CI failures, deployment config, environment setup issues.
**Rules:**
- Every GitHub Actions workflow must have a job-level timeout.
- Secrets must never be hardcoded — use GitHub Secrets or environment variables.
- Docker images must use multi-stage builds to minimize image size.
- All `scripts/` files must be idempotent (safe to run multiple times).

---

## `@security`
**Scope:** Threat modelling, sandbox escape analysis, input validation, authentication.
**Files owned:** `research/04-sandboxing/security-model.md`, reviews across all packages.
**Trigger:** Before any feature that crosses a trust boundary:
- host ↔ sandbox
- user ↔ backend (session token validation)
- backend ↔ Stripe (webhook verification)
- agent ↔ IPC
**Rules:**
- Must be invoked before any session lifecycle code ships to production.
- Must maintain `research/04-sandboxing/security-model.md` as the living threat model.
- For every new trust boundary: document the threat, the attack vector, the mitigation, and the residual risk.
- The five security invariants in `CLAUDE.md` are non-negotiable.
