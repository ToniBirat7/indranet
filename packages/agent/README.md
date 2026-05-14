# Host Agent — C++ Daemon

The IndraNet host agent is a C++ daemon that runs on the host machine and manages the complete session lifecycle from the host side.

## Responsibilities
- Screen capture via DXGI Desktop Duplication API
- Hardware video encoding via NVENC/AMF/VAAPI
- WebRTC streaming via GStreamer webrtcbin
- Input injection via ViGEm Bus (keyboard/mouse/gamepad)
- Windows Sandbox lifecycle management
- Billing heartbeat to backend
- IPC with the Tauri desktop client

## Build

### Windows (MSVC)
```powershell
# Install vcpkg dependencies
vcpkg install --triplet x64-windows

# Configure and build
cmake -B build -DCMAKE_BUILD_TYPE=Release -DCMAKE_TOOLCHAIN_FILE=<vcpkg>/scripts/buildsystems/vcpkg.cmake
cmake --build build --config Release
```

### Linux (experimental, Phase 3)
```bash
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -- -j$(nproc)
```

## Architecture
See `../../docs/specs/host-agent-spec.md` for the full specification.

## Dependencies
- GStreamer 1.20+ (WebRTC, video pipeline)
- NVENC SDK (NVIDIA hardware encoding)
- AMD AMF SDK (AMD hardware encoding)
- ViGEm Client SDK (virtual gamepad input)
- libwebsockets or similar (signaling WebSocket)
- nlohmann/json (JSON serialization)
