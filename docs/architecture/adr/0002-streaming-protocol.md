# ADR-0002: WebRTC for Phase 1 Streaming, QUIC/MoQ for Phase 2

**Status:** Accepted  
**Date:** 2025-05-14  
**Author:** @architect

## Context

IndraNet's core value proposition depends on low-latency interactive streaming. Users must see the host screen with low enough latency to interact meaningfully with games and applications. Target latency budgets:

- Gaming: <30ms end-to-end (glass-to-glass)
- Video editing: <80ms acceptable
- AI workloads (non-real-time): <200ms acceptable

Phase 0 PoC 02 must prove the streaming transport can hit <50ms on LAN. The Phase 1 MVP must work over the internet with reasonable NAT traversal.

Three protocol categories were evaluated:
1. WebRTC (mature, browser-native, built-in NAT traversal)
2. QUIC-based custom protocol (MoQ, raw QUIC) (low latency, no browser NAT help)
3. HLS/DASH (HTTP streaming) (high latency, unacceptable for interactive use)

## Decision

**Phase 1:** Use WebRTC as the streaming transport.

WebRTC is the right choice for MVP because:
- Browser-native: the web frontend can receive streams without plugins
- Built-in ICE/STUN/TURN: NAT traversal is handled by the protocol
- GStreamer has a mature `webrtcbin` element that can consume NVENC-encoded H.264
- libwebrtc (Chromium's implementation) is available for the C++ agent
- Existing signaling patterns (offer/answer/ICE over WebSocket) are well-understood
- The PoC (poc/02) can be built and measured quickly

**Phase 2:** Evaluate QUIC/MoQ (Media over QUIC) to replace WebRTC.

QUIC/MoQ rationale for Phase 2:
- Lower queuing head-of-line blocking than WebRTC's SRTP over UDP
- The IETF MoQ working group is standardizing exactly this use case (low-latency live media)
- More flexible congestion control tuning
- WebRTC's bundled congestion control (GCC) may not be optimal for our workload

## Consequences

**Positive (WebRTC Phase 1):**
- Can ship PoC 02 quickly using GStreamer WebRTC
- Works in browsers without custom plugins
- STUN/TURN infrastructure is off-the-shelf (coturn, Cloudflare TURN)
- Large community and documentation

**Negative (WebRTC Phase 1):**
- WebRTC's peer connection setup adds ~500ms-2s of ICE negotiation on first connect
- The browser WebRTC stack imposes some latency floor we cannot control
- WebRTC's congestion control (Google Congestion Control) may not be optimal for gaming
- Migrating from WebRTC to QUIC/MoQ in Phase 2 will require rewriting the streaming path

**Negative (QUIC/MoQ Phase 2 risk):**
- MoQ standard is still being finalized; implementation maturity is low as of 2025
- Will require custom browser extension or Tauri-only viewer for Phase 2

## Alternatives Considered

**Custom UDP protocol:** Maximum control over latency, but building NAT traversal from scratch is a multi-month detour. Rejected for Phase 1.

**RTMP/SRT:** Latency is 1-3 seconds (intra-frame buffering). Unacceptable for interactive use.

**Parsec SDK:** Parsec offers a licensed streaming SDK but creates vendor lock-in and undisclosed cost structure at scale. Rejected on strategic grounds.

**Moonlight/Sunshine protocol (NVIDIA GameStream):** Open source but tied to NVIDIA hardware on the host side. Our multi-GPU support requirement (NVIDIA + AMD + Intel) rules this out.
