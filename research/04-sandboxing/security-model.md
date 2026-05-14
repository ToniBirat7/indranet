# IndraNet Security Model — Threat Analysis

**Owner:** @security  
**Last updated:** 2025-05-14  
**Status:** Initial draft — must be reviewed before any session lifecycle code ships.

---

## Trust Model

```
FULLY TRUSTED: IndraNet backend infrastructure, Stripe
PARTIALLY TRUSTED: Host operators (verified identity, financial stake)
UNTRUSTED: Session users (anonymous, may be adversarial)
```

The sandbox exists because users are untrusted. Every threat below assumes a motivated adversarial user.

---

## Threat 1: Sandbox Escape

**Description:** A user exploits a vulnerability in Windows Sandbox isolation to execute code on the host OS.

**Attack vectors:**
- Exploiting a kernel vulnerability via the Windows Sandbox hypervisor interface
- Exploiting a bug in the WDDM GPU virtualization layer
- Exploiting a bug in the IndraNet host agent's IPC handler

**Mitigations:**
- Windows Sandbox is maintained by Microsoft; kernel CVEs are patched via Windows Update. Hosts must have auto-updates enabled.
- The agent's IPC handler (named pipe) must validate all input strictly — no eval, no shell injection.
- Agent process runs with reduced privileges (no admin rights) where possible.
- Session user cannot interact with the host desktop even if sandbox escapes to a non-privileged process — agent window is hidden.

**Residual risk:** MEDIUM. Kernel exploits are rare but exist. No sandbox is 100% secure against zero-days. The financial stake required to become a host (Stripe KYC, payout dependency) deters hosts from being attack infrastructure.

---

## Threat 2: Cryptocurrency Mining Abuse

**Description:** A user rents a GPU session to mine cryptocurrency instead of the stated use case.

**Attack vectors:**
- Run a crypto miner inside the sandbox (XMRig, NBminer, etc.)
- Pause the workload when GPU monitoring detects high load

**Mitigations:**
- Monitor GPU utilization inside the sandbox and compare to expected workload profile
- Detect known miner process names in the sandbox process list (XMRig, NBminer, PhoenixMiner, etc.)
- Rate-limit sessions from accounts without prior session history
- Hosts can set "allowed applications" list

**Detection:** GPU running at 100% with no screen rendering activity (no frame changes on DXGI output).

**Residual risk:** MEDIUM. Difficult to fully prevent at Phase 1. Phase 2 should add ML-based workload classification.

---

## Threat 3: Illegal Content Hosting

**Description:** A user uses the session to view or distribute illegal content (CSAM, pirated software, etc.).

**Attack vectors:**
- Any application that can display content can display illegal content

**Mitigations:**
- Terms of Service prohibition with account suspension
- Host can terminate session at will via the desktop app
- Sessions are ephemeral — no content persists after session end
- Platform is not a content CDN; content is not routed through our servers (P2P)

**Note:** IndraNet cannot pre-screen content inside sessions (that would require recording all sessions, which is a privacy violation). Post-session forensics are impossible (ephemeral sandbox). This is a Terms of Service enforcement problem, not a technical one.

**Residual risk:** LOW-MEDIUM. Platform is not meaningfully more enabling than any browser or VPN.

---

## Threat 4: Host Spying on User

**Description:** A host uses the agent to record, screenshot, or otherwise surveil the user's session.

**Attack vectors:**
- A malicious host operator modifies the agent binary to log screenshots
- Host accesses the DXGI capture stream outside the encode pipeline

**Mitigations:**
- Agent binary is distributed and code-signed by IndraNet — hosts cannot modify it without breaking the signature
- DXGI stream is consumed directly by the encoder; the agent must not write frames to disk or network except via WebRTC
- Code signing enforcement: backend refuses connections from unsigned or tampered agent binaries (Phase 2 — attestation)

**Residual risk:** MEDIUM at Phase 1 (no attestation). LOW at Phase 2 with agent attestation.

---

## Threat 5: Persistent Malware on Host

**Description:** A user plants persistent malware on the host machine via the session.

**Attack vectors:**
- Exploit sandbox escape (see Threat 1) and write to host filesystem
- Exploit a bug in file staging (mapped folder with write access)

**Mitigations:**
- Mapped folders are read-only (see sandbox-spec.md)
- Sandbox is fully ephemeral — any writes inside sandbox are wiped on exit
- No write-access mapped folders in Phase 1

**Residual risk:** LOW if sandbox escape is prevented.

---

## Threat 6: Session Token Theft / Replay

**Description:** An attacker intercepts a user's session JWT and uses it to connect to an active session.

**Attack vectors:**
- Man-in-the-middle on HTTPS (requires CA compromise — very low probability)
- Malicious browser extension stealing localStorage JWT
- JWT leaked via referrer header or logging

**Mitigations:**
- JWTs are short-lived (24h for user auth; 1h for session tokens)
- Session tokens are single-use per WebSocket connection — the backend marks them as consumed on connect
- JWTs must not be logged anywhere (structured logging must redact `Authorization` headers)
- `httpOnly` cookies for session tokens in the web client (prevents JS access)

**Residual risk:** LOW.

---

## Threat 7: LAN Network Attack from Sandbox

**Description:** A user uses the sandbox's network access to attack other devices on the host's LAN.

**Attack vectors:**
- Port scan the host's LAN from inside the sandbox
- Exploit unpatched devices on the LAN (NAS, IoT devices, printers)

**Mitigations:**
- Phase 2: Network isolation config in `.wsb` file to restrict sandbox network access to internet-only
- Phase 1: Document this risk in host onboarding — hosts in corporate/sensitive networks should not run Phase 1

**Residual risk:** HIGH at Phase 1 (no network isolation). MEDIUM at Phase 2 (Windows Sandbox network restrictions are coarse).

---

## Stripe Webhook Spoofing

**Description:** An attacker sends a fake Stripe webhook to authorize a session without payment.

**Mitigation:** All webhook payloads must be verified using `stripe.ConstructEvent()` with the webhook signing secret. Reject any webhook that fails HMAC-SHA256 verification. This is a hardcoded requirement — see security invariants in `CLAUDE.md`.

**Residual risk:** NEGLIGIBLE with proper signature verification.

---

## Security Review Gates

These events require @security review before code ships:

| Gate | Trigger |
|------|---------|
| Session lifecycle code | Any PR touching `sessions.go`, `billing/engine.go` |
| Sandbox management code | Any PR touching `windows_sandbox.cpp` |
| WebRTC signaling | Any PR touching `signaling/hub.go` or `signaling/client.go` |
| Payment processing | Any PR touching `payments.go` or Stripe webhook handling |
| Agent IPC | Any PR touching `agent_ipc.cpp` |
