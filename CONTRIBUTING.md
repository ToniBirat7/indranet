# Contributing to IndraNet

## Getting Started

1. Read `CLAUDE.md` — it is the single source of truth for project architecture and invariants.
2. Read `AGENTS.md` — identify which agent scope your work falls under.
3. Check `TODO.md` for immediate priorities.
4. Check `docs/architecture/adr/` for any relevant technology decisions.

## Branch Naming

```
feat/<agent>/<short-description>     # New feature
fix/<agent>/<short-description>      # Bug fix
research/<topic>                     # Research findings
poc/<poc-number>-<description>       # PoC work
chore/<description>                  # Tooling, deps, CI
```

Examples:
- `feat/backend/host-registration`
- `poc/01-screen-capture`
- `research/webrtc-latency`

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(backend): add host registration endpoint
fix(agent): handle DXGI device lost during capture
docs(research): add NVENC low-latency mode findings
poc(01): add DXGI capture skeleton with timing
```

## Pull Requests

- Keep PRs focused. One feature or fix per PR.
- PRs touching session billing or sandbox lifecycle require `@security` review.
- All Go code must pass `golangci-lint`.
- All TypeScript code must pass `pnpm lint` and `pnpm tsc`.
- Add tests for billing logic — billing bugs are money bugs.

## Code Standards

### Go
- `gofmt` + `goimports` on every file.
- Handle all errors. No `_ = err`.
- Use `pgx` for Postgres — no ORM.
- Use `context.Context` for cancellation in all I/O operations.

### TypeScript / React
- Server Components by default in Next.js App Router.
- Client Components only when interactivity is required (WebRTC, form state).
- Validate all external data with Zod schemas from `packages/shared`.

### C++ (Agent)
- RAII for all resources. No raw `new`/`delete`.
- No exceptions in the capture/encode hot path.
- Every latency-sensitive code path must have a `// PERF:` comment with the budget.
- Log initialization success/failure for every subsystem.

## Security

- Never commit `.env` files or API keys.
- Never hardcode credentials — use environment variables.
- If you touch a trust boundary (user↔backend, host↔sandbox, backend↔Stripe), tag `@security` for review.
- The five security invariants in `CLAUDE.md` are non-negotiable. Any PR that violates them will be rejected.

## Research Contributions

Research goes in `research/`. Format:

```markdown
## Question
What specific thing are we trying to learn?

## Method
How did we investigate this?

## Findings
What did we discover?

## Recommendation
What should we do based on this?

## Open Questions
What do we still not know?
```

## PoC Contributions

PoCs go in `poc/`. Every PoC README must include:
- Goal and success criteria
- How to run it
- Results (fill in after running)

PoC code quality bar is low — quick and dirty is fine. The goal is to answer a question, not to ship to production.
