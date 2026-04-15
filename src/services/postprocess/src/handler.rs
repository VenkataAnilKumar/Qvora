use axum::{Json, http::StatusCode};
use serde_json::{Value, json};
use crate::model::{PostprocessCallbackRequest, ProcessRequest, ProcessResponse};
use crate::processor::ProcessorConfig;
use crate::mux::MuxClient;
use crate::error::AppError;

fn classify_failure(error_text: &str) -> &'static str {
    let normalized = error_text.to_lowercase();

    if normalized.contains("ffmpeg") || normalized.contains("transcode") || normalized.contains("codec") {
        return "ffmpeg";
    }
    if normalized.contains("mux") || normalized.contains("playback") || normalized.contains("asset") {
        return "mux";
    }
    if normalized.contains("s3") || normalized.contains("r2") || normalized.contains("bucket") || normalized.contains("upload") || normalized.contains("download") {
        return "storage";
    }
    if normalized.contains("timeout") || normalized.contains("dns") || normalized.contains("connection") || normalized.contains("network") {
        return "network";
    }
    if normalized.contains("callback") || normalized.contains("internal api") {
        return "callback";
    }

    "unknown"
}

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
    if req.request_id.is_empty()
        || req.job_id.is_empty()
        || req.variant_id.is_empty()
        || req.workspace_id.is_empty()
        || req.input_r2_key.is_empty()
    {
        return Err(AppError::BadRequest(
            "request_id, job_id, variant_id, workspace_id and input_r2_key required".to_string(),
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
    let request_id = req.request_id.clone();
    let job_id = req.job_id.clone();
    let variant_id = req.variant_id.clone();
    let input_key = req.input_r2_key.clone();
    let output_key = req.output_r2_key.clone();
    let workspace_id = req.workspace_id.clone();
    let watermark = req.watermark;
    let add_captions = req.add_captions;
    let script = req.script.clone();

    tokio::spawn(async move {
        let processor = ProcessorConfig {
			request_id: request_id.clone(),
			job_id: job_id.clone(),
            variant_id: variant_id.clone(),
			workspace_id: workspace_id.clone(),
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
                    request_id = %request_id,
                    job_id = %job_id,
                    variant_id = %result.variant_id,
                    output_r2_key = %result.output_r2_key,
                    mux_asset_id = %result.mux_asset_id,
                    mux_playable_id = %result.mux_playable_id,
                    "postprocess completed successfully"
                );

                if let Err(callback_err) = send_postprocess_callback(PostprocessCallbackRequest {
                    request_id: request_id.clone(),
                    job_id: job_id.clone(),
                    variant_id: result.variant_id.clone(),
                    workspace_id: workspace_id.clone(),
                    status: "success".to_string(),
                    output_r2_key: result.output_r2_key.clone(),
                    mux_asset_id: Some(result.mux_asset_id.clone()),
                    mux_playable_id: Some(result.mux_playable_id.clone()),
                    duration_ms: Some(result.duration_ms),
                    error_message: None,
                }).await {
                    tracing::error!(
                        request_id = %request_id,
                        variant_id = %variant_id,
                        error = %callback_err,
                        "failed to send postprocess success callback"
                    );
                }
            }
            Err(e) => {
				let error_text = e.to_string();
				let error_type = classify_failure(&error_text);
                tracing::error!(
                    request_id = %request_id,
                    job_id = %job_id,
                    variant_id = %variant_id,
					workspace_id = %workspace_id,
					error_type = %error_type,
                    error = %e,
                    "postprocess failed"
                );

                if let Err(callback_err) = send_postprocess_callback(PostprocessCallbackRequest {
                    request_id: request_id.clone(),
                    job_id: job_id.clone(),
                    variant_id: variant_id.clone(),
                    workspace_id: workspace_id.clone(),
                    status: "failed".to_string(),
                    output_r2_key: output_key.clone(),
                    mux_asset_id: None,
                    mux_playable_id: None,
                    duration_ms: None,
                    error_message: Some(format!("{}: {}", error_type, error_text)),
                }).await {
                    tracing::error!(
                        request_id = %request_id,
                        variant_id = %variant_id,
                        error = %callback_err,
                        "failed to send postprocess failure callback"
                    );
                }
            }
        }
    });

    // Return 202 Accepted (processing in background)
    Ok((
        StatusCode::ACCEPTED,
        Json(ProcessResponse {
            request_id: req.request_id,
            job_id: req.job_id,
            variant_id: req.variant_id,
            output_r2_key: req.output_r2_key,
            status: "accepted".to_string(),
            mux_asset_id: None,
            mux_playable_id: None,
            error: None,
        }),
    ))
}

async fn send_postprocess_callback(payload: PostprocessCallbackRequest) -> anyhow::Result<()> {
    let api_base = std::env::var("API_BASE_URL")
        .map_err(|_| anyhow::anyhow!("API_BASE_URL not set"))?;

    let internal_api_key = std::env::var("INTERNAL_API_KEY")
        .map_err(|_| anyhow::anyhow!("INTERNAL_API_KEY not set"))?;

    let callback_url = format!(
        "{}/api/v1/internal/postprocess/callback",
        api_base.trim_end_matches('/')
    );

    let client = reqwest::Client::new();
    let response = client
        .post(callback_url)
        .header("Content-Type", "application/json")
        .header("X-User-Id", "postprocess-service")
        .header("X-Org-Id", payload.workspace_id.clone())
        .header("X-Org-Role", "worker")
        .header("X-Internal-Api-Key", internal_api_key)
        .json(&payload)
        .send()
        .await
        .map_err(|e| anyhow::anyhow!("postprocess callback request failed: {}", e))?;

    if !response.status().is_success() {
        let status = response.status();
        let body = response
            .text()
            .await
            .unwrap_or_else(|_| "(no body)".to_string());
        return Err(anyhow::anyhow!(
            "postprocess callback failed ({} {}): {}",
            status.as_u16(),
            status.canonical_reason().unwrap_or("Unknown"),
            body
        ));
    }

    Ok(())
}
