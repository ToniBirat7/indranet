# Harness: Peer-to-Peer Interactive Compute Marketplace
## Deep Research Report — Feasibility, Competitive Landscape, Technical Architecture & Roadmap

> Prepared for: Founder / Ideation Phase
> Research Date: May 2026
> Status: Pre-Development Concept Validation

---

## Table of Contents

1. Executive Summary
2. The Problem Being Solved
3. Existing Competitive Landscape
4. Your Unique Differentiation — The Gap in the Market
5. Technical Feasibility Analysis
6. Core Technical Architecture
7. Key Technical Challenges & Bottlenecks
8. Security Architecture Deep Dive
9. Streaming & Networking Layer
10. Business Model & Pricing
11. Legal & Compliance Risks
12. Tech Stack Recommendation
13. Development Roadmap (Phased)
14. Final Verdict & Recommendation

---

## 1. Executive Summary

The idea is technically feasible, commercially viable, and occupies a real gap in the market. The closest equivalent — the **Parsec Cloud Gaming Marketplace** — was shut down because it relied on centralized cloud (AWS), making it expensive for users and unprofitable at scale. **Your version democratizes the supply side**: instead of renting AWS servers, real people with idle gaming rigs, workstations, and GPU-powered laptops become the supply.

The **compute marketplace space** (GPU rentals for AI/ML) is booming — platforms like Vast.ai, RunPod, SaladCloud, and Hyperbolic collectively do millions in monthly revenue. But they only serve headless batch workloads (AI training, rendering queues). **Nobody has cracked interactive, GUI-based, peer-to-peer compute rentals from personal home hardware at scale.** That is your market.

However, building this is genuinely hard. The biggest challenges are:
- **Security** — sandboxing a remote user's session so the host machine is untouchable
- **Latency** — home internet upload bandwidth is asymmetric and unpredictable
- **Trust** — both hosts and users need strong guarantees before any money changes hands
- **Reliability** — a home PC is not a data center; session stability is not guaranteed

This report gives you everything you need: what exists, what doesn't, how to build it, and in what order.

---

## 2. The Problem Being Solved

### Supply Side — Idle Compute is Wasted

GPU utilization rates in even enterprise environments average only 15–30% of capacity. For personal computers — gaming rigs, video editing workstations, AI development machines — the idle time is far worse. A gaming PC with an RTX 4090 sits completely unused for 8–12 hours every night. That's $1,500–$2,500 worth of hardware doing nothing.

Cloud GPU providers like AWS, Azure, and GCP charge $2–$6/hour for similar performance. The hardware owner earns nothing. The market is broken.

### Demand Side — Compute Access is Expensive and Inaccessible

A student in Kathmandu who wants to:
- Play GTA V on their low-spec laptop
- Edit a 4K video for a client
- Fine-tune an AI model for a project
- Run a stable diffusion pipeline for design work

...currently has two options: buy expensive hardware, or pay expensive cloud services. Monthly cloud gaming subscriptions (Shadow PC) start at $20–30/month for mediocre GPU performance. GPU spot instances on Vast.ai are cheaper for batch work, but not useful for interactive sessions.

**The whitespace**: cheap, on-demand, interactive, GUI-driven access to real consumer hardware — not a data center rack.

---

## 3. Existing Competitive Landscape

Understanding who already exists is critical. Here is the full map:

### 3.1 Centralized Cloud Gaming (Interactive, but Centralized)

| Platform | Model | Pricing | Status |
|---|---|---|---|
| **Shadow PC** | Dedicated cloud VM (data center) | ~$30/mo | Active, mainstream |
| **GeForce NOW** | Game-streaming from NVIDIA servers | $10–$20/mo | Active, large |
| **Xbox Cloud Gaming** | Microsoft data centers | Bundled with Game Pass | Active |
| **Boosteroid** | Centralized cloud gaming | $10/mo | Active |
| **AirGPU** | Hourly cloud GPU rental + Parsec | $0.65–$1.20/hr | Active, niche |

**What they solve**: Interactive desktop access. **What they don't solve**: They are centralized data centers — expensive infrastructure, fixed GPU selection, no role for idle personal hardware. Costs are passed to users.

### 3.2 Headless GPU Compute Marketplaces (Decentralized Supply, but Not Interactive)

| Platform | Model | Use Case | Status |
|---|---|---|---|
| **Vast.ai** | P2P GPU marketplace | AI/ML batch training | Active, prominent |
| **RunPod** | GPU cloud + marketplace | AI inference, training | Active, growing fast |
| **SaladCloud** | Distributed consumer GPUs | AI inference | Active |
| **Hyperbolic** | Decentralized GPU network (Hyper-dOS) | AI workloads | Active, 2024 launch |
| **MicroDC** | Pure P2P GPU bid marketplace | AI/ML batch | Active, early stage |
| **Aethir** | Decentralized cloud hosting (blockchain) | AI + cloud gaming | Active |
| **Hivenet** | Pooled compute network | AI/ML | Active |

