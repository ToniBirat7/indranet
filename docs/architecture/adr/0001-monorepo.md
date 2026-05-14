# ADR-0001: Monorepo with Turborepo

**Status:** Accepted  
**Date:** 2025-05-14  
**Author:** @architect

## Context

IndraNet consists of five distinct software components that must be developed in parallel:
- Go backend (REST API + signaling)
- Next.js web frontend (marketplace)
- Tauri desktop client (host + user app)
- C++ host agent (capture/encode/stream)
- Shared TypeScript types and schemas

These components share type definitions, API contracts, and configuration. Independent repositories would create friction: keeping shared types in sync across repos requires versioned packages and publish/consume cycles that slow iteration during Phase 0 and Phase 1.

## Decision

Use a monorepo structure with Turborepo managing the JavaScript/TypeScript pipeline and pnpm workspaces for package linking. The C++ agent and Go backend are not JS packages, so they live in `packages/agent/` and `packages/backend/` but are managed with their native tools (CMake and Go modules).

Turborepo provides:
- Parallel build/lint/test execution with task dependency graph
- Remote caching (CI speedups)
- `turbo run dev` to start all JS/TS packages simultaneously

## Consequences

**Positive:**
- Shared types in `packages/shared/` are immediately available to `packages/web/` and `packages/client/` without versioning
- Single `git clone` gets the entire project
- Refactoring API contracts is a single PR that updates all affected packages atomically
- Simpler CI configuration — one repo, one CI workflow

**Negative:**
- The C++ agent cannot fully participate in Turborepo's pipeline (no `package.json`)
- Large repo size as C++ build artifacts accumulate (mitigated by `.gitignore`)
- pnpm workspace hoisting can cause subtle dependency version conflicts

## Alternatives Considered

**Polyrepo:** Each component in its own GitHub repository. Rejected because during Phase 0 and 1, the pace of cross-cutting changes would make this painful. Can always split later.

**Nx:** More powerful than Turborepo but also significantly more complex to configure. Turborepo's zero-config approach is right for Phase 0.

**Lerna:** Legacy tool, superseded by Turborepo/Nx for monorepo orchestration.
