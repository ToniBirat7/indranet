# FFmpeg Pipeline for NVENC + GStreamer Integration

## GStreamer Pipeline (PoC approach)

```
appsrc → videoconvert → nvh264enc → h264parse → webrtcbin
```

GStreamer NVENC element configuration:
```
nvh264enc bitrate=15000 rc-mode=cbr-ld-hq gop-size=120 b-frames=0
```

## Direct NVENC (Production approach)

For production, bypass GStreamer for the capture→encode path. Use NVENC SDK directly:
1. DXGI → ID3D11Texture2D → NVENC input buffer (zero-copy)
2. NVENC output bitstream → GStreamer `appsrc` for WebRTC packetization

This eliminates the GStreamer video pipeline overhead for the hot path, using GStreamer only for RTP packetization and WebRTC transport.

## TODO
Document the exact GStreamer + WebRTC pipeline after PoC 02 testing.
