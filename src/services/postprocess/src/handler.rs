use axum::{Json, http::StatusCode};
use serde_json::{Value, json};
use crate::model::{ProcessRequest, ProcessResponse};
use crate::processor::ProcessorConfig;
use crate::mux::MuxClient;
use crate::error::AppError;

/// GET /health
pub async fn health() -> Json<Value> {
    Json(json!({
        "ok": true,
        "service": "qvora-postprocess",
        "timestamp": chrono::Utc::now().to_rfc3339_opts(chrono::SecondsFormat::Secs, true)
    }))
}

/// POST /process
/// Accepts a processing job: fetch from R2, apply ffmpeg transforms, upload back to R2
///
/// Expected request body:
/// ```json
/// {
///   "variant_id": "var_abc",
///   "workspace_id": "ws_123",
///   "input_r2_key": "jobs/j1/variants/v1/raw.mp4",
///   "output_r2_key": "jobs/j1/variants/v1/processed.mp4",
///   "watermark": true,
///   "add_captions": false,
///   "script": null
/// }
/// ```
pub async fn process(
    Json(req): Json<ProcessRequest>,
) -> Result<(StatusCode, Json<ProcessResponse>), AppError> {
    tracing::info!(
        variant_id = %req.variant_id,
        workspace_id = %req.workspace_id,
        input_r2_key = %req.input_r2_key,
        "received postprocess request"
    );

    // Validate request
    if req.variant_id.is_empty() || req.input_r2_key.is_empty() {
        return Err(AppError::BadRequest(
            "variant_id and input_r2_key required".to_string(),
        ));
    }

    // Initialize S3 client for R2
    let s3_config = aws_config::load_from_env().await;
    let s3_client = aws_sdk_s3::Client::new(&s3_config);

    let r2_bucket = std::env::var("R2_BUCKET")
        .map_err(|_| AppError::Internal(anyhow::anyhow!("R2_BUCKET not set")))?;
    let r2_endpoint = std::env::var("R2_ENDPOINT")
        .map_err(|_| AppError::Internal(anyhow::anyhow!("R2_ENDPOINT not set")))?;

    // Initialize Mux client
    let mux_access_token = std::env::var("MUX_ACCESS_TOKEN")
        .map_err(|_| AppError::Internal(anyhow::anyhow!("MUX_ACCESS_TOKEN not set")))?;
    let mux_secret_token = std::env::var("MUX_SECRET_TOKEN")
        .map_err(|_| AppError::Internal(anyhow::anyhow!("MUX_SECRET_TOKEN not set")))?;
    let mux_client = MuxClient::new(mux_access_token, mux_secret_token);

    // Spawn async processing task (fire-and-forget)
    let variant_id = req.variant_id.clone();
    let input_key = req.input_r2_key.clone();
    let output_key = req.output_r2_key.clone();
    let _workspace_id = req.workspace_id.clone();
    let watermark = req.watermark;
    let add_captions = req.add_captions;
    let script = req.script.clone();

    tokio::spawn(async move {
        let processor = ProcessorConfig {
            variant_id: variant_id.clone(),
            _workspace_id,
            watermark,
            add_captions,
            script,
            s3_client,
            r2_bucket,
            _r2_endpoint: r2_endpoint,
            mux_client,
        };

        match processor.process(&input_key, &output_key).await {
            Ok(result) => {
                tracing::info!(
                    variant_id = %result.variant_id,
                    output_r2_key = %result.output_r2_key,
                    mux_asset_id = %result.mux_asset_id,
                    mux_playable_id = %result.mux_playable_id,
                    "postprocess completed successfully"
                );
                // TODO: Call back to API to mark variant as complete
                // PATCH /api/v1/variants/{id}/status with mux_asset_id, mux_playable_id, status="complete"
            }
            Err(e) => {
                tracing::error!(
                    variant_id = %variant_id,
                    error = %e,
                    "postprocess failed"
                );
                // TODO: Call back to API to mark variant as failed
                // PATCH /api/v1/variants/{id}/status with status="failed"
            }
        }
    });

    // Return 202 Accepted (processing in background)
    Ok((
        StatusCode::ACCEPTED,
        Json(ProcessResponse {
            variant_id: req.variant_id,
            output_r2_key: req.output_r2_key,
            status: "accepted".to_string(),
            mux_asset_id: None,
            mux_playable_id: None,
            error: None,
        }),
    ))
}
