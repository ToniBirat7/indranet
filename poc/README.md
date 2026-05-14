# IndraNet — Proof of Concept Experiments

Four experiments to validate the core technical hypotheses before building the MVP.

## PoC Overview

| # | Name | Goal | Success Criteria | Status |
|---|------|------|-----------------|--------|
| 01 | Screen Capture + Encode | Capture + NVENC encode | <5ms, 60fps sustained | NOT STARTED |
| 02 | WebRTC Stream | End-to-end video over WebRTC | <50ms LAN latency | NOT STARTED |
| 03 | Sandbox GPU | GPU app inside Windows Sandbox | Confirmed hardware accel | NOT STARTED |
| 04 | Payment Gate | Stripe payment → stream access | Pay $1 → get stream | NOT STARTED |

## Order of Execution

Run the PoCs in order — each builds on the previous:
1. PoC 01 proves capture+encode works
2. PoC 02 proves the stream can be delivered (uses output from PoC 01)
3. PoC 03 proves the sandbox is viable
4. PoC 04 proves payment works end-to-end

## Definition of Success

All four PoCs pass → Phase 1 MVP development begins.

Any PoC fails → research the failure, find a solution, adjust architecture if needed.