**What they solve**: Cheap GPU access for developers and researchers. **What they don't solve**: None of these offer interactive GUI sessions. You can't play a game, edit a video, or run a desktop application — you can only run containerized Docker workloads. No screen. No mouse. No keyboard.

### 3.3 Remote Desktop / Screen Sharing Tools (Interactive, but Not a Marketplace)

| Tool | Model | Key Feature | Gaming-Optimized |
|---|---|---|---|
| **Parsec** | Remote desktop (personal hardware) | 240 FPS, H.265, P2P | Yes — best in class |
| **Moonlight** | Open-source, NVIDIA GameStream | Very low latency | Yes |
| **AnyDesk** | Remote desktop (DeskRT codec) | Lightweight, fast | Partial |
| **Steam Remote Play** | Steam-native streaming | Steam library only | Yes |
| **RustDesk** | Open-source AnyDesk alternative | Self-hosted | No |
| **NoMachine** | Remote desktop | Enterprise-grade | Partial |

**What they solve**: Connecting to YOUR OWN machine remotely. **What they don't solve**: There is no marketplace. You can't monetize your machine. You can't browse available compute and pay for it.

### 3.4 The Closest Thing That Existed — And Why It Closed

**Parsec Cloud Gaming Marketplace** was the most similar idea ever launched. It let users spin up a cloud gaming PC on AWS/Paperspace via Parsec's streaming protocol — hourly, pay-as-you-go. It was critically acclaimed by the gaming community.

It was shut down because: it relied on AWS/Paperspace as the compute backend. Parsec was essentially a reseller of centralized cloud compute. The margins were thin. The GPU costs were centralized infrastructure costs — not idle peer hardware.

**The key insight**: Parsec had the streaming layer right, but the wrong supply model. Your idea flips the supply model to idle personal hardware — which dramatically changes the unit economics.

---

## 4. Your Unique Differentiation — The Gap in the Market

The gap your idea fills is a very specific combination that no current platform offers:

**Interactive GUI sessions** + **Peer-to-peer personal hardware supply** + **Sandboxed host protection** + **Marketplace with PAYG pricing**

Plotted on two axes:

```
                    INTERACTIVE (GUI)
                          |
    Shadow PC             |         << YOUR IDEA >>
    GeForce NOW           |     (Harness Marketplace)
    Boosteroid            |
                          |
CENTRALIZED ——————————————+—————————————— PEER-TO-PEER
                          |
                          |     Vast.ai, RunPod,
                          |     SaladCloud, MicroDC
                          |
                    HEADLESS (Batch)
```

This is a genuinely unoccupied position in the market. The question is whether the technical and operational challenges can be overcome to make it work.

---

## 5. Technical Feasibility Analysis

### 5.1 The Core Technical Stack — Does It Exist?

Every individual component you need already exists and is mature. The challenge is integrating them into a cohesive, production-grade system.

| Component | Technology | Maturity |
|---|---|---|
| Screen capture & encoding | NVENC, AMD VCE, VAAPI, H.264/H.265/AV1 | Production-ready |
| Low-latency streaming | WebRTC, QUIC/MoQ, RTP | Production-ready |
| Input (keyboard/mouse) forwarding | USB/HID emulation, virtual input drivers | Production-ready |
| Sandboxing / isolation | Hyper-V, KVM, Windows Sandbox, VirtualBox | Mature |
| GPU passthrough into VM | NVIDIA vGPU, AMD MxGPU, PCIe passthrough | Complex but viable |
| Payment gating | Stripe, crypto escrow, smart contracts | Mature |
| P2P signaling / NAT traversal | STUN/TURN, ICE, libp2p | Mature |

Nothing requires inventing new technology. The difficulty is the architecture that ties everything together and makes it seamless.

### 5.2 GPU Passthrough — The Critical Technical Requirement

This is the hardest part of the entire system. When a user rents compute from a host, they need access to the host's GPU inside a sandboxed VM. Simply running a VM without GPU access gives you terrible performance — no gaming, no video editing, no real compute.

**Options in order of complexity:**

**Option A — NVIDIA vGPU (Best for gaming/graphics)**
NVIDIA's Virtual GPU (vGPU) technology allows a single physical GPU to be partitioned and shared among multiple VMs. This is production-grade in enterprise. However, it requires NVIDIA vGPU licenses (expensive, ~$500+/year per GPU) and only works on data center GPUs (A-series, T-series) — not consumer RTX cards.

