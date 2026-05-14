# Screen Capture Benchmarks

## Methodology

Capture 1000 frames, record per-frame timing for each stage.

```
t1 = before AcquireNextFrame()
t2 = after AcquireNextFrame() returns
capture_latency = t2 - t1
```

Report p50, p95, p99, max.

## Results

TODO: Fill in after running poc/01-screen-capture-encode.

| Configuration | p50 | p95 | p99 | Max |
|---------------|-----|-----|-----|-----|
| 1080p60, bare metal | - | - | - | - |
| 1080p60, inside sandbox | - | - | - | - |
| 4K60, bare metal | - | - | - | - |
