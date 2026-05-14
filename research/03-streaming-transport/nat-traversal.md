# NAT Traversal Strategy

## The Problem
Most hosts and users are behind NAT routers. WebRTC requires ICE (Interactive Connectivity Establishment) to find a working network path.

## ICE Candidate Priority
1. **Host candidates** (direct LAN connection): lowest latency, works on same LAN
2. **Server-reflexive candidates** (STUN): works through most home NATs
3. **Relay candidates** (TURN): always works, highest latency (+20-40ms)

## STUN Server
For PoC and Phase 1 dev: `stun:stun.l.google.com:19302` (free, reliable)
For production: deploy own coturn or use Cloudflare TURN (commercial)

## TURN Server
Required for ~15-20% of connections (symmetric NAT, corporate firewalls).

Phase 1: Use a managed TURN service (Cloudflare TURN, Twilio TURN, or Metered.ca)
Phase 2: Deploy coturn on dedicated servers in multiple regions

## Estimated TURN Traffic
Assumption: 20% of sessions require TURN relay. At 15 Mbps per session, 1 hour:
`0.20 * 15 Mbps * 3600s = ~13.5 GB per relayed session-hour`

TURN bandwidth costs must be factored into pricing for high-volume use.

## TODO
Research coturn configuration and TURN cost models for Phase 2 planning.
