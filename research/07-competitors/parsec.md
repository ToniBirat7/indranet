# Parsec Analysis

## What They Do
Remote desktop with gaming-quality streaming. Used by game studios for remote work and by gamers to stream their home PC from anywhere. Not a marketplace.

## Strengths
- Excellent streaming latency (they are the gold standard, ~25ms LAN)
- Windows and Linux clients
- WebRTC-based transport
- Well-funded (acquired by Unity in 2021)

## Weaknesses
- Not a marketplace — you connect to YOUR machine, not a stranger's
- No payment infrastructure
- No sandboxing (full desktop access)
- Requires host machine to be always-on and managed by the user

## Technical Lessons
Parsec proves the streaming technology works. Their latency numbers are our benchmark. Their architecture (P2P WebRTC after STUN) is what we're building.

## What They Don't Do
- No host marketplace (you can't rent Parsec capacity from a stranger)
- No payment gating
- No sandboxing/isolation
- No idle hardware monetization

## Relevance
Parsec is the closest thing to IndraNet's streaming quality. We should test Parsec's latency as our benchmark to beat.
