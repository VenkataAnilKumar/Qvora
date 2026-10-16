use serde::{Deserialize, Serialize};

/// Input for POST /process
#[derive(Debug, Deserialize)]
pub struct ProcessRequest {
    pub request_id: String,
    pub job_id: String,
    pub variant_id: String,
    pub workspace_id: String,
    pub input_r2_key: String,
    pub output_r2_key: String,
    #[serde(default)]
    pub watermark: bool,
    #[serde(default)]
    pub add_captions: bool,
    pub script: Option<String>,
    /// R2 key of the brand logo (PNG/SVG) to overlay. When absent, `watermark`
    /// falls back to a static text mark so existing behavior doesn't regress.
    #[serde(default)]
    pub logo_r2_key: Option<String>,
    /// Target aspect ratio for the output frame. Defaults to "9:16" (current
    /// behavior). Accepts "9:16", "1:1", or "16:9" — ready for the Export
    /// Engine to request per-platform renditions once it exists.
    #[serde(default)]
    pub aspect_ratio: Option<String>,
}

/// Output for POST /process (synchronously returned, background operation)
#[derive(Debug, Serialize)]
pub struct ProcessResponse {
    pub request_id: String,
    pub job_id: String,
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

#[derive(Debug, Serialize)]
pub struct PostprocessCallbackRequest {
    pub request_id: String,
    pub job_id: String,
    pub variant_id: String,
    pub workspace_id: String,
    pub status: String, // success | failed
    pub output_r2_key: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mux_asset_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mux_playable_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub duration_ms: Option<u64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error_message: Option<String>,
}
