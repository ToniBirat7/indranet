#include <iostream>
#include <string>
#include <functional>

// WebRTC streaming via GStreamer webrtcbin (PoC phase)
// Phase 2 will migrate to libwebrtc directly for lower overhead.
//
// Latency budget for this module: < 5ms packetization + < 30ms network

class WebRTCStreamer {
public:
    struct Config {
        std::string signalingUrl;   // wss://backend/v1/signal?session=<id>
        std::string agentToken;
        int rtpPayloadType = 96;    // H.264 dynamic PT
    };

    using OnInputCallback = std::function<void(const std::string& inputJson)>;

    bool init(const Config& cfg, OnInputCallback onInput) {
        // TODO: Connect WebSocket to signalingUrl with Authorization: Bearer <agentToken>
        // TODO: Initialize GStreamer pipeline:
        //   appsrc → h264parse → rtph264pay → webrtcbin
        // TODO: Handle offer/answer exchange via signaling WS
        // TODO: Open 'input' data channel for receiving user input events
        std::cout << "[stream] WebRTC streamer init (stub)\n";
        return false;
    }

    // Push a single encoded H.264 NAL unit into the GStreamer pipeline
    void pushFrame(const uint8_t* data, size_t size, long long timestampUs) {
        // TODO: Create GstBuffer, push to appsrc
    }

    void shutdown() {
        // TODO: Send ENDED signal over signaling WebSocket
        // TODO: Close GStreamer pipeline
        // TODO: Close WebSocket
    }
};
