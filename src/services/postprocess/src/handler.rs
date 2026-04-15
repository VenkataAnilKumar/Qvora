use axum::{Json, http::StatusCode};
use serde_json::{Value, json};
use crate::model::{ProcessRequest, ProcessResponse};
use crate::error::AppError;

/// GET /health
pub async fn health() -> Json<Value> {
    Json(json!({
        "ok": true,
        "service": "qvora-postprocess",
        "timestamp": chrono_now()
    }))
}

/// POST /process
/// Accepts a processing job: fetch from R2, apply ffmpeg transforms, upload back to R2
pub async fn process(
    Json(req): Json<ProcessRequest>,
) -> Result<(StatusCode, Json<ProcessResponse>), AppError> {
    tracing::info!(
        variant_id = %req.variant_id,
        input_r2_key = %req.input_r2_key,
        "starting postprocess job"
    );

    // TODO: implement full pipeline in Phase 3:
    //   1. Download from R2 (presigned GET)
    //   2. Write to tempfile
    //   3. Run ffmpeg: watermark overlay, caption burn-in (if requested), 9:16 crop/pad, H.264 transcode
    //   4. Upload processed file to R2 (presigned PUT)
    //   5. Return output_r2_key

    Ok((
        StatusCode::ACCEPTED,
        Json(ProcessResponse {
            variant_id: req.variant_id,
            output_r2_key: req.output_r2_key,
            status: "accepted".to_string(),
        }),
    ))
}

fn chrono_now() -> String {
    use std::time::{SystemTime, UNIX_EPOCH};
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs())
        .unwrap_or(0);
    format!("{}Z", secs)
}
