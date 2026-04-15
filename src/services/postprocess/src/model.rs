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

/// Output for POST /process
#[derive(Debug, Serialize)]
pub struct ProcessResponse {
    pub variant_id: String,
    pub output_r2_key: String,
    pub status: String,
}
