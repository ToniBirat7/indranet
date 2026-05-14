# IndraNet — Immediate TODO

## Phase 0 — Research & Validation
**Deadline goal: Prove the 4 PoCs work within 6 weeks**

---

### @researcher

- [ ] Complete `research/01-screen-capture/dxgi-notes.md` — read MSDN Desktop Duplication API docs, summarize quirks (monitor count, HDR, protected content handling)
- [ ] Complete `research/02-video-encoding/nvenc-notes.md` — document NVENC SDK initialization, B-frame support, low-latency mode settings (NVENC_ENC_PARAMS_RC_CBR_LOWDELAY_HQ)
- [ ] Complete `research/03-streaming-transport/webrtc-notes.md` — evaluate libwebrtc vs GStreamer WebRTC for our use case (complexity, latency, platform support)
- [ ] Complete `research/04-sandboxing/windows-sandbox.md` — test GPU access behavior inside Windows Sandbox; document which DirectX feature levels are available
- [ ] Complete `research/05-input-forwarding/vigem-notes.md` — document ViGEm Bus driver installation, virtual controller types (X360 vs DS4), latency characteristics
- [ ] Complete `research/06-payment/stripe-connect.md` — document Connect onboarding flow (Express accounts), payout timing, platform fee mechanics, webhook event list

---

### @poc-builder

**PoC 01 — Screen Capture + Encode**
- [ ] Get DXGI Desktop Duplication running on a test machine — initialize IDXGIOutputDuplication
- [ ] Add NVENC encode — measure encode latency at 1080p60 using NVENC low-latency profile
- [ ] Record benchmark results in `poc/01-screen-capture-encode/README.md`
- [ ] Success criteria: <5ms sustained capture+encode at 60fps

**PoC 02 — WebRTC Stream**
- [ ] Get GStreamer WebRTC pipeline running: `nvenc → appsink → webrtcbin`
- [ ] Connect to Go signaling server with offer/answer exchange
- [ ] Measure end-to-end latency on LAN (capture → display) using frame timestamp watermark
- [ ] Record benchmark results in `poc/02-webrtc-stream/README.md`
- [ ] Success criteria: <50ms end-to-end on LAN

**PoC 03 — Sandbox GPU Access**
- [ ] Run `launch-sandbox.ps1` on a Windows machine with a GPU
- [ ] Inside sandbox: run `dxdiag` or `DirectX Caps Viewer` — confirm GPU device listed
- [ ] Run a simple DirectX/OpenGL benchmark inside sandbox — compare FPS to bare metal
- [ ] Record findings in `poc/03-sandbox-test/README.md`
- [ ] Success criteria: GPU-accelerated rendering confirmed inside sandbox

**PoC 04 — Payment Gate**
- [ ] Configure Stripe test keys in `.env`
- [ ] Run `poc/04-payment-gate/server/main.go` locally
- [ ] Open `poc/04-payment-gate/client/index.html` — complete test payment
- [ ] Verify JWT received and stream token valid
- [ ] Record results in `poc/04-payment-gate/README.md`
- [ ] Success criteria: Pay $1 → receive JWT → connect to mock stream

---

### @backend-engineer

- [ ] Initialize Go module: `go mod init github.com/ToniBirat7/indranet/packages/backend`
- [ ] Implement `GET /health` endpoint — DB ping + Redis ping + JSON status
- [ ] Write `001_init.sql` — users, hosts, sessions, payments tables with proper indices
- [ ] Implement host registration (`POST /hosts/register`) with machine fingerprint hashing
- [ ] Set up `golangci-lint` config (`.golangci.yml`) with: errcheck, gosimple, govet, staticcheck
- [ ] Write billing engine unit tests in `packages/backend/tests/billing_test.go`

---

### @devops

- [ ] Verify `docker-compose up` starts all services with health checks passing
- [ ] Create a test PR on a feature branch to verify `.github/workflows/ci.yml` runs
- [ ] Add pre-commit hook config (`.pre-commit-config.yaml`): go fmt, goimports, pnpm lint
- [ ] Verify `scripts/setup.sh` works on a clean Ubuntu 22.04 environment

---

### @architect

- [ ] Finalize `docs/specs/api-spec.md` — document all REST endpoints with request/response schemas
- [ ] Write `docs/architecture/diagrams/session-lifecycle.md` — full state machine with ASCII diagram
- [ ] Review and finalize `research/04-sandboxing/security-model.md` threat model before any sandbox code ships
