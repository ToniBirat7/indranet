# GPU Passthrough Comparison

## Options by Phase

| Option | Phase | OS | GPU Performance | Complexity | Notes |
|--------|-------|----|----------------|------------|-------|
| Windows Sandbox WDDM paravirt | 1 | Windows | 70-90% | Low | No CUDA |
| Hyper-V DDA | 2 | Windows | 95-98% | Medium | CUDA works, exclusive GPU |
| KVM + VFIO | 3 | Linux | 95-98% | High | Full community support |
| Docker GPU | Never | Any | 60-80% | Low | No DirectX graphics |

## Decision
See ADR-0003 and ADR-0004 for the architectural decisions.

Phase 1: Windows Sandbox (shared GPU, no CUDA)
Phase 2: Hyper-V DDA (dedicated GPU, full CUDA)
Phase 3: KVM + VFIO (Linux hosts)
