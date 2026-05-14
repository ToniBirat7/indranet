# KVM + VFIO GPU Passthrough Notes

## Relevance
Phase 3 (Linux host support). KVM + VFIO is the Linux equivalent of Hyper-V DDA.

## Overview
VFIO (Virtual Function I/O) allows a physical GPU to be bound to a VM via KVM with near-native performance. This is widely used in the homelab/gaming community.

## Requirements
- Linux kernel 5.10+
- IOMMU support (Intel VT-d or AMD-Vi)
- GPU in separate IOMMU group
- `vfio-pci` kernel module
- libvirt/QEMU for VM management

## Considerations
- More complex agent code on Linux
- PipeWire for audio capture (replacing WASAPI)
- uinput for virtual input injection (replacing ViGEm)
- Excellent community documentation (Level1Techs forums, ArchWiki)

## TODO
Defer to Phase 3 planning.
