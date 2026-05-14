# NVIDIA NVENC SDK Notes

## Question
What NVENC SDK settings achieve <3ms encode latency at 1080p60?

## Key Configuration for Low Latency

```cpp
NV_ENC_INITIALIZE_PARAMS init_params = {};
NV_ENC_CONFIG encode_config = {};

init_params.encodeGUID = NV_ENC_CODEC_H264_GUID;
init_params.presetGUID = NV_ENC_PRESET_P1_GUID;  // Fastest preset
init_params.encodeWidth = 1920;
init_params.encodeHeight = 1080;
init_params.frameRateNum = 60;
init_params.frameRateDen = 1;
init_params.encodeConfig = &encode_config;

// Rate control: low-latency CBR
encode_config.rcParams.rateControlMode = NV_ENC_PARAMS_RC_CBR_LOWDELAY_HQ;
encode_config.rcParams.averageBitRate = 15000000;  // 15 Mbps
encode_config.rcParams.maxBitRate = 15000000;

// No B-frames (they add latency)
encode_config.encodeCodecConfig.h264Config.idrPeriod = 120;  // Keyframe every 2s at 60fps
encode_config.encodeCodecConfig.h264Config.maxNumRefFrames = 1;
// B-frames default to 0 in low-latency presets

// Tune for low latency
encode_config.encodeCodecConfig.h264Config.outputBufferingPeriodSEI = 0;
encode_config.encodeCodecConfig.h264Config.outputPictureTimingSEI = 0;
```

## Findings
TODO: Fill in after running poc/01.

## Key Facts from Documentation

- `NV_ENC_PARAMS_RC_CBR_LOWDELAY_HQ`: Constant bitrate with low-latency tuning. Uses look-ahead = 0.
- `NV_ENC_PRESET_P1_GUID`: Fastest quality/speed tradeoff preset (was `NVENC_PRESET_LOW_LATENCY_HQ_GUID` in old API)
- B-frames: Setting `maxNumRefFrames = 1` disables B-frame references, reducing encode delay
- Zero-copy path: Pass `ID3D11Texture2D` directly to NVENC input buffer to avoid GPU→CPU→GPU round-trip

## Minimum SDK Version
NVENC SDK 12.x (ships with NVIDIA drivers 525+). Required for AV1 support. H.264 works with SDK 9+.

## Open Questions
- What is the actual p99 encode latency with this configuration on an RTX 3080?
- Does the zero-copy D3D11 path work when the DXGI surface is from inside a Windows Sandbox?
- Is there a measurable quality difference between P1 and P7 presets at 15 Mbps?
