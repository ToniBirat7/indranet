# IndraNet

**The compute network that never sleeps. Indra's Net for idle GPUs.**

IndraNet is a peer-to-peer interactive compute marketplace. People with idle gaming PCs rent live, interactive desktop sessions to users who need GPU compute — for gaming, video editing, AI workloads, and more.

Unlike headless GPU marketplaces (Vast.ai, RunPod), IndraNet delivers **real interactive GUI sessions**: live screen streaming, keyboard and mouse input, gamepad forwarding, and audio — all sandboxed so the host machine is protected.

Named after Indra's Net from Hindu/Buddhist cosmology: an infinite cosmic web where a jewel at every node reflects every other jewel. Each host machine is a node; each user session illuminates the whole network.

---

## Architecture

```
  USER CLIENT (Browser / Tauri App)
       │  WebRTC (P2P, direct)
       │  Keyboard / Mouse / Gamepad input
       ▼
  ┌─────────────────────────────────────────────────────┐
  │               HOST MACHINE                           │
  │                                                      │
  │  ┌──────────────────────────────────────────────┐   │
  │  │          Windows Sandbox (ephemeral)          │   │
  │  │   Game / App running here                     │   │
  │  │   GPU-accelerated, isolated filesystem        │   │
  │  └──────────────────────────────────────────────┘   │
  │          ↑ screen pixels           ↓ HID events      │
  │  ┌──────────────────────────────────────────────┐   │
  │  │          C++ Host Agent Daemon                │   │
  │  │  DXGI Capture → NVENC Encode → WebRTC Send   │   │
  │  │  WebRTC Recv → ViGEm Inject → Sandbox Input  │   │
  │  └──────────────────────────────────────────────┘   │
  └─────────────────────────────────────────────────────┘
       │  WebSocket signaling only
       │  Billing ticks
       ▼
  ┌────────────────────────────────┐
  │   Go Backend (Stateless API)   │
  │   PostgreSQL + Redis + NATS    │
  │   Stripe Connect billing       │
  └────────────────────────────────┘
       │
       ▼
  ┌────────────────────────────────┐
  │   Next.js Marketplace          │
  │   Browse hosts, book sessions  │
  │   Wallet, earnings dashboard   │
  └────────────────────────────────┘
```

Video and audio stream **peer-to-peer** — it never touches our servers. The backend only handles signaling, billing, and authentication.

---

## Current Status

**Phase 0 — Research & Validation**

| PoC | Goal | Status |
|-----|------|--------|
| PoC 01: Screen Capture + Encode | <5ms capture+encode latency at 1080p60 via NVENC | NOT STARTED |
| PoC 02: WebRTC Stream | <50ms end-to-end latency on LAN | NOT STARTED |
| PoC 03: Windows Sandbox | GPU-accelerated app running inside Windows Sandbox | NOT STARTED |
| PoC 04: Payment Gate | Stripe payment → JWT token → WebRTC stream access | NOT STARTED |

When all 4 PoCs pass, Phase 1 MVP development begins.

---

## Getting Started (Local Dev)

### Prerequisites
- Docker + docker-compose
- Go 1.22+
- Node.js 20+ + pnpm 9+
- Rust (for Tauri client builds)

### Setup

```bash
# Mac / Linux
./scripts/setup.sh

# Windows
.\scripts\setup.ps1
```

### Start local services (Postgres, Redis, NATS, MinIO)
```bash
docker-compose up -d
```

### Start backend
```bash
cd packages/backend
make dev
```

### Start web marketplace
```bash
cd packages/web
pnpm dev
```

---

## Repository Structure

```
packages/backend/   Go REST API + WebSocket signaling
packages/web/       Next.js 14 marketplace frontend
packages/client/    Tauri v2 desktop app
packages/agent/     C++ host daemon
packages/shared/    TypeScript types + Zod schemas
docs/               Architecture docs, ADRs, specs
research/           Phase 0 research notes by topic
poc/                Proof of concept experiments
scripts/            Dev setup and utility scripts
```

---

## Agent Guide

This project uses specialized Claude Code sub-agents with defined ownership. See [AGENTS.md](AGENTS.md) for the full list. When contributing, name which agent role you are working as.

---

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the full phase breakdown.

| Phase | Goal | Timeline |
|-------|------|----------|
| Phase 0 | Prove the 4 core PoCs | Weeks 1-6 |
| Phase 1 | MVP: one host, one user, one paid session | Weeks 7-16 |
| Phase 2 | Multi-host marketplace, Hyper-V GPU passthrough | Months 5-9 |
| Phase 3 | Mobile clients, Linux hosts, global CDN relay | Months 10-15 |

---

## License

MIT
