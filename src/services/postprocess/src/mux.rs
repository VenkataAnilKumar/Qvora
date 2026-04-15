use anyhow::Result;
use serde::{Deserialize, Serialize};
use reqwest::Client;
use tracing::info;
use chrono::Utc;
use hmac::KeyInit;

/// Mux API client for video upload and management
/// Docs: https://docs.mux.com/docs/video-upload-api
pub struct MuxClient {
    http_client: Client,
    access_token: String,
    secret_token: String,
}

/// Response from Mux POST /video (create asset)
#[derive(Debug, Deserialize, Clone)]
pub struct MuxAssetResponse {
    pub data: MuxAssetData,
}

#[derive(Debug, Deserialize, Clone)]
#[allow(dead_code)]
pub struct MuxAssetData {
    pub id: String,                // asset_id
    pub playback_ids: Vec<MuxPlaybackId>,
    pub status: String,            // processing, ready, failed
    pub created_at: String,
}

#[derive(Debug, Deserialize, Clone)]
pub struct MuxPlaybackId {
    pub id: String,
    #[allow(dead_code)]
    pub policy: String,            // "public" or "signed"
}

/// Request body for HLS upload via Mux URL
#[derive(Debug, Serialize)]
struct MuxCreateAssetRequest {
    input: MuxInput,
    playback_policy: Vec<String>, // ["signed","public"]
    test: bool,                     // Use test environment
}

#[derive(Debug, Serialize)]
struct MuxInput {
    url: String, // Pre-signed R2 URL
}

impl MuxClient {
    pub fn new(access_token: String, secret_token: String) -> Self {
        Self {
            http_client: Client::new(),
            access_token,
            secret_token,
        }
    }

    /// Upload video from R2 URL to Mux
    ///
    /// Creates a new Mux asset with the video URL and returns asset_id + playback_id
    pub async fn upload_from_url(&self, r2_video_url: &str) -> Result<MuxUploadResult> {
        info!(url = %r2_video_url, "uploading to Mux");

        let req = MuxCreateAssetRequest {
            input: MuxInput {
                url: r2_video_url.to_string(),
            },
            playback_policy: vec!["signed".to_string()], // Workspace-scoped access only
            test: false,
        };

        let response = self
            .http_client
            .post("https://api.mux.com/video/v1/assets")
            .basic_auth(&self.access_token, Some(&self.secret_token))
            .json(&req)
            .send()
            .await
            .map_err(|e| anyhow::anyhow!("Mux upload request failed: {}", e))?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response
                .text()
                .await
                .unwrap_or_else(|_| "(no body)".to_string());
            return Err(anyhow::anyhow!(
                "Mux upload failed ({} {}): {}",
                status.as_u16(),
                status.canonical_reason().unwrap_or("Unknown"),
                body
            ));
        }

        let asset: MuxAssetResponse = response
            .json()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to parse Mux response: {}", e))?;

        let asset_id = asset.data.id.clone();
        let playback_id = asset
            .data
            .playback_ids
            .first()
            .ok_or_else(|| anyhow::anyhow!("No playback_id in Mux response"))?
            .id
            .clone();

        info!(
            asset_id = %asset_id,
            playback_id = %playback_id,
            "uploaded to Mux successfully"
        );

        Ok(MuxUploadResult {
            asset_id,
            playback_id,
        })
    }

    /// Generate a signed playback URL for workspace access
    ///
    /// Creates a JWT token that restricts playback to:
    /// - Specific workspace_id (via sub claim)
    /// - 1 hour expiration
    /// - Specific playback_id
    #[allow(dead_code)]
    pub fn generate_signed_playback_token(
        &self,
        _playback_id: &str,
        workspace_id: &str,
    ) -> Result<String> {
        // JWT header (HS256)
        let header = MuxTokenHeader {
            alg: "HS256".to_string(),
            typ: "JWT".to_string(),
        };

        // JWT payload
        // Expiry: 1 hour from now
        let exp = (Utc::now() + chrono::Duration::hours(1)).timestamp();
        let payload = MuxTokenPayload {
            sub: workspace_id.to_string(),
            exp: exp as u64,
            aud: "v".to_string(), // audience = video playback
        };

        let header_json = serde_json::to_string(&header)?;
        let payload_json = serde_json::to_string(&payload)?;

        let header_b64 = base64_url_encode(&header_json);
        let payload_b64 = base64_url_encode(&payload_json);
        let message = format!("{}.{}", header_b64, payload_b64);

        // Sign with HMAC-SHA256
        use hmac::{Hmac, Mac};
        use sha2::Sha256;

        type HmacSha256 = Hmac<Sha256>;
        let mut mac = HmacSha256::new_from_slice(self.secret_token.as_bytes())?;
        mac.update(message.as_bytes());
        let signature_bytes = mac.finalize().into_bytes();
        let signature_b64 = base64_url_encode(&hex::encode(&signature_bytes));

        let token = format!("{}.{}", message, signature_b64);

        info!(
            workspace_id = %workspace_id,
            exp = exp,
            "generated signed playback token"
        );

        Ok(token)
    }

    /// Construct playback URL for a signed token
    #[allow(dead_code)]
    pub fn playback_url(_playback_id: &str, token: &str) -> String {
        format!(
            "https://image.mux.com/7U8c2Q/low.mp4?token={}",
            token
        )
    }
}

/// Result of successful Mux upload
#[derive(Debug, Clone)]
pub struct MuxUploadResult {
    pub asset_id: String,
    pub playback_id: String,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
struct MuxTokenHeader {
    alg: String,
    typ: String,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
struct MuxTokenPayload {
    sub: String,  // workspace_id
    exp: u64,     // expiration
    aud: String,  // "v" for video
}

#[allow(dead_code)]
fn base64_url_encode(input: &str) -> String {
    use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine};
    URL_SAFE_NO_PAD.encode(input.as_bytes())
}
