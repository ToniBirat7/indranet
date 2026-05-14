# DXGI Desktop Duplication API Notes

## Question
How do we capture the screen on Windows with minimal latency using DXGI?

## Method
TODO: Read MSDN documentation. Test on a Windows machine with an NVIDIA GPU.

## Key API Flow

```cpp
// 1. Get DXGI output (the monitor we want to capture)
IDXGIAdapter1* adapter;
IDXGIOutput* output;
IDXGIOutput1* output1;
output->QueryInterface(__uuidof(IDXGIOutput1), (void**)&output1);

// 2. Create desktop duplication
IDXGIOutputDuplication* duplication;
output1->DuplicateOutput(d3d11_device, &duplication);

// 3. Capture loop
DXGI_OUTDUPL_FRAME_INFO frame_info;
IDXGIResource* desktop_resource;

duplication->AcquireNextFrame(
    timeout_ms,        // 16ms for 60fps target
    &frame_info,
    &desktop_resource
);

// 4. Get D3D11 texture from resource
ID3D11Texture2D* texture;
desktop_resource->QueryInterface(__uuidof(ID3D11Texture2D), (void**)&texture);

// 5. Pass texture to NVENC (zero-copy if same GPU)
// ...

// 6. Release frame (MUST be called before next AcquireNextFrame)
duplication->ReleaseFrame();
```

## Known Quirks

TODO: Document after hands-on testing. Known issues from docs:
- `DXGI_ERROR_ACCESS_LOST`: Returned when desktop is locked, DWM restarts, or GPU driver update. Must call `DuplicateOutput` again.
- `DXGI_ERROR_WAIT_TIMEOUT`: No new frame in the timeout window. Normal at <60fps source.
- Protected content: DRM-protected content (Netflix, etc.) appears as black frame.
- Multi-adapter: If the game runs on a secondary GPU (laptop discrete GPU), capture must happen on the same adapter.

## Findings
TODO

## Recommendation
TODO

## Open Questions
- What is the actual p99 latency of `AcquireNextFrame` at 60fps?
- Does running inside Windows Sandbox change capture behavior?
- How do we handle the case where the user's game is on a secondary monitor?
