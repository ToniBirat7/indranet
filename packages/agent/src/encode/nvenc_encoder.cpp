#include <iostream>

// NVIDIA NVENC hardware video encoder
// SDK: https://developer.nvidia.com/nvidia-video-codec-sdk
//
// Performance target: encode latency < 2ms per frame at 1080p60
// Accepts ID3D11Texture2D directly from DXGI (zero GPU-CPU copy).

#ifdef HAVE_NVENC
#include <nvEncodeAPI.h>

class NVENCEncoder {
public:
    struct Config {
        int width = 1920;
        int height = 1080;
        int fps = 60;
        int bitrateMbps = 20;
        // H.264 for Phase 1; Phase 2 adds H.265/HEVC
        NV_ENC_CODEC codec = NV_ENC_CODEC_H264_GUID;
        // Low latency preset — maximize for interactive use, not VOD quality
        NV_ENC_PRESET preset = NV_ENC_PRESET_P1_GUID; // fastest
    };

    bool init(void* d3d11Device, const Config& cfg) {
        // TODO: Load nvEncodeAPI64.dll / nvEncodeAPI.so
        // TODO: NvEncOpenEncodeSessionEx with D3D11 device
        // TODO: NvEncGetEncodePresetConfigEx
        // TODO: NvEncInitializeEncoder with low-latency params:
        //   - enablePTD = 1 (picture-type decision)
        //   - repeatSPSPPS = 1 (for stream recovery)
        //   - idrPeriod = fps * 2 (keyframe every 2s)
        //   - rateControlMode = NV_ENC_PARAMS_RC_CBR_LOWDELAY_HQ
        std::cout << "[encode] NVENC init (stub)\n";
        return false;
    }

    // Encodes a single frame. Output NAL units are appended to outBitstream.
    bool encodeFrame(void* d3d11Texture, std::vector<uint8_t>& outBitstream) {
        // TODO: NvEncMapInputResource
        // TODO: NvEncEncodePicture
        // TODO: NvEncLockBitstream → copy NAL data
        // TODO: NvEncUnlockBitstream, NvEncUnmapInputResource
        return false;
    }

    void shutdown() {
        // TODO: NvEncDestroyEncoder
    }
};
#endif // HAVE_NVENC
