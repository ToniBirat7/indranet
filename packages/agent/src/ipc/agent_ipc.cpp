#include <iostream>
#include <string>
#include <functional>

// IPC channel between the C++ agent and the Tauri desktop client.
// The Tauri app (running in the same host machine) communicates over a
// local Unix domain socket (Linux) or named pipe (Windows).
//
// Messages are newline-delimited JSON.

class AgentIPC {
public:
    using OnMessageCallback = std::function<void(const std::string& msgJson)>;

    bool start(const std::string& socketPath, OnMessageCallback onMessage) {
        // TODO (Windows): CreateNamedPipe(\\.\pipe\indranet-agent-<pid>)
        // TODO (Linux): bind() on AF_UNIX socket
        // TODO: Accept connection from Tauri app
        // TODO: Spawn read thread — parse newline-delimited JSON, call onMessage
        std::cout << "[ipc] Listening on " << socketPath << " (stub)\n";
        return false;
    }

    // Send a JSON message to the Tauri desktop client.
    void send(const std::string& msgJson) {
        // TODO: Write msgJson + "\n" to the IPC socket/pipe
    }

    void stop() {
        // TODO: Close socket/pipe, join read thread
    }
};
