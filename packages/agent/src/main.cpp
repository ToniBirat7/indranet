#include <iostream>
#include <string>
#include <signal.h>
#include <atomic>
#include <thread>
#include <chrono>
#include <nlohmann/json.hpp>

// Forward declarations
void runCaptureLoop();
void runBillingHeartbeat(const std::string& sessionId, const std::string& backendUrl);
void handleInput();

static std::atomic<bool> g_running{true};

void signalHandler(int sig) {
    std::cout << "[agent] Received signal " << sig << ", shutting down...\n";
    g_running = false;
}

struct AgentConfig {
    std::string backendUrl;
    std::string sessionId;
    std::string agentToken;
    int targetFps = 60;
    int targetBitrateMbps = 20;
};

AgentConfig parseArgs(int argc, char* argv[]) {
    AgentConfig cfg;
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        if (arg == "--backend" && i + 1 < argc)      cfg.backendUrl = argv[++i];
        else if (arg == "--session" && i + 1 < argc) cfg.sessionId = argv[++i];
        else if (arg == "--token" && i + 1 < argc)   cfg.agentToken = argv[++i];
        else if (arg == "--fps" && i + 1 < argc)     cfg.targetFps = std::stoi(argv[++i]);
    }
    return cfg;
}

int main(int argc, char* argv[]) {
    signal(SIGINT, signalHandler);
    signal(SIGTERM, signalHandler);

    AgentConfig cfg = parseArgs(argc, argv);

    if (cfg.backendUrl.empty() || cfg.sessionId.empty() || cfg.agentToken.empty()) {
        std::cerr << "Usage: indranet-agent --backend <url> --session <id> --token <jwt>\n";
        return 1;
    }

    std::cout << "[agent] Starting for session " << cfg.sessionId << "\n";
    std::cout << "[agent] Backend: " << cfg.backendUrl << "\n";

    // TODO: Initialize DXGI capture
    // TODO: Initialize NVENC/AMF encoder
    // TODO: Initialize WebRTC streamer
    // TODO: Initialize ViGEm input handler
    // TODO: Launch Windows Sandbox for session
    // TODO: Connect IPC to Tauri desktop client

    // Main event loop — runs until signal or session end
    while (g_running) {
        // TODO: tick capture → encode → stream pipeline
        std::this_thread::sleep_for(std::chrono::milliseconds(16)); // ~60fps tick
    }

    // Graceful shutdown
    std::cout << "[agent] Shutting down...\n";
    // TODO: Destroy sandbox (MUST happen — security invariant)
    // TODO: Send final heartbeat to backend with ENDED state
    // TODO: Release ViGEm device
    // TODO: Close WebRTC peer connection

    return 0;
}