**Option B — GPU Passthrough (PCIe Passthrough)**
Pass the entire physical GPU directly to a single VM. This gives the VM near-native GPU performance. Works with consumer cards. The host cannot use the GPU while it's passed through. Requires IOMMU support in the CPU/motherboard. Linux hosts support this well with KVM/QEMU. Windows hosts require Hyper-V with DDA (Discrete Device Assignment). This is the most practical approach for your use case.

**Option C — API Interception/Forwarding (VirtualGL / Looking Glass)**
Tools like Looking Glass allow the guest VM to share the host's GPU by intercepting graphics API calls. Very low overhead but complex to set up and maintain. Used in the gaming VM community (VFIO community). Very promising for your use case but requires significant engineering to productize.

**Option D — Application-Level Sandbox (Simpler but Less Isolated)**
Instead of full VM isolation, run the user's session in an application-level sandbox (Windows Sandbox, Firejail, AppArmor + namespace isolation). Lighter weight, easier to implement, but weaker security guarantees. GPU access is shared directly — good performance, weaker isolation.

**Recommended Approach for MVP**: Start with Option D (application-level sandbox with strict policies) for the initial release to get to market faster. Migrate to Option B (PCIe passthrough + Hyper-V DDA) for production.

---

## 6. Core Technical Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    HARNESS PLATFORM                              │
│                                                                  │
│  ┌─────────────────┐          ┌─────────────────────────────┐   │
│  │   HOST SIDE     │          │       USER SIDE             │   │
│  │                 │          │                             │   │
│  │  Host Machine   │ ◄──────► │  Harness App (Thin Client)  │   │
│  │                 │  QUIC/   │  (Win/Mac/Linux/Web)        │   │
│  │ ┌─────────────┐ │  WebRTC  │                             │   │
│  │ │ Sandbox/VM  │ │          │  - Screen rendering         │   │
│  │ │ (Isolated)  │ │          │  - Keyboard/mouse input     │   │
│  │ │             │ │          │  - Audio                    │   │
│  │ │  User Apps  │ │          │  - File upload interface    │   │
│  │ │  GPU Access │ │          │                             │   │
│  │ └─────────────┘ │          └─────────────────────────────┘   │
│  │                 │                       ▲                     │
│  │  Harness Agent  │                       │                     │
│  │  (C++/Rust)     │                       │                     │
│  └────────┬────────┘                       │                     │
│           │                                │                     │
│           ▼                                ▼                     │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    HARNESS BACKEND                        │   │
│  │                                                          │   │
│  │  ┌────────────┐  ┌──────────┐  ┌────────────────────┐   │   │
│  │  │ Marketplace│  │ Payment  │  │  Signaling Server   │   │   │
│  │  │  API       │  │ Gateway  │  │  (STUN/TURN/ICE)    │   │   │
│  │  └────────────┘  └──────────┘  └────────────────────┘   │   │
│  │                                                          │   │
│  │  ┌────────────┐  ┌──────────┐  ┌────────────────────┐   │   │
│  │  │ Session    │  │ Billing  │  │  File Transfer      │   │   │
│  │  │ Manager    │  │ Engine   │  │  (Pre-staging)      │   │   │
│  │  └────────────┘  └──────────┘  └────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

### 6.1 The Host Agent (Harness Agent)

This is the software the host installs on their machine. It is the most critical piece.

**Responsibilities:**
- Register the host's machine and capabilities (GPU model, RAM, CPU, installed software) with the Harness backend
- Launch and manage the isolated sandbox/VM environment for user sessions
- Capture the sandbox's screen output, encode it (H.264/H.265/AV1 via hardware encoder), and stream it to the user
- Receive keyboard/mouse/gamepad input from the user and inject it into the sandbox via virtual HID
- Forward audio from the sandbox to the user
- Monitor session health, billing clock, and session termination conditions
- Upload/download files from the pre-staging area into/out of the sandbox
- Enforce policy (no access outside sandbox boundaries, resource limits)

**Language**: C++ or Rust. C++ has more ecosystem support for the codec/capture libraries (FFMPEG, NVENC SDK). Rust is safer and more modern.

**Key Libraries:**
- NVENC SDK / AMD AMF for hardware-accelerated video encoding
- FFmpeg for codec handling
- libyuv for pixel format conversion
- libnice or libjuice for ICE/NAT traversal
- quiche (Cloudflare) or msquic (Microsoft) for QUIC transport

### 6.2 The Sandbox Environment

**For Windows Hosts (Primary Target):**

The most practical architecture uses **Hyper-V with Nested Virtualization** or **Windows Sandbox** as the isolation layer, with GPU access via Discrete Device Assignment (DDA).

