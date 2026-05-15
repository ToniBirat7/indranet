use std::process::{Child, Command};
use std::sync::Mutex;
use tauri::State;

pub struct AgentState {
    pub process: Mutex<Option<Child>>,
}

#[tauri::command]
pub async fn start_agent(
    session_id: String,
    backend_url: String,
    token: String,
    state: State<'_, AgentState>,
) -> Result<(), String> {
    let agent_bin = if cfg!(windows) {
        "indranet-agent.exe"
    } else {
        "indranet-agent"
    };

    let child = Command::new(agent_bin)
        .arg("--session").arg(&session_id)
        .arg("--backend").arg(&backend_url)
        .arg("--token").arg(&token)
        .spawn()
        .map_err(|e| format!("Failed to start agent: {}", e))?;

    *state.process.lock().unwrap() = Some(child);
    Ok(())
}

#[tauri::command]
pub async fn stop_agent(state: State<'_, AgentState>) -> Result<(), String> {
    if let Some(mut child) = state.process.lock().unwrap().take() {
        child.kill().map_err(|e| format!("Failed to stop agent: {}", e))?;
    }
    Ok(())
}

#[tauri::command]
pub async fn get_agent_status(state: State<'_, AgentState>) -> Result<String, String> {
    let mut guard = state.process.lock().unwrap();
    match guard.as_mut() {
        None => Ok("stopped".to_string()),
        Some(child) => match child.try_wait() {
            Ok(None) => Ok("running".to_string()),
            Ok(Some(status)) => {
                *guard = None;
                Ok(format!("exited:{}", status.code().unwrap_or(-1)))
            }
            Err(e) => Err(e.to_string()),
        },
    }
}
