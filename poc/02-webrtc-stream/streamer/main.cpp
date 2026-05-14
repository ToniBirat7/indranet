// PoC 02: WebRTC C++ Streamer
// Connects to the signaling server, creates a WebRTC peer connection, and streams
// encoded video frames via GStreamer webrtcbin.
//
// TODO: Implement this once PoC 01 confirms NVENC is working.
// For now this is a stub showing the structure.

#include <cstdio>
#include <string>

// TODO: Include GStreamer headers
// #include <gst/gst.h>
// #include <gst/webrtc/webrtc.h>

int main(int argc, char* argv[]) {
    std::string signaling_url = "ws://localhost:8765/signal";
    std::string room_id = "test123";

    // Parse CLI args
    for (int i = 1; i < argc; i++) {
        if (std::string(argv[i]) == "--signaling" && i + 1 < argc) {
            signaling_url = argv[++i];
        } else if (std::string(argv[i]) == "--session" && i + 1 < argc) {
            room_id = argv[++i];
        }
    }

    printf("IndraNet PoC 02 - WebRTC Streamer\n");
    printf("Signaling: %s?room=%s&role=host\n\n", signaling_url.c_str(), room_id.c_str());

    // TODO: Initialize GStreamer
    // gst_init(&argc, &argv);

    // TODO: Build GStreamer pipeline
    // Pipeline: videotestsrc → nvh264enc → h264parse → rtph264pay → webrtcbin
    // For PoC: start with videotestsrc, then swap in DXGI capture from PoC 01
    //
    // GstElement* pipeline = gst_pipeline_new("poc02");
    // GstElement* src      = gst_element_factory_make("videotestsrc", "src");
    // GstElement* encoder  = gst_element_factory_make("nvh264enc", "encoder");
    // GstElement* parse    = gst_element_factory_make("h264parse", "parse");
    // GstElement* pay      = gst_element_factory_make("rtph264pay", "pay");
    // GstElement* webrtc   = gst_element_factory_make("webrtcbin", "webrtc");

    // TODO: Configure encoder for low latency
    // g_object_set(encoder,
    //     "bitrate", 15000,          // 15 Mbps
    //     "rc-mode", 4,              // CBR low-latency
    //     "gop-size", 120,           // Keyframe every 2s at 60fps
    //     "b-frames", 0,             // No B-frames
    //     NULL);

    // TODO: Connect to signaling server via WebSocket
    // Use libwebsockets or a simple HTTP upgrade implementation
    // Connect as role=host with the room ID

    // TODO: Handle WebRTC negotiation
    // on-negotiation-needed → create offer → send via WebSocket
    // on-answer-received → apply remote description
    // on-ice-candidate → send via WebSocket
    // on-ice-candidate-from-ws → add to peer connection

    // TODO: Start pipeline and run GLib main loop
    // gst_element_set_state(pipeline, GST_STATE_PLAYING);
    // GMainLoop* loop = g_main_loop_new(NULL, FALSE);
    // g_main_loop_run(loop);

    printf("TODO: Implement GStreamer WebRTC pipeline\n");
    printf("See poc/02-webrtc-stream/README.md for implementation plan\n");

    return 0;
}
