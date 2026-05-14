# Hyper-V Discrete Device Assignment (DDA) Notes

## Relevance
Phase 2 technology. DDA enables passing a physical GPU directly to a Hyper-V VM, bypassing paravirtualization.

## What DDA Provides
- The VM gets exclusive, native access to the GPU
- Full CUDA/ROCm support (native driver in VM)
- Near-native GPU performance (5-10% overhead vs bare metal)
- Required for AI/ML workloads in Phase 2

## Requirements
- Host must have IOMMU support (Intel VT-d / AMD-Vi) enabled in BIOS
- GPU must be assigned to a separate IOMMU group
- Host cannot use the assigned GPU while it's attached to a VM
- Requires Windows Server or Windows 11 Enterprise (not Pro) for Hyper-V with DDA

## Implications for Phase 2 Business Model
- Hosts with DDA: can have multiple GPUs and rent each independently
- Single-GPU hosts: must take GPU offline during session (loses gaming ability)
- Multi-GPU hosts (e.g., gaming PC + mining rig): ideal for DDA

## TODO
Research when Phase 2 begins. Key questions:
- Can Hyper-V DDA work on Windows 11 Pro (not just Enterprise)?
- What is the GPU teardown/attach time for DDA?
- How do we manage GPU assignment from the host agent?