- Windows Sandbox: lightweight, built into Windows 10/11 Pro. Starts in seconds. Based on same dynamic base image technology. Does not persist state across sessions (good for security). Limited GPU support but improving.
- Hyper-V with DDA: passes a physical GPU directly to the VM. Full GPU performance. Requires the host to relinquish the GPU for the duration. Best for high-performance gaming sessions.

**For Linux Hosts:**

KVM/QEMU with PCIe passthrough (VFIO) is the gold standard. The Linux VFIO community has refined this for years. Very high GPU passthrough performance achievable. Requires IOMMU-enabled CPU and motherboard.

**Session Lifecycle:**
1. User pays and session is authorized → backend sends token to host agent
2. Host agent spins up sandbox (pre-warmed pool for fast startup)
3. GPU is bound to sandbox (if passthrough mode)
4. Screen capture begins, QUIC stream is established to user
5. User works normally inside the sandbox
6. Session ends (timeout, user terminates, or balance depleted) → sandbox is torn down and disk state wiped
7. Host agent confirms session end to backend → billing finalized

### 6.3 File Transfer — The "Harness" Pre-Staging Layer

Your concept of pre-uploading files before a session is architecturally sound. Here's how it works in practice:

**Pre-Session File Staging:**
- User uploads files (video clips, game saves, datasets) through the Harness app or web portal BEFORE the session starts
- Files are uploaded to a temporary encrypted object store (S3-compatible, e.g., Cloudflare R2 — no egress fees)
- When the session begins, the agent mounts these files as a read-only or copy-on-write virtual disk inside the sandbox
- Files the user creates inside the sandbox are captured to a "save state" partition
- On session end, the user can download their output files from the platform

This avoids the latency and bandwidth cost of transferring large files during the session itself.

### 6.4 Payment Flow

```
User browses marketplace → selects host listing → sees price/hr
        ↓
User deposits funds (credit card / crypto) into Harness wallet
        ↓
User initiates session → 50% of estimated session cost held in escrow
        ↓
Session runs → billing ticks per-minute in backend
        ↓
Balance check every 60 seconds → if < 5 minutes remaining, warn user
        ↓
Balance = $0 → session gracefully terminates (15-second warning)
        ↓
Final billing calculated → host receives payout (minus platform fee)
        ↓
User's output files available for download for 24 hours
```

---

## 7. Key Technical Challenges & Bottlenecks

These are the honest, hard problems you will face. Know them before you start.

### 7.1 Home Upload Bandwidth — The Biggest Bottleneck

This is the number one technical challenge. Unlike a data center with 10 Gbps symmetric fiber, a home internet connection in most countries has asymmetric bandwidth:
- Download: 100–500 Mbps (good)
- **Upload: 10–50 Mbps** (this is what streams to the user)

For a good gaming experience, you need roughly:
- 1080p 60fps: ~10–15 Mbps
- 1440p 60fps: ~20–30 Mbps
- 4K 60fps: ~40–60 Mbps

This means a single user on a typical home connection can get 1080p streaming. Multiple simultaneous users per host is difficult unless the host has fiber (symmetric) or a high-upload plan. This naturally limits most hosts to one user session at a time — which is fine for your model (it matches the sandbox model anyway).

**Mitigation:**
- The Harness agent tests upload bandwidth during registration and caps the session resolution/bitrate accordingly
- Adaptive bitrate encoding (like Netflix) adjusts on the fly
- Hosts can list their capabilities honestly — "1080p max" or "720p max"
- Incentivize hosts with high-speed fiber connections

### 7.2 Latency and Geographic Distance

The user experience degrades badly when the physical distance between user and host is large. At 100ms+ round-trip latency, gaming becomes unplayable. Video editing is tolerable at higher latency, but AI/ML jobs (headless) are fine at any latency.

The platform cannot control where hosts and users are located. A user in India connecting to a host in Germany will have 150–200ms latency — unacceptable for gaming.

