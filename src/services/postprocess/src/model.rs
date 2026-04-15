use serde::{Deserialize, Serialize};

/// Input for POST /process
#[derive(Debug, Deserialize)]
pub struct ProcessRequest {
    pub variant_id: String,
    pub workspace_id: String,
    pub input_r2_key: String,
    pub output_r2_key: String,
    #[serde(default)]
    pub watermark: bool,
    #[serde(default)]
    pub add_captions: bool,
    pub script: Option<String>,
}

/// Output for POST /process (synchronously returned, background operation)
#[derive(Debug, Serialize)]
pub struct ProcessResponse {
    pub variant_id: String,
    pub output_r2_key: String,
    pub status: String, // "accepted" on successful queue
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mux_asset_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mux_playable_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}
