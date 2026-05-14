# IndraNet — Claude Code Context

## What is IndraNet?
A peer-to-peer interactive compute marketplace. People with idle gaming PCs (hosts) rent
live, interactive desktop sessions to users who need GPU compute. The key innovation:
unlike headless GPU marketplaces (Vast.ai, RunPod), IndraNet delivers **real interactive
GUI sessions** — screen streaming, keyboard/mouse/gamepad input, audio — all sandboxed.

Named after Indra's Net: an infinite cosmic web where every node reflects every other.

## Project Phase
Currently in: **Phase 0 (Research & Validation) + Phase 1 MVP setup**

MVP goal: One host machine streams a game over WebRTC to a user client,
gated behind a Stripe payment, with the game running inside Windows Sandbox.
If this works, the core hypothesis is proven.

## Critical Architecture Facts
- The **HOST AGENT** is a C++ daemon. It runs on the host's machine. It owns:
  screen capture (DXGI), hardware encoding (NVENC/AMF), WebRTC streaming,
  input injection (ViGEm), and sandbox lifecycle management.
- The **BACKEND** is Go. It is stateless and only orchestrates: host registration,
  session creation, billing ticks, payment webhooks, WebRTC signaling.
- The **WEB/CLIENT** are TypeScript. The web is Next.js (marketplace browsing, session viewer).
  The desktop client (Tauri) wraps the host agent for host-side UX.
- Sessions are **PEER-TO-PEER** after handshake. Video/audio/input never goes through our servers.
- Sandboxes are **EPHEMERAL**. All disk state is destroyed when session ends.

## Tech Stack
```
Host Agent      → C++ (MSVC/Clang, Windows-primary)
Backend         → Go (Golang) — REST API + WebSocket signaling
Web Marketplace → Next.js 14 (App Router) + TypeScript + Tailwind CSS
Desktop Client  → Tauri v2 (Rust shell + React/TypeScript webview)
Database        → PostgreSQL (via pgx driver in Go)
Cache           → Redis
Message Queue   → NATS
File Storage    → Cloudflare R2 (S3-compatible)
Payment         → Stripe (Connect for marketplace payouts)
Streaming       → WebRTC (libwebrtc / GStreamer for PoC), QUIC later
Sandboxing      → Windows Sandbox API (Phase 1), Hyper-V DDA (Phase 2)
Video Encoding  → NVENC SDK (NVIDIA), AMF (AMD), VAAPI (Intel) via FFmpeg
Virtual Input   → ViGEm Bus (Windows), uinput (Linux)
Monorepo        → Turborepo + pnpm workspaces (JS/TS packages)
Container       → Docker + docker-compose (local dev)
CI/CD           → GitHub Actions
```

## Monorepo Structure
```
packages/backend/   → Go REST API + WebSocket signaling server
packages/web/       → Next.js 14 marketplace frontend
packages/client/    → Tauri v2 desktop app (host UX + user session viewer)
packages/agent/     → C++ host daemon (capture, encode, stream, input, sandbox)
packages/shared/    → Shared TypeScript types + Zod schemas
```

See `AGENTS.md` for which agent owns which directory.
pnpm workspaces for JS/TS packages. Go modules for backend. CMake for C++ agent.

## Phase 0 PoC Success Criteria
- [ ] poc/01: Screen captured at 60fps with <5ms encode latency (NVENC)
- [ ] poc/02: WebRTC stream delivered with <50ms end-to-end latency on LAN
- [ ] poc/03: Windows Sandbox launched with GPU-accelerated app running
- [ ] poc/04: $1 Stripe payment gates access to a WebRTC stream token

## Key Technical Decisions (see docs/architecture/adr/)
- ADR-0001: Monorepo with Turborepo
- ADR-0002: WebRTC for Phase 1 streaming (QUIC/MoQ in Phase 2)
- ADR-0003: Windows Sandbox (app-level) for MVP, Hyper-V DDA in Phase 2
- ADR-0004: GPU passthrough deferred to Phase 2; shared GPU in MVP
- ADR-0005: Stripe Connect for marketplace payments

## Do Not Touch Without Reading
- Any session billing logic → read `docs/specs/api-spec.md` first
- Any sandbox code → read `research/04-sandboxing/security-model.md` first
- Any streaming code → read `research/03-streaming-transport/latency-targets.md` first
- Any payment code → read `research/06-payment/billing-model.md` first

## Running Locally
See `scripts/setup.sh` (Mac/Linux) or `scripts/setup.ps1` (Windows).
`docker-compose up` starts postgres, redis, nats, and a local S3 (minio).
`cd packages/backend && make dev` starts the Go backend with hot reload.
`cd packages/web && pnpm dev` starts Next.js dev server.

## Agent Invocation
See `AGENTS.md`. When starting a new feature, state which agent you are acting as.
Each agent has defined file ownership and rules. Respect agent boundaries.

## Session State Machine
```
CREATED → AUTHORIZED → ACTIVE → ENDING → ENDED
                              ↓
                           FAILED
```
- CREATED: Session record exists, payment pending
- AUTHORIZED: Payment confirmed, host notified
- ACTIVE: Host agent confirmed stream is live
- ENDING: Kill signal sent (balance exhausted or user disconnect)
- ENDED: Session terminated, final billing calculated
- FAILED: Any error state (payment failure, host crash, timeout)

## Billing Model
- Per-minute billing. Users pre-fund a wallet (Stripe).
- Backend billing engine ticks every 60 seconds per active session.
- At 5 minutes remaining balance: emit WARNING event to client.
- At 0 balance: emit KILL event → host agent terminates sandbox.
- Platform takes 20% commission; host receives 80% (Stripe Connect payout).

## Security Invariants (never break these)
1. The sandbox MUST be destroyed when a session ends — no exceptions.
2. Host filesystem MUST NOT be accessible from inside the sandbox.
3. Input events from users MUST be injected into the sandbox only, not the host desktop.
4. The backend MUST validate every session token before accepting WebSocket connections.
5. Stripe webhooks MUST be verified with the signing secret before processing.
