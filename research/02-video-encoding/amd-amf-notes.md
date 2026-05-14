# AMD AMF (Advanced Media Framework) Notes

## Question
Can AMD GPUs meet our <3ms encode latency target using AMF?

## Key Configuration

```cpp
// AMF H.264 low-latency encoder settings
AMFFactory* factory;
AMFContext* context;
AMFComponent* encoder;

factory->CreateComponent(context, AMFVideoEncoderVCE_AVC, &encoder);

encoder->SetProperty(AMF_VIDEO_ENCODER_USAGE, AMF_VIDEO_ENCODER_USAGE_LOWLATENCY);
encoder->SetProperty(AMF_VIDEO_ENCODER_PROFILE, AMF_VIDEO_ENCODER_PROFILE_HIGH);
encoder->SetProperty(AMF_VIDEO_ENCODER_TARGET_BITRATE, 15000000);  // 15 Mbps
encoder->SetProperty(AMF_VIDEO_ENCODER_RATE_CONTROL_METHOD, 
                     AMF_VIDEO_ENCODER_RATE_CONTROL_METHOD_CBR);
encoder->SetProperty(AMF_VIDEO_ENCODER_B_PIC_PATTERN, 0);  // No B-frames
```

## Findings
TODO: Test on AMD hardware.

## Open Questions
- Is AMD AMF latency comparable to NVENC at 1080p60?
- Does AMF support the same zero-copy D3D11 surface input path as NVENC?
- What AMF SDK version is required? Is it bundled with AMD drivers?
