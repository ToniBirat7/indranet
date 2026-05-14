# IndraNet — Roadmap

## Phase 0: Research & Validation (Weeks 1–6)

**Goal:** Prove the four core technical hypotheses before writing production code.
**Success condition:** All four PoCs pass their success criteria.

### PoC 01 — Screen Capture + Hardware Encode
- DXGI Desktop Duplication API on Windows
- NVENC encoding at 1080p60
- **Success:** <5ms combined capture + encode latency, sustained 60fps

### PoC 02 — WebRTC Streaming
- GStreamer WebRTC or libwebrtc for encoded frame delivery
- Go signaling server for offer/answer/ICE exchange
- **Success:** <50ms end-to-end latency (capture → display) on LAN

### PoC 03 — Sandboxed GPU Session
- Windows Sandbox API: launch ephemeral container
- GPU device visible inside sandbox (DirectX capable)
- **Success:** 3D application runs inside sandbox with hardware acceleration

### PoC 04 — Payment-Gated Stream
- Stripe Checkout → payment success webhook → JWT session token
- JWT gates WebRTC stream access
- **Success:** User pays $1 → receives stream token → connects to stream

---

## Phase 1: MVP (Weeks 7–16)

**Goal:** One host machine rents a live, interactive, sandboxed GPU session to one paying user through the web marketplace.

### Milestone 1.1 — Backend Foundation (Weeks 7–9)
- [ ] Host registration and machine fingerprinting
- [ ] Session lifecycle API (CREATED → AUTHORIZED → ACTIVE → ENDED)
- [ ] Stripe Connect onboarding for hosts
- [ ] Per-minute billing engine with balance checks
- [ ] WebSocket signaling hub for WebRTC

### Milestone 1.2 — Host Agent v0.1 (Weeks 8–11)
- [ ] DXGI capture + NVENC encode pipeline (from PoC 01)
- [ ] WebRTC sender using GStreamer (from PoC 02)
- [ ] Windows Sandbox launch/teardown (from PoC 03)
- [ ] ViGEm virtual input injection
- [ ] Billing heartbeat to backend
- [ ] Tauri desktop app wrapping the agent

### Milestone 1.3 — Web Marketplace v0.1 (Weeks 9–12)
- [ ] Host listing page (browse available machines)
- [ ] Host detail + book session flow
- [ ] Stripe Checkout payment flow
- [ ] Active session page (WebRTC viewer + input capture)
- [ ] User dashboard (session history, wallet balance)
- [ ] Host dashboard (earnings, uptime, session history)

### Milestone 1.4 — Integration & Testing (Weeks 13–16)
- [ ] End-to-end test: user pays → session starts → stream works → session ends → host paid
- [ ] Billing accuracy test: session ends on balance exhaustion
- [ ] Sandbox destruction verified after every session
- [ ] Input forwarding tested with keyboard + mouse + gamepad
- [ ] Latency measurement: target <100ms end-to-end over internet

---

## Phase 2: Multi-Host Marketplace (Months 5–9)

**Goal:** Multiple hosts list their machines. Users can browse, filter, and book. GPU passthrough enabled via Hyper-V DDA.

### Key deliverables
- [ ] Host onboarding flow (self-serve signup, machine verification, pricing)
- [ ] Marketplace search and filtering (GPU model, VRAM, price, location)
- [ ] Hyper-V Discrete Device Assignment (DDA) for dedicated GPU passthrough
- [ ] Session quality ratings and host reputation system
- [ ] TURN server infrastructure for NAT traversal at scale
- [ ] Admin dashboard (fraud detection, dispute resolution)
- [ ] Linux host support (KVM + VFIO + PipeWire + VA-API)

---

## Phase 3: Scale & Polish (Months 10–15)

**Goal:** Consumer-ready product with mobile clients, global relay network, and developer API.

### Key deliverables
- [ ] Mobile viewer apps (iOS + Android — view-only sessions)
- [ ] QUIC/MoQ transport replacing WebRTC for lower latency
- [ ] Adaptive bitrate streaming (bandwidth-responsive quality)
- [ ] Global relay network (TURN at multiple regions)
- [ ] Developer API (programmatic session creation)
- [ ] Automated host verification (benchmark on registration)
- [ ] Fraud detection ML model (crypto mining abuse, TOS violations)
- [ ] Referral and host incentive program

---

## Non-Goals (explicitly out of scope)

- **Phase 1:** Multi-GPU sessions, Linux hosts, mobile viewers, global relay
- **Phase 2:** Mobile apps, QUIC transport, developer API
- **All phases:** Becoming a VDI provider, enterprise/corporate focus, non-gaming workloads as primary market
