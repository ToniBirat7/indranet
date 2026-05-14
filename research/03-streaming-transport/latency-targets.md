# Streaming Latency Targets and Budget

## The Latency Problem

Interactive use requires that input-to-display latency be low enough that the user feels "present" in the remote session. Research on human perception of latency:

- **<50ms:** Imperceptible latency — feels local
- **50-100ms:** Slightly perceptible, still comfortable for most uses
- **100-150ms:** Noticeable, acceptable for casual gaming and productivity
- **150-250ms:** Distinctly laggy; acceptable for turn-based, unacceptable for FPS
- **>250ms:** Unacceptably laggy for any interactive use

## Latency Budget: 100ms Internet Target

We target <100ms glass-to-glass latency over home broadband connections within the same continent.

| Stage | Min | Target | Max | Notes |
|-------|-----|--------|-----|-------|
| DXGI frame capture | 0.5ms | 1ms | 2ms | Hardware DMA readback |
| NVENC encode | 1ms | 2.5ms | 5ms | Low-latency CBR profile |
| Encode→send queue | 0.1ms | 0.5ms | 2ms | Ring buffer, minimal blocking |
| WebRTC send buffer | 0.5ms | 2ms | 5ms | WebRTC jitter buffer (send side) |
| Network transit | 10ms | 30ms | 80ms | Same-continent assumption |
| Packet loss handling | 0ms | 1ms | 10ms | NACK or FEC (no PLI preferred) |
| WebRTC receive jitter | 2ms | 8ms | 20ms | Adaptive jitter buffer |
| H.264 hardware decode | 1ms | 3ms | 8ms | GPU decode (DXVA2/D3D11VA) |
| Display pipeline | 2ms | 6ms | 12ms | Monitor refresh at 60Hz |
| **Total** | **17ms** | **54ms** | **144ms** | |

**LAN target (PoC):** Network transit ≈ 0.5ms → total ≈ 24ms achievable.

## Measurement Methodology

### End-to-End Latency (PoC 02)

Embed a timestamp watermark in the captured frame:
1. Host agent stamps current monotonic timestamp in a corner pixel pattern (encode the timestamp as a pixel value in a 8x8 block)
2. Receiver reads the timestamp from decoded frames
3. `latency = receive_time - encoded_timestamp`

This measures glass-to-glass excluding display pipeline.

### Capture + Encode Latency (PoC 01)

1. Record `QueryPerformanceCounter()` before `AcquireNextFrame()`
2. Record after `NvEncEncodePicture()` returns with a bitstream
3. `capture_encode_latency = t2 - t1`

Measure over 1000 frames and report p50, p95, p99.

## Bottleneck Analysis

### If p95 encode latency > 5ms

Check:
- Is `NV_ENC_PARAMS_RC_CBR_LOWDELAY_HQ` enabled?
- Is B-frame count = 0?
- Is NVENC running on the same GPU as the DXGI capture? (Cross-GPU copy adds latency)
- Is the system under thermal throttling?

### If network transit > 50ms

Check:
- Geographic distance to TURN server
- Are STUN/TURN candidates being used instead of direct P2P?
- Is there an intermediate relay adding hops?

### If jitter buffer > 15ms

WebRTC's adaptive jitter buffer grows when packet arrival is irregular. Causes:
- QoS/packet shaping by the host's router
- Cellular network (high jitter by nature)
- Wi-Fi with interference

Mitigation: FEC (Forward Error Correction) can reduce jitter buffer needs at the cost of bandwidth.

## Open Questions

- What is the actual NVENC encode latency distribution at 4K60 vs 1080p60?
- Does GStreamer's `webrtcbin` introduce additional buffering vs raw libwebrtc?
- What is the minimum achievable jitter buffer setting in Chrome's WebRTC stack?
- Does Windows Sandbox add capture latency (paravirtualized display DMA path)?
