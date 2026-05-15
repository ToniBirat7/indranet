#include <iostream>
#include <string>

// Windows Sandbox lifecycle management
// Docs: https://learn.microsoft.com/en-us/windows/security/application-security/application-isolation/windows-sandbox/windows-sandbox-overview
//
// SECURITY INVARIANT: sandbox MUST be destroyed when session ends.
// The destroy() call is non-optional — it must run even on crash paths.

#ifdef _WIN32
#include <windows.h>

class WindowsSandbox {
public:
    struct Config {
        bool enableGpu = true;           // VGpu Enabled — required for GPU workloads
        bool enableNetworking = true;    // Required for game downloads, updates
        std::string mappedFolderPath;    // Host path to expose read-only (game install)
        std::string startupCommand;      // Command to run when sandbox starts
    };

    // Launches a Windows Sandbox using a generated .wsb config file.
    // Returns false if sandbox fails to start within timeoutMs.
    bool launch(const Config& cfg, int timeoutMs = 30000) {
        // TODO: Write .wsb XML to temp file:
        //   <Configuration>
        //     <VGpu>Enable</VGpu>
        //     <Networking>Enable</Networking>
        //     <MappedFolders>...</MappedFolders>
        //     <LogonCommand><Command>...</Command></LogonCommand>
        //   </Configuration>
        // TODO: ShellExecuteEx("WindowsSandbox.exe", wsb_path)
        // TODO: Wait for sandbox process to appear, timeout if not
        std::cout << "[sandbox] Launch (stub)\n";
        return false;
    }

    // Terminates the sandbox. MUST be called on all exit paths.
    // This is the enforcement point for the ephemeral isolation guarantee.
    void destroy() {
        if (!m_launched) return;
        // TODO: TerminateProcess(m_sandboxHandle, 0)
        // TODO: WaitForSingleObject to confirm termination
        // TODO: DeleteFile(m_wsbPath) to clean up config
        m_launched = false;
        std::cout << "[sandbox] Destroyed\n";
    }

    bool isRunning() const {
        // TODO: WaitForSingleObject(m_sandboxHandle, 0) == WAIT_TIMEOUT
        return false;
    }

    ~WindowsSandbox() {
        destroy(); // Safety net — prefer calling destroy() explicitly
    }

private:
    bool m_launched = false;
    // HANDLE m_sandboxHandle = nullptr;
    // std::string m_wsbPath;
};
#endif // _WIN32
