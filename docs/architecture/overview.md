# IndraNet — System Architecture Overview

## System Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         USER SIDE                                    │
│                                                                      │
│  ┌──────────────────┐          ┌──────────────────────────────────┐ │
│  │  Web Browser     │          │  Tauri Desktop Client            │ │
│  │  (Next.js)       │          │  (React + Rust)                  │ │
│  │  Marketplace     │          │  Session viewer + host UX        │ │
│  └────────┬─────────┘          └──────────────┬───────────────────┘ │
│           │ HTTPS                             │ WebRTC (P2P)         │
└───────────┼───────────────────────────────────┼─────────────────────┘
            │                                   │
            ▼                                   │
┌─────────────────────────────────────────────────────────────────────┐
│                       INDRANET BACKEND                               │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Go REST API                                                  │   │
│  │  POST /hosts/register   POST /sessions   GET /hosts          │   │
│  │  PUT /sessions/:id/start   DELETE /sessions/:id              │   │
│  │  POST /webhooks/stripe                                        │   │
│  └──────────────────────┬───────────────────────────────────────┘   │
│                         │                                            │
│  ┌──────────────────┐   │   ┌──────────────┐   ┌────────────────┐  │
│  │  WebSocket Hub   │   │   │  PostgreSQL  │   │  Redis         │  │
│  │  (WebRTC signal) │   │   │  (sessions,  │   │  (session      │  │
│  │  offer/answer    │   │   │   users,     │   │   state cache) │  │
│  │  ICE relay       │   │   │   billing)   │   └────────────────┘  │
│  └──────────────────┘   │   └──────────────┘                        │
│                         │                                            │
│  ┌──────────────────┐   │                                            │
│  │  Billing Engine  │   │                                            │
│  │  60s tick loop   │   │                                            │
│  │  balance checks  │   │                                            │
│  └──────────────────┘   │                                            │
└─────────────────────────┼────────────────────────────────────────── ┘
                          │ WebSocket (signaling only)
                          │ Billing heartbeat
                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         HOST MACHINE                                 │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │             C++ Host Agent Daemon                             │   │
│  │                                                               │   │
│  │  IPC ←→ Tauri Client (named pipe / local socket)             │   │
│  │                                                               │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │   │
│  │  │ DXGI Capture│→ │ NVENC Encode│→ │ WebRTC Send         │  │   │
│  │  │ ~1ms        │  │ ~2ms        │  │ (GStreamer/libwebrtc)│  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  │   │
│  │                                              │ P2P video      │   │
│  │  ┌─────────────────────────────────────────┐│               │   │
│  │  │ WebRTC Receive ← Input events           ││               │   │
│  │  │ ViGEm inject → Sandbox HID              ││               │   │
│  │  └─────────────────────────────────────────┘│               │   │
│  │                                              │               │   │
│  │  ┌──────────────────────────────────────┐   │               │   │
│  │  │  Windows Sandbox (ephemeral)         │   │               │   │
│  │  │  Game / Application runs here        │   │               │   │
│  │  │  Isolated filesystem, shared GPU     │◄──┘               │   │
│  │  │  Destroyed on session end            │                   │   │
│  │  └──────────────────────────────────────┘                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flows

### Session Start Flow
1. User browses marketplace, selects host
2. User completes Stripe Checkout
3. Stripe webhook → backend marks session AUTHORIZED
4. Backend notifies host agent via WebSocket
5. Host agent creates Windows Sandbox
6. Host agent initiates WebRTC offer
7. Offer relayed via backend signaling hub to user client
8. ICE negotiation completes → P2P connection established
9. Host agent begins DXGI capture → NVENC encode → WebRTC send
10. Session marked ACTIVE in backend

### Billing Flow
1. Backend billing engine ticks every 60 seconds
2. For each ACTIVE session: deduct per-minute rate from user balance
3. At 5 min balance remaining: send WARNING to user client via WebSocket
4. At 0 balance: send KILL signal to host agent
5. Host agent receives KILL → terminates sandbox → cleans up
6. Session marked ENDED → Stripe payout to host (80%)

### Input Flow
1. User client captures keyboard/mouse events
2. Events serialized as HID packets
3. Sent to host agent via existing WebRTC data channel
4. Host agent injects events via ViGEm into sandbox only (not host desktop)

## Key Design Principles

**Peer-to-peer after handshake**: Video, audio, and input never touch IndraNet servers in steady state. The backend is only involved in signaling (WebRTC offer/answer/ICE) and billing.

**Ephemeral sandboxes**: The host's machine is protected by running all user-visible workloads inside Windows Sandbox. When the session ends, all state is wiped.

**Stateless backend**: The Go backend holds no in-memory session state that can't be reconstructed from Postgres + Redis. This enables horizontal scaling.

**Hardware encoding first**: NVENC/AMF/VAAPI are required for the latency budget. Software encoding cannot meet the <5ms encode target at 1080p60.
