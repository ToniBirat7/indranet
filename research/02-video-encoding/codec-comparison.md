# Codec Comparison: H.264 vs H.265 vs AV1

## For IndraNet's Use Case

| Factor | H.264 | H.265 | AV1 |
|--------|-------|-------|-----|
| Hardware encode support | Ubiquitous | Wide (2018+ GPUs) | RTX 40-series, Intel Arc |
| Hardware decode (client) | Universal | Wide | Growing (2022+ GPUs) |
| Bitrate efficiency | Baseline | ~30% better than H.264 | ~40% better than H.265 |
| Encode latency (HW) | <3ms | <4ms | ~5-8ms |
| Browser decode | Native (all browsers) | Native (most browsers) | Native (2023+ Chrome/Firefox) |
| Patent / licensing | MPEG-LA (royalty-free for free streaming) | HEVC Advance (complex) | Royalty-free |

## Recommendation by Phase

- **Phase 1:** H.264 only. Universal support, lowest latency, zero licensing risk.
- **Phase 2:** Add H.265 for hosts/clients that support it (reduces bandwidth 30%).
- **Phase 3:** Add AV1 when hardware encode is common (RTX 40+ mainstream).

## Latency Impact of Codec Choice

The latency difference between H.264 and H.265 hardware encode is small (~1ms). The bigger factor is configuration (B-frames, lookahead, profile). B-frames are the latency killer: each B-frame reference adds 1-2 frame delays (~16-33ms at 60fps).

**Rule:** No B-frames in any codec, any phase.
