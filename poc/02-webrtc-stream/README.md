# PoC 02: WebRTC End-to-End Stream

## Goal
Stream encoded video frames over WebRTC from a C++ sender to a browser viewer, achieving <50ms end-to-end latency on a LAN.

## Architecture

```
[C++ Streamer] --WebSocket signaling--> [Go Signaling Server] <--WebSocket-- [Browser Viewer]
[C++ Streamer] ============= WebRTC P2P (video) ============================> [Browser Viewer]
```

## Success Criteria
- [ ] Signaling server successfully relays offer/answer/ICE candidates
- [ ] WebRTC P2P connection established
- [ ] Video stream appears in browser
- [ ] End-to-end latency < 50ms on LAN (measured via timestamp watermark)
- [ ] Stream sustains 60fps without stuttering

## How to Build

### Signaling Server (Go)
```bash
cd server
go run main.go --port 8765
```

### C++ Streamer
Prerequisites: GStreamer 1.20+, GStreamer WebRTC plugin
```bash
# Windows (vcpkg)
vcpkg install gstreamer:x64-windows

# Build
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
.\build\Release\poc02_streamer.exe --signaling ws://localhost:8765 --session test123
```

### Browser Viewer
Open `viewer/index.html` in Chrome.
Enter session ID: `test123`
Click "Connect"

## Results

TODO: Fill in after running.

| Metric | Value |
|--------|-------|
| LAN end-to-end p50 | TODO |
| LAN end-to-end p95 | TODO |
| Sustained FPS | TODO |
| ICE connection time | TODO |
| Signaling library | GStreamer webrtcbin |

## Latency Measurement Method
Embed a timestamp in each frame as a pixel-encoded value in the top-left corner.
Viewer JavaScript decodes the timestamp from the video frame and computes latency.
