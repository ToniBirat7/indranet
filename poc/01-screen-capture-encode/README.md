# PoC 01: Screen Capture + Hardware Encode

## Goal
Capture the screen using Windows DXGI Desktop Duplication API and encode each frame using NVIDIA NVENC, achieving <5ms combined capture+encode latency at 1080p60.

## Success Criteria
- [ ] DXGI capture initializes successfully
- [ ] NVENC encoder initializes in low-latency CBR mode
- [ ] Sustained 60fps capture+encode without dropped frames
- [ ] p95 combined latency < 5ms over 1000 frames
- [ ] p99 combined latency < 8ms

## How to Build

### Prerequisites
- Windows 10/11 with an NVIDIA GPU (RTX series recommended)
- NVIDIA NVENC SDK (ships with CUDA Toolkit or standalone download)
- CMake 3.20+
- Visual Studio 2022 or LLVM/Clang

```powershell
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
```

## How to Run
```powershell
.\build\Release\poc01_capture.exe
```
Output: prints per-frame timing stats to stdout. Runs for 1000 frames then exits.

## Results

TODO: Fill in after running.

| Metric | Value |
|--------|-------|
| Hardware | TODO |
| NVENC Preset | P1 / CBR_LOWDELAY_HQ |
| Resolution | 1080p |
| Target FPS | 60 |
| p50 capture latency | TODO |
| p95 capture latency | TODO |
| p50 encode latency | TODO |
| p95 encode latency | TODO |
| p95 combined | TODO |
| Sustained FPS | TODO |

## Notes
TODO: Record any gotchas, DXGI_ERROR_ACCESS_LOST occurrences, etc.
