#include <iostream>
#include <string>
#include <chrono>
#include <thread>
#include <atomic>
#include <functional>

// Billing heartbeat client — sends periodic ticks to the backend.
// The backend deducts per-minute cost and returns the current session state.
// If the backend returns "kill", the agent must immediately destroy the sandbox.

class BillingClient {
public:
    enum class Action { Continue, Warn, Kill };

    using OnActionCallback = std::function<void(Action action, int secondsRemaining)>;

    struct Config {
        std::string backendUrl;
        std::string sessionId;
        std::string agentToken;
        int intervalSeconds = 60;
    };

    void start(const Config& cfg, OnActionCallback onAction) {
        m_cfg = cfg;
        m_running = true;
        m_thread = std::thread([this, onAction]() {
            while (m_running) {
                std::this_thread::sleep_for(std::chrono::seconds(m_cfg.intervalSeconds));
                if (!m_running) break;

                Action action = sendHeartbeat();
                // secondsRemaining would come from the response in real impl
                onAction(action, -1);
            }
        });
    }

    void stop() {
        m_running = false;
        if (m_thread.joinable()) m_thread.join();
    }

private:
    Action sendHeartbeat() {
        // TODO: POST /v1/sessions/<id>/heartbeat with Authorization: Bearer <token>
        // Response: {"action":"continue"|"warn"|"kill","seconds_remaining":300}
        std::cout << "[billing] Heartbeat tick (stub)\n";
        return Action::Continue;
    }

    Config m_cfg;
    std::atomic<bool> m_running{false};
    std::thread m_thread;
};
