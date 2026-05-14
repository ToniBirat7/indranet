#include <windows.h>
#include <d3d11.h>
#include <dxgi1_2.h>
#include <wrl/client.h>
#include <cstdio>
#include <vector>
#include <chrono>
#include <algorithm>
#include <numeric>

// TODO: Include NVENC SDK headers once NVENC_SDK_PATH is set
// #include <nvEncodeAPI.h>

using Microsoft::WRL::ComPtr;
using Clock = std::chrono::high_resolution_clock;

// Measurement over N frames
static const int FRAME_COUNT = 1000;

struct FrameTiming {
    double capture_ms;
    double encode_ms;
};

// TODO: Initialize NVENC encoder
// Returns true on success. Currently a stub.
bool init_nvenc(ID3D11Device* device, int width, int height) {
    printf("[NVENC] TODO: Initialize NVENC encoder\n");
    printf("[NVENC]   - Load nvEncodeAPI.dll\n");
    printf("[NVENC]   - Call NvEncOpenEncodeSessionEx()\n");
    printf("[NVENC]   - Set preset: NV_ENC_PRESET_P1_GUID\n");
    printf("[NVENC]   - Set rate control: NV_ENC_PARAMS_RC_CBR_LOWDELAY_HQ\n");
    printf("[NVENC]   - Set bitrate: 15 Mbps\n");
    printf("[NVENC]   - Set B-frames: 0\n");
    printf("[NVENC]   - Set keyframe interval: 120 (2s at 60fps)\n");
    return false; // Remove when implemented
}

// TODO: Encode one frame with NVENC
// Returns encoded bytes or -1 on error. Currently a stub that simulates ~2ms encode.
int encode_frame(ID3D11Texture2D* texture) {
    // TODO: Map texture to NVENC input buffer (zero-copy D3D11 path)
    // TODO: Call NvEncEncodePicture()
    // TODO: Get output bitstream

    // Simulate encode latency for timing framework testing
    Sleep(2);
    return 1024; // Simulated encoded bytes
}