**Mitigation:**
- Display latency estimate to users before booking (ping the host from the platform's relay servers)
- Filter/sort marketplace listings by geographic distance / latency
- Match users and hosts by region (add "Region" tag to listings: South Asia, Southeast Asia, etc.)
- Incentivize hosts in underserved regions with higher revenue shares

### 7.3 Host Availability and Reliability

A home PC is not a server. The host may:
- Reboot their computer unexpectedly
- Have a power outage
- Update Windows and auto-restart
- Decide they want to use the computer themselves mid-session

**Mitigation:**
- Agent daemon prevents Windows updates during active sessions
- Host sets "availability windows" — sessions can only start during these windows
- Users get a session credit/refund if the host disconnects without warning
- Build a reputation/rating system — hosts with frequent unexpected disconnects get flagged
- Require hosts to acknowledge they won't disrupt sessions when "online"

### 7.4 ISP Terms of Service

Many residential ISP plans explicitly prohibit using the connection for commercial hosting or running servers. A host renting out compute via Harness could technically violate their ISP's ToS.

**Reality check**: In practice, ISPs rarely enforce this against individual users unless they are consuming extraordinary bandwidth. But it is a legal grey area the platform must acknowledge.

**Mitigation:**
- Terms of Service clearly state hosts are responsible for compliance with their ISP's policies
- Educate hosts; the actual traffic is no different from a heavy gaming session + video call
- Business/fiber internet plans generally allow this

### 7.5 GPU Passthrough Complexity for Hosts

Configuring PCIe GPU passthrough (VFIO) requires technical knowledge and BIOS changes that a non-technical host cannot do. This dramatically limits the host pool if it is required.

**Mitigation (Phased approach):**
- Phase 1: No VM passthrough. Use application-level sandbox with shared GPU access. Simpler, broader host compatibility. Lower security guarantee but acceptable for MVP.
- Phase 2: Offer Hyper-V DDA guide in app for technical hosts. Unlock "dedicated GPU" badge on listings for higher pricing.
- Phase 3: Build automated GPU passthrough configuration tool (like how VFIO-autoinstall scripts work on Linux).

---

## 8. Security Architecture Deep Dive

Security is the hardest and most important problem. If a malicious user can escape the sandbox and access host files, the platform is dead. Here is the full threat model and mitigations.

### 8.1 Threat Model

**Threat 1 — Sandbox Escape**
User finds a vulnerability in the VM/sandbox hypervisor and escapes to the host OS. Access to host files, passwords, personal data.

**Mitigation**: Use hardware-enforced VM boundaries (Hyper-V Type 1 hypervisor or KVM). Keep sandbox OS fully patched. Disable all unnecessary host-to-guest communication channels. Only expose specific virtual devices (GPU output, virtual HID, virtual audio). This is the same threat model that powers cloud data centers hosting thousands of tenant VMs — it is a solved problem with well-understood mitigations.

**Threat 2 — Network Attacks from Sandbox**
User inside sandbox tries to attack other devices on the host's local network (the host's phone, NAS, router).

**Mitigation**: Give the sandbox a separate virtualized network adapter behind a NAT. The sandbox should see only a virtual gateway with no route to the host's LAN. Only allow outbound internet traffic from the sandbox, no LAN access.

**Threat 3 — Crypto Mining Abuse**
User rents compute, turns off screen, runs a cryptocurrency miner to extract maximum GPU value beyond what they are paying for.

**Mitigation**: Agent monitors GPU utilization pattern and flags anomalous sustained 100% GPU usage without corresponding interactive display activity. Set session resource limits. Rate-limit sessions per user.

**Threat 4 — Illegal Content**
User runs illegal operations inside the sandbox (hosting illegal files, CSAM, etc.).

**Mitigation**: This is the same problem every cloud provider faces. The sandbox is ephemeral — all data is wiped at session end. Log session metadata (timing, resource usage, file sizes) but not content. Clear DMCA/abuse policy. Cooperate with legal requests. Implement per-user KYC for high-risk categories.

**Threat 5 — Host Spying on User**
Host installs keyloggers or screen recording software on the host machine and records the user's session.

