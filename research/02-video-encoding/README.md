# Video Encoding Research

## Questions to Answer

1. What NVENC API settings achieve minimum encode latency at 1080p60?
2. How do we configure CBR low-latency mode in NVENC?
3. What is the NVENC latency penalty for enabling B-frames? (Spoiler: we disable them)
4. How does AMD AMF compare to NVENC for latency? (For non-NVIDIA hosts)
5. What FFmpeg filter graph do we need to integrate NVENC into GStreamer?

## Research Status

| Document | Status |
|----------|--------|
| nvenc-notes.md | NOT STARTED |
| amd-amf-notes.md | NOT STARTED |
| codec-comparison.md | NOT STARTED |
| ffmpeg-pipeline.md | NOT STARTED |
