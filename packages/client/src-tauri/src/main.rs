// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
use commands::agent::AgentState;
use std::sync::Mutex;

fn main() {
    tauri::Builder::default()
        .manage(AgentState { process: Mutex::new(None) })
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            commands::agent::start_agent,
            commands::agent::stop_agent,
            commands::agent::get_agent_status,
            commands::session::get_session_info,
            commands::system::get_system_info,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
