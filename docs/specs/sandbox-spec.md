# Sandbox Security Specification

## Requirements

1. **Isolation:** Guest process cannot read host filesystem outside mapped paths.
2. **Ephemeral:** All guest state destroyed on session end — no persistence.
3. **GPU access:** DirectX rendering must work for gaming use cases.
4. **Input isolation:** Injected input affects only the sandbox application, not host desktop.
5. **Network policy:** Configurable per session (allow internet / deny / LAN-deny).

## Phase 1: Windows Sandbox

### Configuration Template

```xml
<Configuration>
  <VGpu>Enable</VGpu>
  <Networking>Enable</Networking>
  <MappedFolders>
    <!-- Read-only game assets if pre-staged -->
    <MappedFolder>
      <HostFolder>C:\IndraNet\Sessions\{session_id}\assets</HostFolder>
      <SandboxFolder>C:\Users\WDAGUtilityAccount\Desktop\assets</SandboxFolder>
      <ReadOnly>true</ReadOnly>
    </MappedFolder>
  </MappedFolders>
  <LogonCommand>
    <Command>C:\Users\WDAGUtilityAccount\Desktop\assets\start.bat</Command>
  </LogonCommand>
</Configuration>
```

### GPU Access Policy

`<VGpu>Enable</VGpu>` enables paravirtualized GPU access (WDDM virtual adapter). The sandbox shares the host GPU via Windows' Hyper-V GPU virtualization layer. DirectX 12 feature level is available; performance is typically 70-90% of bare metal for rendering workloads.

### Network Policy

Phase 1: Enable networking (users need internet access in some workloads).
Phase 2: Implement per-session network policy options (game-only, internet, air-gapped).

### Limitations

- No GPU compute (CUDA/ROCm) in Phase 1 — only graphics
- Sandbox cannot span multiple monitors
- `VGpu` requires the host to have a DirectX 12 capable GPU

## Sandbox Destruction Protocol

On session end, the agent must:

1. Send WM_CLOSE to the application window inside sandbox
2. Wait 5 seconds for graceful exit
3. Call `TerminateProcess()` on the sandbox host process if not exited
4. Verify the sandbox process tree is fully terminated (no child processes)
5. Delete the `.wsb` configuration file
6. Delete the session asset folder `C:\IndraNet\Sessions\{session_id}\`
7. Log: `sandbox_destroyed session_id={id} cleanup_time_ms={t}`

Step 4 (verify clean termination) must not be skipped. A leaked sandbox process can persist indefinitely and is a security violation.

## Phase 2: Hyper-V DDA

See `research/04-sandboxing/hyperv-dda.md` for research notes.
See ADR-0003 for the decision record.