**Mitigation**: You cannot prevent a host from looking at their own monitor (screen duplication inherently means the host's display shows the session). Be transparent about this in the ToS. For sensitive work (passwords, banking), advise users not to use the platform. This mirrors using a public computer at an internet café — known, disclosed risk.

### 8.2 Security Layers Summary

| Layer | Technology | Purpose |
|---|---|---|
| VM Isolation | Hyper-V (Type 1) / KVM | Prevent sandbox escape to host OS |
| Network Isolation | Virtual NAT adapter, no LAN route | Prevent lateral network attacks |
| Storage Isolation | Ephemeral disk (wiped post-session) | No persistent malware or data on host |
| Process Isolation | Strict AppContainer / seccomp policies | Limit dangerous syscalls |
| Resource Limits | vCPU/RAM/Disk quotas | Prevent DoS against host |
| Encrypted Transport | TLS 1.3 on all streams | Protect streaming data in transit |
| Session Audit | Timestamped session logs | Non-repudiation, dispute resolution |

---

## 9. Streaming & Networking Layer

### 9.1 Codec Selection

For interactive gaming and desktop sessions, the codec pipeline is critical. Based on current research and production deployments:

**Video Encoding (on host GPU/CPU):**
- H.264: Universal compatibility, well-supported hardware encoders (NVENC, AMD VCE). Good up to 1080p. Fastest encoding latency. Best for MVP.
- H.265 (HEVC): 40–50% better compression than H.264 at same quality. Important for 4K. Broad hardware support now. Recommended for production.
- AV1: Best compression, royalty-free. AV1 hardware encoding (RTX 4000 series, RX 7000 series). Future standard. Add in Phase 3.

**Transport Protocol:**

Research published in 2025 confirms QUIC-based protocols (Media over QUIC / RTP over QUIC) outperform WebRTC by approximately 30% in latency and 60% in connection startup times. WebRTC is more mature in implementation, but QUIC is clearly the direction.

**Recommended approach:**
- **Phase 1/2**: WebRTC (battle-tested, works in browsers, mature SDKs)
- **Phase 3**: Migrate to QUIC/MoQ as QUIC libraries mature (Cloudflare launched a public MoQ relay network in 2025 — this infrastructure can be leveraged)

**Signaling:**
- Host agent registers with signaling server on startup
- User client sends session request → signaling server brokers STUN/ICE exchange
- Direct peer-to-peer connection established (UDP-based)
- TURN relay server for cases where P2P fails (symmetric NAT situations)

### 9.2 Input Forwarding

Input (keyboard, mouse, gamepad) needs to be captured client-side and forwarded to the host with minimum latency. The host agent injects this input into the sandbox via a virtual HID (Human Interface Device) driver.

**Approach:**
- Client: capture raw HID events (scancode level, not OS-abstracted) at polling rate of 125–1000Hz
- Transmit as low-priority QUIC stream (unreliable datagrams — dropping an old input event is better than queuing it)
- Host agent: virtual HID driver (Windows ViGEm Bus or similar, Linux uinput) injects events into sandbox

**Latency target**: < 5ms input-to-injection latency on the host side. The network round-trip is unavoidable.

---

## 10. Business Model & Pricing

### 10.1 Revenue Model

**Platform Commission**: Take 15–20% of every transaction. This is the standard marketplace take rate (Vast.ai takes 10–15%, most platforms 15–25%).

**Premium Listings**: Hosts can pay to appear higher in search results.

**Harness Pro (User Subscription)**: Monthly subscribers get priority booking, lower fees, guaranteed session start times.

**Enterprise / Team Plans**: Studios, AI labs, and small businesses can access dedicated compute pools with SLA guarantees.

### 10.2 Pricing Examples

Based on current GPU rental market rates and the peer-to-peer discount model:

| Use Case | Recommended Host Rate | Harness Platform Fee | User Pays |
|---|---|---|---|
| Light gaming (GTX 1080) | $0.30/hr | $0.06/hr | $0.36/hr |
| 1080p gaming (RTX 3070) | $0.60/hr | $0.12/hr | $0.72/hr |
| 4K gaming (RTX 4090) | $1.00/hr | $0.20/hr | $1.20/hr |
| AI/ML workloads (A100-class) | $1.50/hr | $0.30/hr | $1.80/hr |
| Video editing (RTX 4080) | $0.80/hr | $0.16/hr | $0.96/hr |

For comparison: Shadow PC charges $30/month flat (~$1/hr assuming 30hrs use), GeForce NOW charges $20/month. Harness at $0.72–$1.20/hr for better hardware is highly competitive.

### 10.3 Host Earnings Estimate

A host with an RTX 4090 running 6 hours/night at $1.00/hr earns:
- $6/night × 25 active days = **$150/month** passive income
- Annual: **$1,800** from a GPU they already own

This is a compelling value proposition. For comparison, crypto mining on an RTX 4090 in 2025 earns $2–$4/day. Harness could earn more without the electricity overhead of proof-of-work mining.

---

## 11. Legal & Compliance Risks

### 11.1 Key Legal Issues

**Game License Agreements**: Most AAA game EULAs prohibit running the game on a commercial service or renting access to it. If a host installs GTA V and a user pays to play it via Harness, Rockstar/Take-Two may argue this violates their EULA.

**Mitigation**: This is a grey area that even cloud gaming providers (GeForce NOW, Shadow) have navigated. The legal argument is that the host owns a legitimate copy and is granting remote access to their own licensed machine — similar to Steam Remote Play. Monitor legal precedent closely. Consult a games IP attorney early.

**DMCA / Copyright**: Users might use compute to pirate software, transcode copyrighted films, etc.

**Mitigation**: DMCA safe harbor (for US) applies if you have a compliant notice-and-takedown procedure. Implement this from day one.

**Data Protection (GDPR, etc.)**: User files and session data have privacy implications.

**Mitigation**: Ephemeral sandboxes help. Minimum data retention. Clear privacy policy. GDPR compliance for European users.

**Tax and Payment Compliance**: Marketplace income requires proper tax handling in each jurisdiction.

**Mitigation**: Stripe's Connect product handles marketplace payouts and tax reporting (1099s in the US, etc.) automatically.

### 11.2 ISP Policy

As noted in Section 7.4, residential ISP ToS may be an issue. The platform should:
- Not explicitly advertise itself as a "server hosting" or "commercial hosting" service
- Frame it as "resource sharing" (analogous to Airbnb framing home rental as "sharing")
- Recommend that hosts with business internet plans use them

---

## 12. Tech Stack Recommendation

### Backend (Orchestration, Marketplace, Billing)

| Component | Technology | Rationale |
|---|---|---|
| API Server | Go (Golang) | High concurrency, fast, great for network services |
| Database | PostgreSQL | Reliable, relational, great for billing/sessions |
| Cache / Session State | Redis | In-memory, fast pub/sub for session signaling |
| Message Queue | NATS or RabbitMQ | Host ↔ backend event bus |
| File Storage | Cloudflare R2 | No egress fees, S3-compatible, cheap |
| Signaling Server | Custom WebSocket server (Go) | Broker ICE/STUN handshakes |
| TURN Server | coturn (open source) | Relay fallback for NAT scenarios |
| Payment | Stripe (Connect for marketplace) | Industry standard, handles payouts |
| Containerization | Docker + Kubernetes | Backend microservices |
| CDN | Cloudflare | DDoS protection, global edge |

### Frontend (Web + Desktop App)

| Component | Technology | Rationale |
|---|---|---|
| Web Marketplace | Next.js (React) | SSR for SEO, great ecosystem |
| Desktop App | Electron (Chromium shell) OR Tauri (Rust + webview) | Cross-platform; Tauri is lighter, more modern |
| Streaming Renderer | WebCodecs API + WebRTC / WebTransport | Hardware-accelerated decode in browser |
| UI Library | Tailwind CSS + shadcn/ui | Fast to build, clean design |

### Host Agent (The Core)

| Component | Technology | Rationale |
|---|---|---|
| Agent Core | C++ (primary) or Rust (if new team) | Low-level system access, codec libraries |
| Screen Capture | Windows: DXGI Desktop Duplication API | Zero-copy GPU capture |
| Screen Capture | Linux: PipeWire / KMS/DRM | Modern Linux display capture |
| Video Encoding | NVENC SDK (NVIDIA), AMF SDK (AMD), libva (Intel) | Hardware GPU encoding for low latency |
| Codec Container | FFmpeg / libavcodec | Wraps codec SDKs, mux/demux |
| Streaming Transport | libwebrtc (Google's C++ library) or quiche (QUIC) | P2P streaming |
| Virtual Input | Windows: ViGEm Bus Driver, SendInput | Inject keyboard/mouse/gamepad |
| Virtual Input | Linux: uinput kernel module | Inject input events |
| Sandbox Management | Windows Sandbox API / Hyper-V PowerShell API / libvirt (Linux KVM) | VM lifecycle control |
| IPC | Named pipes (Windows) / Unix sockets (Linux) | Agent ↔ sandbox communication |

### Client App (Thin Client — What Users Install)

- **Primary**: Electron/Tauri app with WebRTC renderer
- **Fallback**: Browser-based (WebRTC + WebCodecs, no install required)
- **Mobile**: React Native or Flutter with custom native video decoder

---

## 13. Development Roadmap (Phased)

### Phase 0 — Research & Architecture (Month 1–2)

Validate all assumptions before writing production code.

- Build a raw prototype: screen capture on one PC, encode to H.264, stream over WebRTC to another PC on LAN. Measure latency.
- Test Windows Sandbox with a game — does GPU sharing work? What's the performance hit?
- Set up a simple Stripe test integration — can you gate access to the stream behind a payment?
- Survey 20–30 potential hosts (gamers with high-end PCs) and 20–30 potential users about pricing, trust, and use cases.
- Hire (or decide to be) a C++ systems programmer comfortable with Windows internals.

**Deliverable**: 1-page architecture decision record. Proof-of-concept demo video.

### Phase 1 — MVP: Private Alpha (Month 3–6)

Build the minimal product that proves the core concept works end-to-end. Manual, curated, invite-only.

**Host Agent (Windows only):**
- Install wizard
- Screen capture (DXGI) + H.264 encoding (NVENC)
- WebRTC stream to user
- Input injection (keyboard + mouse, no gamepad yet)
- Application-level sandbox (Windows Sandbox)
- Simple IPC to receive start/stop commands from backend

**Backend:**
- Host registration API
- Session creation API
- Simple Stripe payment (no marketplace yet — fixed hourly price)
- Session timer + auto-termination on balance depletion
- Session logging

**User Client:**
- Electron app (Windows first)
- WebRTC decoder + render
- Keyboard/mouse capture and forwarding
- Basic UI: "Connect to session" button

**Exclusions for MVP**: No marketplace UI, no file transfer, no mobile support, no gamepad, no GPU passthrough VM.

**Success Metrics**: 10 host–user sessions with < 50ms end-to-end latency on LAN. Payment flows correctly. Sandbox holds (user cannot access host files).

### Phase 2 — Beta: Marketplace Launch (Month 7–12)

Open the platform to a wider audience with real marketplace features.

**Marketplace:**
- Host listing creation (set price, describe hardware, set availability hours)
- Browse/search/filter by region, GPU, price, use case, rating
- User reviews and host ratings
- Stripe Connect integration for host payouts

**Improved Streaming:**
- H.265 support
- Adaptive bitrate (auto-downscale if bandwidth is insufficient)
- Gamepad/controller support (ViGEm passthrough)
- Audio streaming (sandwich into WebRTC audio track)

**File Transfer:**
- Pre-session file upload (Harness staging area)
- Post-session output download
- In-session clipboard sharing (text only for security)

**Security Hardening:**
- Strict network isolation (sandbox behind NAT, no LAN access)
- Session monitoring (GPU utilization anomaly detection)
- User KYC for hosts (email verification + payment method required)

**Platform Expansion:**
- Mac client (for users — remote in to Windows hosts)
- Browser-based thin client (no install for users)
- Linux host agent (KVM/QEMU sandbox)

**Success Metrics**: 100 hosts registered. 500 paid sessions. Average session NPS > 7. Zero successful sandbox escapes.

### Phase 3 — Growth: Advanced Features (Month 13–24)

**Technical Advances:**
- GPU Passthrough (Hyper-V DDA on Windows, VFIO on Linux) with guided setup wizard
- AV1 codec support (RTX 4000 / RX 7000 hosts unlock "AV1 listing" badge)
- QUIC/MoQ transport replacing WebRTC for further latency reduction
- Multi-monitor support
- Mobile clients (Android + iOS)

**Platform Features:**
- Harness Pro subscription (lower fees, priority matching)
- Enterprise/Team accounts with dedicated pools
- Scheduled reservations ("book this host every night 9 PM–1 AM")
- Host software pre-configuration (OS image with games/software pre-installed)
- Public API for third-party integrations

**Business Development:**
- Partner with PC hardware brands (ASUS ROG, MSI Gaming) to pre-install agent
- Partner with ISPs for promoted plans optimized for hosting
- Explore integration with game launchers (Steam, Epic) for library verification

**Success Metrics**: $100K+ MRR. 1,000+ active hosts. 5,000+ monthly active users. Geographic presence in at least 5 regions.

---

## 14. Final Verdict & Recommendation

### Is this possible? Yes.

Every component exists. The streaming tech (WebRTC, QUIC), the sandboxing tech (Hyper-V, KVM), the payment infrastructure (Stripe) — none of this needs to be invented. The engineering is hard but tractable.

### Has it been done before? Partially, and not well.

The Parsec Cloud Gaming Marketplace was the closest thing and it was shut down — but for business reasons (wrong supply model), not technical ones. The GPU AI compute marketplaces (Vast.ai, RunPod) prove the P2P model works commercially. Nobody has merged interactive GUI sessions with P2P personal hardware supply. That gap is real.

### What are the make-or-break challenges?

1. **GPU Passthrough UX**: Making sandbox + GPU passthrough easy enough for non-technical hosts is the hardest UX challenge. Phase 1 sidesteps this with application-level sandboxing — acceptable for MVP.

2. **Host Reliability**: Users will abandon the platform after two or three dropped sessions. The host incentive and reputation system must be excellent from day one.

3. **Latency on Real Networks**: The product will be amazing on local networks (LAN), good on fiber (< 30ms RTT), and mediocre on typical home broadband (20–50ms RTT). Be honest about use cases — gaming needs < 30ms, video editing tolerates 80ms, AI/ML batch jobs tolerate any latency.

4. **Legal (Game EULAs)**: Get legal advice before launch. Frame carefully.

### What is the right first step?

Build the proof of concept described in Phase 0. Specifically: one Windows PC streaming a game over WebRTC to another PC, gated behind a $1 Stripe payment, with the game running inside Windows Sandbox. If you can do that in a weekend, you have proven the core hypothesis and can raise a pre-seed round or start building Phase 1.

The idea is sound. The market is real. The tech is ready. **Build it.**

---

*Report compiled using data from Aethir, Hyperbolic, Vast.ai, RunPod, SaladCloud, MicroDC, Parsec, Shadow PC, AirGPU, Fluence Network, and research from IEEE (BMSB 2025), Cloudflare (MoQ 2025), and Microsoft documentation.*
