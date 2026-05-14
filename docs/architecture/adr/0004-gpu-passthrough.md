# ADR-0004: Defer Full GPU Passthrough to Phase 2

**Status:** Accepted  
**Date:** 2025-05-14  
**Author:** @architect

## Context

GPU passthrough is technically complex and has significant Phase 1 scope implications. The Phase 1 MVP needs to prove the core hypothesis: streaming a GPU-accelerated app over WebRTC, gated by a payment, with isolation. It does not need to prove maximum GPU performance.

The tradeoff: build the hardest thing first (full GPU passthrough via Hyper-V DDA) or ship a working MVP with shared GPU and upgrade later.

## Decision

Phase 1 MVP uses shared GPU via Windows Sandbox paravirtualized WDDM. Hyper-V DDA is deferred to Phase 2.

**What this means for Phase 1:**
- The sandbox can render DirectX games — performance is somewhat reduced but functional
- NVENC is still available inside the sandbox (hardware encoding via the paravirtualized GPU)
- Users can play games, run rendering software, and interact with GPU applications
- AI/ML compute workloads (CUDA) may not work reliably in Phase 1

**Measurement gate before Phase 2:** If PoC 03 shows that Windows Sandbox GPU performance is within 80% of bare metal for target workloads, we defer DDA to Phase 2 as planned. If performance is unacceptable, this ADR should be revisited.

## Consequences

**Positive:**
- Phase 1 agent code is significantly simpler (no Hyper-V VM management)
- No IOMMU requirement on host hardware
- Faster to ship, faster to validate market

**Negative:**
- Phase 1 cannot fully serve AI/ML compute market
- GPU is shared with host — host gaming performance degrades during active sessions
- Requires Phase 2 rework of the agent's sandbox management subsystem

## Alternatives Considered

**Ship Hyper-V DDA in Phase 1:** Adds 8-12 weeks of agent development. The core hypothesis (streaming + payment) does not require dedicated GPU. Rejected.
