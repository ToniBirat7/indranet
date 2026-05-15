#include <iostream>

// DXGI Desktop Duplication API screen capture
// Docs: https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api
//
// Performance target: capture latency < 1ms per frame at 60fps
// The captured ID3D11Texture2D is passed directly to NVENC for zero-copy GPU encoding.

#ifdef _WIN32
#include <d3d11.h>
#include <dxgi1_2.h>

struct CaptureFrame {
    // Pointer to GPU texture — do NOT copy to CPU; pass directly to encoder
    void* texture = nullptr;
    long long captureTimestampUs = 0;
    int width = 0;
    int height = 0;
};

class DXGICapture {
public:
    bool init(int adapterIndex = 0, int outputIndex = 0) {
        // TODO: Create D3D11 device
        // TODO: Get IDXGIAdapter, IDXGIOutput, IDXGIOutput1
        // TODO: Create IDXGIOutputDuplication via DuplicateOutput()
        std::cout << "[capture] DXGI init (stub)\n";
        return false;
    }

    // AcquireFrame blocks until a new frame is available or timeout elapses.
    // Returns false on timeout (no new frame) or fatal error.
    bool acquireFrame(CaptureFrame& frame, int timeoutMs = 100) {
        // TODO: Call AcquireNextFrame()
        // TODO: QueryInterface to get ID3D11Texture2D
        // TODO: Record timestamp with std::chrono
        return false;
    }

    void releaseFrame() {
        // TODO: Call ReleaseFrame() on IDXGIOutputDuplication
    }

    void shutdown() {
        // TODO: Release all COM objects in reverse order
    }

private:
    // TODO: ID3D11Device, IDXGIOutputDuplication members
};
#endif // _WIN32
