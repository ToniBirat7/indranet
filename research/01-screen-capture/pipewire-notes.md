# PipeWire / KMS Screen Capture (Linux)

## Relevance
Linux host support is deferred to Phase 3. These notes are for future reference.

## Question
How do we capture the screen on Linux with minimal latency?

## Options

1. **PipeWire ScreenCast portal** (Wayland + X11): Modern, works with compositors that support `xdg-desktop-portal`. Latency unknown.
2. **KMS/DRM direct**: Read from DRM framebuffer directly. Low latency but requires root and compositor cooperation.
3. **X11 SHM (XShmGetImage)**: Available on X11 sessions. Higher latency than KMS.

## TODO
Research when Phase 3 Linux host support begins.
