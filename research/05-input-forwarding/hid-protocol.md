# HID Event Protocol for WebRTC Data Channel

## Design Goals
- Minimal serialization overhead (<100 bytes per event)
- Typed (distinguishes keyboard, mouse, gamepad)
- Language-agnostic (TypeScript client, C++ agent)

## Event Format (JSON over WebRTC data channel)

### Keyboard
```json
{ "t": "k", "e": "d", "c": 65, "k": 65 }
```
- `t`: type = "k" (keyboard)
- `e`: event = "d" (keydown) or "u" (keyup)
- `c`: scancode
- `k`: virtual key code

### Mouse Move (relative)
```json
{ "t": "m", "dx": 5, "dy": -3 }
```

### Mouse Button
```json
{ "t": "mb", "e": "d", "b": 0 }
```
- `b`: button index (0=left, 1=middle, 2=right)

### Mouse Scroll
```json
{ "t": "ms", "dx": 0, "dy": -1 }
```

### Gamepad State (full report, 60Hz)
```json
{ "t": "gp", "lx": 0.5, "ly": -0.2, "rx": 0.0, "ry": 0.0, "lt": 0.0, "rt": 1.0, "b": 4096 }
```
- `b`: bitmask of buttons (Xbox 360 XUSB_REPORT format)

## Phase 2: Binary Format
JSON adds ~10-20 bytes overhead per event. Phase 2 should use a compact binary format (e.g., FlatBuffers or hand-crafted struct) for lower latency.
