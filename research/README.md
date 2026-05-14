# Phase 0 Research

This directory contains technical research notes for IndraNet's Phase 0 validation.

## Goals

Before writing production code, we must answer these questions with running experiments:

1. Can we capture and encode a screen at 1080p60 with <5ms combined latency?
2. Can WebRTC deliver that stream with <50ms end-to-end latency on LAN?
3. Does a GPU-accelerated app run acceptably inside Windows Sandbox?
4. Can a Stripe payment reliably gate access to a WebRTC stream?

## Research Areas

| Area | Directory | Status |
|------|-----------|--------|
| Screen capture | 01-screen-capture/ | NOT STARTED |
| Video encoding | 02-video-encoding/ | NOT STARTED |
| Streaming transport | 03-streaming-transport/ | NOT STARTED |
| Sandboxing | 04-sandboxing/ | NOT STARTED |
| Input forwarding | 05-input-forwarding/ | NOT STARTED |
| Payment | 06-payment/ | NOT STARTED |
| Competitors | 07-competitors/ | NOT STARTED |

## Format

Each research document follows this template:

```markdown
## Question
What are we trying to learn?

## Method
How did we investigate this?

## Findings
What did we discover?

## Recommendation
What should we do based on this?

## Open Questions
What do we still not know?
```
