# Screen Capture Research

## Questions to Answer

1. What is the minimum achievable DXGI Desktop Duplication capture latency on Windows?
2. Does capture latency change when the source is inside Windows Sandbox (paravirtualized WDDM)?
3. What happens on multi-monitor setups? Can we specify which output to capture?
4. How does DXGI handle HDR / 10-bit color? Do we need to tonemap?
5. What is the behavior when the captured display is at 144Hz vs 60Hz?

## Research Status

| Document | Status |
|----------|--------|
| dxgi-notes.md | NOT STARTED |
| pipewire-notes.md | NOT STARTED |
| benchmarks.md | NOT STARTED |

## Key Resources

- [MSDN: Desktop Duplication API](https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api)
- [MSDN: IDXGIOutputDuplication](https://learn.microsoft.com/en-us/windows/win32/api/dxgi1_2/nn-dxgi1_2-idxgioutputduplication)
