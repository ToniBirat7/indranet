use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct SessionInfo {
    pub session_id: String,
    pub state: String,
    pub host_id: String,
    pub elapsed_seconds: u64,
}

#[tauri::command]
pub async fn get_session_info(
    session_id: String,
    backend_url: String,
    token: String,
) -> Result<SessionInfo, String> {
    // TODO: GET /v1/sessions/<id> with Authorization: Bearer <token>
    // For now return a stub response
    Ok(SessionInfo {
        session_id,
        state: "active".to_string(),
        host_id: "stub".to_string(),
        elapsed_seconds: 0,
    })
}