int main() {
    printf("=== IndraNet PoC 01: Screen Capture + NVENC Encode ===\n\n");

    // ─── Step 1: Create D3D11 device ─────────────────────────────────────────
    ComPtr<ID3D11Device> device;
    ComPtr<ID3D11DeviceContext> context;
    D3D_FEATURE_LEVEL feature_level;

    HRESULT hr = D3D11CreateDevice(
        nullptr,                    // Default adapter
        D3D_DRIVER_TYPE_HARDWARE,
        nullptr,                    // No software rasterizer
        0,
        nullptr, 0,                 // Default feature levels
        D3D11_SDK_VERSION,
        &device, &feature_level, &context
    );
    if (FAILED(hr)) {
        printf("[ERROR] D3D11CreateDevice failed: 0x%08X\n", hr);
        return 1;
    }
    printf("[D3D11] Device created (feature level: 0x%04X)\n", feature_level);

    // ─── Step 2: Get DXGI output for the primary monitor ─────────────────────
    ComPtr<IDXGIDevice> dxgi_device;
    device->QueryInterface(IID_PPV_ARGS(&dxgi_device));

    ComPtr<IDXGIAdapter> adapter;
    dxgi_device->GetAdapter(&adapter);

    // TODO: Allow selecting specific monitor via CLI arg
    ComPtr<IDXGIOutput> output;
    hr = adapter->EnumOutputs(0, &output);
    if (FAILED(hr)) {
        printf("[ERROR] No display output found: 0x%08X\n", hr);
        return 1;
    }

    DXGI_OUTPUT_DESC output_desc;
    output->GetDesc(&output_desc);
    int width  = output_desc.DesktopCoordinates.right - output_desc.DesktopCoordinates.left;
    int height = output_desc.DesktopCoordinates.bottom - output_desc.DesktopCoordinates.top;
    printf("[DXGI] Output: %dx%d\n", width, height);

    // ─── Step 3: Create Desktop Duplication ──────────────────────────────────
    ComPtr<IDXGIOutput1> output1;
    output->QueryInterface(IID_PPV_ARGS(&output1));

    ComPtr<IDXGIOutputDuplication> duplication;
    hr = output1->DuplicateOutput(device.Get(), &duplication);
    if (FAILED(hr)) {
        printf("[ERROR] DuplicateOutput failed: 0x%08X\n", hr);
        printf("[HINT]  Try running as administrator, or check if another app owns the desktop.\n");
        return 1;
    }
    printf("[DXGI] Desktop Duplication initialized\n");

    // ─── Step 4: Initialize NVENC ────────────────────────────────────────────
    bool nvenc_ok = init_nvenc(device.Get(), width, height);
    if (!nvenc_ok) {
        printf("[WARN] NVENC not initialized — running capture-only timing test\n");
    }

    // ─── Step 5: Capture loop ─────────────────────────────────────────────────
    printf("\n[CAPTURE] Starting %d-frame timing test...\n\n", FRAME_COUNT);

    std::vector<FrameTiming> timings;
    timings.reserve(FRAME_COUNT);

    int captured = 0;
    int timeouts = 0;

    while (captured < FRAME_COUNT) {
        auto t_capture_start = Clock::now();

        // Acquire next frame — timeout 16ms (allows 60fps with some slack)
        DXGI_OUTDUPL_FRAME_INFO frame_info = {};
        ComPtr<IDXGIResource> desktop_resource;

        hr = duplication->AcquireNextFrame(16, &frame_info, &desktop_resource);

        if (hr == DXGI_ERROR_WAIT_TIMEOUT) {
            timeouts++;
            continue; // No new frame, try again
        }
        if (hr == DXGI_ERROR_ACCESS_LOST) {
            // TODO: Re-initialize duplication (monitor change, DWM restart, etc.)
            printf("[WARN] DXGI_ERROR_ACCESS_LOST — need to re-initialize\n");
            break;
        }
        if (FAILED(hr)) {
            printf("[ERROR] AcquireNextFrame failed: 0x%08X\n", hr);
            break;
        }

        auto t_capture_end = Clock::now();

        // Get D3D11 texture from DXGI resource
        ComPtr<ID3D11Texture2D> texture;
        desktop_resource->QueryInterface(IID_PPV_ARGS(&texture));

        // TODO: Pass texture to NVENC zero-copy input buffer
        auto t_encode_start = Clock::now();
        encode_frame(texture.Get());
        auto t_encode_end = Clock::now();

        // PERF: Critical path — record per-frame timing
        timings.push_back({
            std::chrono::duration<double, std::milli>(t_capture_end - t_capture_start).count(),
            std::chrono::duration<double, std::milli>(t_encode_end - t_encode_start).count()
        });

        duplication->ReleaseFrame();
        captured++;

        if (captured % 100 == 0) {
            printf("[CAPTURE] Frame %d/%d\n", captured, FRAME_COUNT);
        }
    }

    printf("\n[CAPTURE] Done. Captured %d frames, %d timeouts\n\n", captured, timeouts);

    // ─── Step 6: Print timing statistics ─────────────────────────────────────
    if (timings.empty()) {
        printf("[ERROR] No timing data collected\n");
        return 1;
    }

    auto percentile = [&](std::vector<double>& v, double p) {
        std::sort(v.begin(), v.end());
        return v[static_cast<size_t>(v.size() * p / 100.0)];
    };

    std::vector<double> capture_ms, encode_ms, combined_ms;
    for (auto& t : timings) {
        capture_ms.push_back(t.capture_ms);
        encode_ms.push_back(t.encode_ms);
        combined_ms.push_back(t.capture_ms + t.encode_ms);
    }

    printf("=== RESULTS ===\n");
    printf("\nCapture latency (ms):\n");
    printf("  p50:  %.2f\n", percentile(capture_ms, 50));
    printf("  p95:  %.2f\n", percentile(capture_ms, 95));
    printf("  p99:  %.2f\n", percentile(capture_ms, 99));
    printf("  max:  %.2f\n", *std::max_element(capture_ms.begin(), capture_ms.end()));

    printf("\nEncode latency (ms) [%s]:\n", nvenc_ok ? "NVENC" : "SIMULATED");
    printf("  p50:  %.2f\n", percentile(encode_ms, 50));
    printf("  p95:  %.2f\n", percentile(encode_ms, 95));
    printf("  p99:  %.2f\n", percentile(encode_ms, 99));

    printf("\nCombined capture+encode (ms):\n");
    printf("  p50:  %.2f\n", percentile(combined_ms, 50));
    printf("  p95:  %.2f\n", percentile(combined_ms, 95));
    printf("  p99:  %.2f\n", percentile(combined_ms, 99));

    printf("\n=== SUCCESS CRITERIA ===\n");
    double p95_combined = percentile(combined_ms, 95);
    printf("  p95 combined < 5ms: %s (actual: %.2fms)\n",
        p95_combined < 5.0 ? "PASS" : "FAIL",
        p95_combined);

    return 0;
}
