# Host Agent Specification

The host agent is a C++ daemon that runs on the host machine and manages the entire session lifecycle from the host side.

## Startup & Registration

1. Parse CLI arguments: `--backend-url`, `--session-token`, `--sandbox-type`
2. Verify network connectivity to backend URL
3. Authenticate with backend using session token (JWT)
4. Retrieve session configuration: sandbox type, resource limits, game path
5. Report machine capabilities: GPU model, VRAM, CPU, RAM, display resolution
6. Register as ready: `POST /sessions/:id/agent-ready`

## Capture Loop

The capture loop runs on a dedicated thread with highest-available OS priority.

1. Initialize DXGI Output Duplication on the primary display (or the display assigned to the sandbox)
2. Each frame iteration:
   a. `AcquireNextFrame(timeout=16ms)` — blocks until new frame or timeout
   b. On `DXGI_ERROR_WAIT_TIMEOUT`: no new frame, skip encode, continue loop
   c. On `DXGI_ERROR_ACCESS_LOST`: re-initialize DXGI (monitor change, sleep/wake)
   d. On success: get `IDXGISurface` pointer → pass to encoder
3. Frame timestamp recorded at acquisition for latency measurement
4. Target: 60 frames per second, <1ms per capture call

## Encode Pipeline

The encode pipeline runs on the capture thread (to avoid GPU-to-CPU-to-GPU round-trip).

1. NVENC session initialized with:
   - Codec: H.264
   - Rate control: `NV_ENC_PARAMS_RC_CBR_LOWDELAY_HQ`
   - B-frames: 0
   - Keyframe interval: 120 frames (2 seconds at 60fps)
   - Bitrate: configurable, default 15 Mbps
2. On each captured frame:
   a. Map DXGI surface to NVENC input buffer (zero-copy on same GPU)
   b. `NvEncEncodePicture()` → blocks until encode completes
   c. Get encoded NAL units from output buffer
   d. Push NAL units to streaming queue
3. Dynamic bitrate adjustment: respond to `BITRATE_CHANGE` events from stream thread

## Stream Transmission

The stream thread pulls from the encode queue and sends via WebRTC.

1. Initialize GStreamer WebRTC pipeline (or libwebrtc)
2. Connect to backend signaling WebSocket
3. Create peer connection, generate SDP offer
4. Send offer → backend relays to user client
5. Receive answer → apply remote description
6. ICE candidate exchange via backend relay
7. On `ICE_CONNECTED`: begin feeding encoded frames to WebRTC sender
8. Audio: WASAPI loopback capture → Opus encode → WebRTC audio track
9. Data channel: open bidirectional `input` channel for HID events

## Input Reception

Input events arrive on the WebRTC data channel.

1. On data channel message: deserialize HID packet
2. Route by event type:
   - Keyboard: `INPUT` struct via `SendInput()`, target HWND inside sandbox
   - Mouse: `INPUT` struct, relative movement only
   - Gamepad: ViGEm virtual Xbox controller state update
3. Target: <5ms from data channel receive to `SendInput()` call

**Critical:** Input must be injected into the sandbox application window, not the host desktop. This requires finding the correct HWND inside the sandbox process tree.

## Sandbox Lifecycle

1. On session AUTHORIZED:
   a. Create `.wsb` config file (mapped drives, networking policy, GPU access)
   b. Start Windows Sandbox process
   c. Wait for sandbox desktop to be responsive (poll via WMI or named pipe)
   d. Launch target application inside sandbox
   e. Notify backend: `session_ready`

2. On session ENDING (kill signal received):
   a. Send graceful terminate to application inside sandbox
   b. Wait up to 5 seconds for clean exit
   c. Force-terminate sandbox process if not exited
   d. Verify sandbox process is gone (no orphaned processes)
   e. Delete temporary `.wsb` file and any temp assets
   f. Notify backend: `sandbox_destroyed`

3. On unexpected sandbox exit (crash):
   a. Detect via process handle monitoring
   b. Emit `session_failed` to backend
   c. Clean up as in step 2

## Billing Heartbeat

1. Every 30 seconds: send heartbeat to backend (`PUT /sessions/:id/heartbeat`)
2. If backend returns `KILL` in heartbeat response: begin session teardown
3. If heartbeat fails 3 times consecutively: log error, continue session (network hiccup tolerance)
4. If heartbeat fails 10 times consecutively: treat as backend disconnect, end session gracefully

## Graceful Shutdown

On `SIGTERM` or `SIGINT`:
1. Stop new frame captures
2. Drain encode queue (max 500ms)
3. Close WebRTC peer connection
4. Terminate sandbox (as in ENDING procedure)
5. Flush all logs
6. Exit with code 0

On unhandled exception or assertion failure:
1. Attempt sandbox teardown (critical — must not leak sandboxes)
2. Write crash dump to `%TEMP%\indranet-crash-{timestamp}.dmp`
3. Exit with non-zero code

## IPC with Tauri Client

The Tauri desktop client communicates with the agent via a named pipe (Windows) or Unix domain socket (Linux).

Pipe name: `\\.\pipe\IndraNetAgent` (Windows)

Messages (JSON lines):
- Client → Agent: `{"cmd": "start", "session_token": "...", "backend_url": "..."}`
- Client → Agent: `{"cmd": "stop"}`
- Agent → Client: `{"event": "status", "state": "active", "fps": 60, "bitrate_mbps": 15.2}`
- Agent → Client: `{"event": "error", "code": "SANDBOX_LAUNCH_FAILED", "message": "..."}`
