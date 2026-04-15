use axum::{http::StatusCode, response::{IntoResponse, Response}, Json};
use serde_json::json;

#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("R2 download failed: {0}")]
    #[allow(dead_code)]
    R2Download(String),
    #[error("R2 upload failed: {0}")]
    #[allow(dead_code)]
    R2Upload(String),
    #[error("FFmpeg error: {0}")]
    #[allow(dead_code)]
    Ffmpeg(String),
    #[error("Invalid request: {0}")]
    BadRequest(String),
    #[error("Internal error: {0}")]
    Internal(#[from] anyhow::Error),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            AppError::BadRequest(_) => (StatusCode::BAD_REQUEST, self.to_string()),
            _ => (StatusCode::INTERNAL_SERVER_ERROR, self.to_string()),
        };
        (status, Json(json!({ "error": message }))).into_response()
    }
}
