# WebRTC Implementation Notes

## Question
Should we use GStreamer WebRTC (`webrtcbin`) or libwebrtc (Chromium's implementation) for the host agent?

## Options

### Option A: GStreamer + webrtcbin
- GStreamer is a mature C library with good Windows support
- `webrtcbin` wraps libwebrtc internally (GStreamer 1.20+)
- Pipeline-based API fits well with our capture → encode → send flow
- Available via vcpkg

### Option B: libwebrtc directly
- Maximum control over WebRTC internals
- Complex build (Chromium's build system)
- No pipeline abstraction — more boilerplate

## Recommendation
Start with GStreamer + `webrtcbin` for PoC 02. If latency is unacceptable, consider libwebrtc directly.

## GStreamer Signaling Integration

The Go signaling server sends SDP offers/answers and ICE candidates over WebSocket. The C++ agent must:
1. Create a GStreamer pipeline with `webrtcbin`
2. Connect `on-negotiation-needed` signal → generate offer, send via WebSocket
3. Apply remote SDP answer received from WebSocket
4. Exchange ICE candidates via WebSocket

## Findings
TODO: Fill in after PoC 02.

## Open Questions
- What is GStreamer's overhead vs raw libwebrtc for RTP packetization?
- Does `webrtcbin` support custom congestion control settings?
- What is the minimum achievable jitter buffer size via GStreamer's WebRTC API?
