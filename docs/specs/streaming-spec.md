# IndraNet Streaming Specification

## Latency Targets by Use Case

| Use Case | Glass-to-Glass Target | Acceptable Max | Notes |
|----------|----------------------|----------------|-------|
| Gaming (FPS/action) | <30ms | <50ms | Anything above 50ms feels sluggish |
| Gaming (strategy/RPG) | <80ms | <120ms | Less input-sensitive |
| Video editing (playback) | <80ms | <150ms | Color grading needs accurate preview |
| 3D rendering preview | <200ms | <500ms | Latency matters less than quality |
| AI/ML inference UI | <200ms | 1000ms | Not interactive in the HID sense |

**Phase 1 MVP target:** <100ms glass-to-glass over the internet (home broadband, same continent).
**Phase 0 PoC target:** <50ms on LAN (proves the transport and encode pipeline).

---

## Latency Budget Breakdown

Total budget: 100ms (internet target)

| Stage | Budget | Notes |
|-------|--------|-------|
| DXGI frame capture | ~1ms | Hardware DMA, no CPU copy |
| NVENC encode (1080p60) | ~2-3ms | Low-latency CBR profile |
| WebRTC send buffer | ~2ms | Minimal buffering in low-latency mode |
| Network transit (same continent) | ~20-40ms | Varies with routing |
| WebRTC receive jitter buffer | ~5-10ms | Adaptive, target 1 frame |
| Browser/client decode (H.264 HW) | ~2-4ms | GPU decode |
| Display pipeline | ~5-8ms | Monitor refresh cycle |
| **Total** | **~37-68ms** | **Within 100ms budget** |

See `research/03-streaming-transport/latency-targets.md` for detailed measurement methodology.

---

## Codec Pipeline

### Phase 1: H.264 (Baseline)
- **Encoder:** NVENC (NVIDIA), AMF (AMD), VAAPI (Intel via FFmpeg)
- **Profile:** Baseline or High, no B-frames (B-frames add latency)
- **Level:** 5.1 (supports up to 4K30 or 1080p120)
- **Rate control:** CBR with low-latency tuning
  - NVENC: `NVENC_ENC_PARAMS_RC_CBR_LOWDELAY_HQ`
  - AMF: `AMF_VIDEO_ENCODER_USAGE_LOWLATENCY`
- **Keyframe interval:** Every 2 seconds max (allows seeking/reconnect recovery)
- **B-frames:** Disabled (zero B-frames) — they add 1-2 frames of latency

### Phase 2: H.265 (HEVC)
- Same encoder APIs with HEVC codec
- 30-40% bitrate reduction at same quality vs H.264
- Wider AMD/Intel hardware decode support by 2025+
- Not used in Phase 1 because of lingering decode compatibility edge cases

### Phase 3: AV1
- Best quality/bitrate ratio
- Hardware encoders (NVENC AV1 on RTX 40-series, Intel Arc)
- Target when hardware encode is mainstream (2026+)

---

## Bitrate Targets

| Resolution | FPS | H.264 Gaming | H.264 Desktop | H.265 Gaming |
|-----------|-----|-------------|---------------|-------------|
| 720p | 60 | 8 Mbps | 5 Mbps | 5 Mbps |
| 1080p | 60 | 15 Mbps | 8 Mbps | 9 Mbps |
| 1080p | 120 | 25 Mbps | N/A | 15 Mbps |
| 1440p | 60 | 25 Mbps | 15 Mbps | 15 Mbps |
| 4K | 60 | 50 Mbps | 30 Mbps | 30 Mbps |

**Phase 1 default:** 1080p60, 15 Mbps H.264. Host can set max; client adapts down.

---

## Adaptive Bitrate Strategy (Phase 1)

WebRTC's built-in congestion control (GCC — Google Congestion Control) handles adaptive bitrate automatically. The encoder must support dynamic bitrate change without re-initialization.

NVENC supports dynamic bitrate via `NvEncReconfigureEncoder()`. This must be called on the encoding thread only.

Minimum viable bitrate (stream remains watchable): 3 Mbps at 720p30.
If GCC drops below 2 Mbps, the session should emit a `QUALITY_WARNING` event to the user client.

---

## Audio

- **Codec:** Opus (WebRTC native)
- **Sample rate:** 48kHz
- **Channels:** Stereo
- **Bitrate:** 128 Kbps
- **Capture:** Windows WASAPI loopback (captures all system audio from inside sandbox)
- **Latency target:** <20ms audio-to-speaker

Audio sync with video is handled by WebRTC's built-in A/V sync mechanism.

---

## Input Forwarding

Input events must not add perceptible latency. The pipeline:

```
User keyboard/mouse event (browser/client)
  → Serialize as HID packet (JSON or binary)
  → Send via WebRTC data channel (unreliable, ordered: false)
  → Host agent receives on data channel
  → Inject via ViGEm / uinput into sandbox
  → Application receives input
```

**Target:** <5ms from user keypress to application receiving the event (host-side injection latency only).

Input event types:
- Keyboard: keydown, keyup (scancode + VK code)
- Mouse: mousemove (relative delta), mousedown, mouseup, scroll
- Gamepad: axis values, button states (Xbox controller HID report format)

Relative mouse mode: the user client must lock the pointer (`requestPointerLock()`) during active sessions to prevent mouse drift.
