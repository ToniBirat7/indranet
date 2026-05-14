# Windows Sandbox API Research

## Question
Does Windows Sandbox provide GPU access sufficient for 3D gaming, and what is the performance overhead?

## Key Facts from Documentation

- Windows Sandbox is available on Windows 10/11 Pro, Enterprise, and Education (NOT Home)
- Uses Hyper-V technology under the hood with a minimal base image
- The sandbox is stateless: every launch starts from a clean Windows image
- `<VGpu>Enable</VGpu>` in the `.wsb` config enables GPU paravirtualization
- GPU is shared via Microsoft's Hyper-V GPU Virtualization (WDDM over Hyper-V)
- DirectX 12 feature levels are available inside the sandbox

## Configuration

```xml
<Configuration>
  <VGpu>Enable</VGpu>
  <Networking>Enable</Networking>
</Configuration>
```

## Startup Time
Typical Windows Sandbox startup: 5-15 seconds. Phase 1 target: < 30 seconds from session AUTHORIZED to stream active.

## GPU Performance Expectations
From community reports and Microsoft documentation:
- Gaming: typically 70-90% of bare metal performance
- Encoding (NVENC inside sandbox): uncertain — needs PoC testing
- CUDA/compute: may not work (compute APIs may not be virtualized)

## TODO: Measure in PoC 03
- Launch sandbox on test machine
- Run `dxdiag` inside sandbox — check GPU model listed
- Run 3DMark Time Spy inside sandbox
- Compare score to bare metal

## Open Questions
- Does NVENC work inside Windows Sandbox (needed for capture-inside-sandbox design)?
- What is the actual FPS drop for a GPU-intensive game inside sandbox?
- Does Windows Sandbox support audio capture (WASAPI loopback) for audio streaming?
