# PoC 03: Windows Sandbox GPU Validation

## Goal
Launch a Windows Sandbox with GPU virtualization enabled and verify that hardware-accelerated 3D rendering works inside the sandbox.

## Success Criteria
- [ ] Windows Sandbox launches successfully with `<VGpu>Enable</VGpu>`
- [ ] GPU device visible inside sandbox (`dxdiag` shows GPU model)
- [ ] DirectX feature level available (target: 12_0 or higher)
- [ ] 3D application runs with hardware acceleration (not WARP software renderer)
- [ ] GPU benchmark score inside sandbox > 60% of bare metal score

## Prerequisites
- Windows 10/11 Pro or Enterprise
- Virtualization enabled in BIOS
- Windows Sandbox feature enabled (see `docs/onboarding/host-setup.md`)
- GPU with DirectX 12 support

## How to Run

```powershell
# Enable Windows Sandbox if not already enabled
Enable-WindowsOptionalFeature -Online -FeatureName "Containers-DisposableClientVM" -All

# Launch the test sandbox
.\launch-sandbox.ps1
```

The sandbox will open. Inside the sandbox:
1. Open Command Prompt
2. Run `dxdiag` — check that your GPU is listed under Display tab
3. Download and run a DirectX benchmark tool
4. Note the score

## Results

TODO: Fill in after running.

| Test | Result |
|------|--------|
| Sandbox launched | TODO |
| GPU visible in dxdiag | TODO |
| DirectX feature level | TODO |
| GPU benchmark (bare metal) | TODO |
| GPU benchmark (sandbox) | TODO |
| Performance ratio | TODO % |

## Notes
TODO: Record any issues with GPU access, DirectX version, etc.
