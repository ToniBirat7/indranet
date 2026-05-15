use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct SystemInfo {
    pub gpu_vendor: String,
    pub gpu_model: String,
    pub ram_gb: u32,
    pub cpu_cores: u32,
    pub os: String,
}

#[tauri::command]
pub async fn get_system_info() -> Result<SystemInfo, String> {
    // TODO: Query real hardware via platform APIs
    // Windows: DXGI adapter enumeration for GPU, WMI for RAM/CPU
    Ok(SystemInfo {
        gpu_vendor: "NVIDIA".to_string(),
        gpu_model: "Unknown (stub)".to_string(),
        ram_gb: 16,
        cpu_cores: num_cpus(),
        os: std::env::consts::OS.to_string(),
    })
}

fn num_cpus() -> u32 {
    // std::thread::available_parallelism gives logical cores
    std::thread::available_parallelism()
        .map(|n| n.get() as u32)
        .unwrap_or(4)
}
