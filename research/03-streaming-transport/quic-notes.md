# QUIC / Media over QUIC (MoQ) Notes

## Relevance
Phase 2 streaming transport research. Not needed for Phase 1.

## Why QUIC for Phase 2?

WebRTC's SRTP over UDP has queuing HOL (head-of-line) blocking at the RTP level. QUIC provides per-stream multiplexing without HOL blocking, which can reduce tail latency.

The IETF MoQ (Media over QUIC) working group is standardizing a low-latency live media transport over QUIC. Draft spec: `draft-ietf-moq-transport`.

## Current Status (2025)
- MoQ spec is in draft; no production-ready implementations
- Facebook's `mvfst` and Google's `quiche` are QUIC libraries
- No browser-native support for sending raw QUIC streams (only via WebTransport API)

## Phase 2 Plan
Re-evaluate MoQ when:
1. A stable implementation exists (likely 2026)
2. WebTransport has wider browser support
3. WebRTC latency proves to be a bottleneck in production

## TODO
Monitor IETF MoQ working group progress. Re-evaluate in Q1 2026.
