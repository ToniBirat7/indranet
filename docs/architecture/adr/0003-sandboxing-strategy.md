# ADR-0003: Windows Sandbox API for MVP, Hyper-V DDA for Phase 2

**Status:** Accepted  
**Date:** 2025-05-14  
**Author:** @architect

## Context

IndraNet's security model requires that user workloads run in an isolated environment on the host machine. The host must be protected from:
- User code accessing the host's filesystem
- User code pivoting to attack the LAN
- Persistent malware surviving session end
- GPU workloads interfering with other sessions

This requirement exists at Phase 1 (single host, single session) and becomes more critical at Phase 2 (multiple simultaneous sessions on one machine).

The candidate isolation strategies are:
1. Windows Sandbox API (app container, no GPU passthrough)
2. Hyper-V with Discrete Device Assignment (DDA) (full VM, dedicated GPU)
3. KVM + VFIO (Linux, full GPU passthrough)
4. Docker / Windows Containers (process isolation, no GPU passthrough in Phase 1)
5. Bare metal (no isolation — unacceptable)

## Decision

**Phase 1 MVP:** Use Windows Sandbox API.

Windows Sandbox provides:
- Ephemeral filesystem (wiped on exit) — satisfies the cleanup requirement
- Process-level isolation from the host desktop
- DirectX 12 capable (GPU rendering works inside sandbox via WDDM paravirtualization)
- Simple API: drop a `.wsb` config file, call `Start-Process WindowsSandbox`
- Ships with Windows 10/11 Pro and Enterprise
- Zero additional licensing cost for hosts

**Phase 2:** Migrate to Hyper-V with Discrete Device Assignment (DDA).

DDA provides dedicated GPU passthrough, enabling:
- Native GPU performance (not paravirtualized DirectX)
- Multiple VMs can each get a dedicated GPU partition
- Required for compute workloads (AI/ML) that need full CUDA/ROCm
- Isolation at the hypervisor level rather than app container level

## Consequences

**Positive (Windows Sandbox Phase 1):**
- Available on all Windows 10/11 Pro machines that are likely hosts
- No VM overhead beyond the sandbox container
- Proven ephemeral: Microsoft guarantees state wipe on exit
- Fast startup (5-15 seconds vs 30-60 seconds for full VM)
- DirectX rendering works — games run inside sandbox

**Negative (Windows Sandbox Phase 1):**
- GPU is shared with host (paravirtualized WDDM), not dedicated
- GPU compute (CUDA, ROCm, OpenCL) may have reduced capability
- Cannot support Linux guests
- Cannot isolate network from host on the same LAN in all configurations
- Windows Sandbox only available on Pro/Enterprise SKUs — not Windows Home

**Negative (DDA Phase 2 risk):**
- Requires Hyper-V hardware support (IOMMU)
- Host cannot use the GPU simultaneously with the VM
- More complex agent code to manage VM lifecycle vs sandbox

## Alternatives Considered

**Docker / Windows Containers:** Does not support GPU passthrough in the current stable release. The `--gpus` flag on Windows requires WSL2 and doesn't give isolated DirectX contexts. Rejected.

**Bare metal (no isolation):** User workload runs directly on host OS. Security is unacceptable — user can access host files, spy on host desktop, install malware. Rejected permanently.

**KVM + VFIO:** Excellent isolation and true GPU passthrough on Linux. But Phase 1 targets Windows hosts where GPU drivers are better supported for gaming. Deferred to Phase 3 (Linux host support).

**Qubes OS:** Too exotic for hosts to run. Not viable as a product requirement.
