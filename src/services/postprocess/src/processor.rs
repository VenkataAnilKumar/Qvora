use std::path::PathBuf;
use aws_sdk_s3::Client as S3Client;
use anyhow::Result;
use tracing::info;
use tempfile::NamedTempFile;
use std::io::Write;
use crate::mux::MuxClient;

/// Processor config for a single job
pub struct ProcessorConfig {
    pub variant_id: String,
    pub _workspace_id: String,
    pub watermark: bool,
    pub add_captions: bool,
    pub script: Option<String>,
    pub s3_client: S3Client,
    pub r2_bucket: String,
    pub _r2_endpoint: String,
    pub mux_client: MuxClient,
}

impl ProcessorConfig {
    /// Full pipeline: download → ffmpeg → upload
    pub async fn process(
        &self,
        input_r2_key: &str,
        output_r2_key: &str,
    ) -> Result<ProcessResult> {
        info!(
            variant_id = %self.variant_id,
            input_r2_key = %input_r2_key,
            "starting postprocess pipeline"
        );

        // Step 1: Download from R2
        let input_path = self.download_from_r2(input_r2_key).await?;
        info!(variant_id = %self.variant_id, "downloaded from R2");

        // Step 2: Run ffmpeg transforms
        let output_path = self
            .run_ffmpeg(
                &input_path,
                self.watermark,
                self.add_captions,
                self.script.as_deref(),
            )
            .await?;
        info!(variant_id = %self.variant_id, "ffmpeg processing complete");

        // Step 3: Upload processed video to R2
        let r2_presigned_url = self.upload_to_r2(&output_path, output_r2_key).await?;
        info!(variant_id = %self.variant_id, "uploaded to R2");

        // Step 4: Upload to Mux HLS from R2 URL
        let mux_result = self.mux_client.upload_from_url(&r2_presigned_url).await?;
        info!(
            variant_id = %self.variant_id,
            asset_id = %mux_result.asset_id,
            playback_id = %mux_result.playback_id,
            "uploaded to Mux"
        );

        // Cleanup
        let _ = std::fs::remove_file(&input_path);
        let _ = std::fs::remove_file(&output_path);

        Ok(ProcessResult {
            variant_id: self.variant_id.clone(),
            output_r2_key: output_r2_key.to_string(),
            mux_asset_id: mux_result.asset_id,
            mux_playable_id: mux_result.playback_id,
            duration_ms: 0, // TODO: extract from ffmpeg
        })
    }

    /// Download video from R2 to local tempfile
    async fn download_from_r2(&self, r2_key: &str) -> Result<PathBuf> {
        let mut temp_file = NamedTempFile::new()?;
        let temp_path = temp_file.path().to_path_buf();

        let body = self
            .s3_client
            .get_object()
            .bucket(&self.r2_bucket)
            .key(r2_key)
            .send()
            .await
            .map_err(|e| anyhow::anyhow!("R2 download failed: {}", e))?;

        let bytes = body
            .body
            .collect()
            .await
            .map_err(|e| anyhow::anyhow!("Failed to read R2 stream: {}", e))?
            .into_bytes();

        temp_file.write_all(&bytes)?;
        temp_file.flush()?;

        Ok(temp_path)
    }

    /// Run ffmpeg transformations on input video
    ///
    /// Pipeline:
    /// 1. Video scale to 1080×1920 (9:16 aspect)
    /// 2. Add letterbox if needed (pad to exact 1080×1920)
    /// 3. Overlay watermark (10% opacity, top-left)
    /// 4. Burn in captions (if script provided)
    /// 5. Transcode to H.264 (2500kbps video, 128kbps audio)
    /// 6. Output MP4 (H.264 + AAC)
    #[allow(unused_variables)]
    async fn run_ffmpeg(
        &self,
        input_path: &PathBuf,
        watermark: bool,
        add_captions: bool,
        script: Option<&str>,
    ) -> Result<PathBuf> {
        #[cfg(feature = "ffmpeg")]
        {
            return self.run_ffmpeg_with_cli(input_path, watermark, add_captions, script).await;
        }

        #[cfg(not(feature = "ffmpeg"))]
        {
            Err(anyhow::anyhow!(
                "ffmpeg feature not enabled; rebuild with --features ffmpeg in Docker"
            ))
        }
    }

    /// FFmpeg CLI implementation (enabled with --features ffmpeg)
    #[cfg(feature = "ffmpeg")]
    async fn run_ffmpeg_with_cli(
        &self,
        input_path: &PathBuf,
        watermark: bool,
        add_captions: bool,
        script: Option<&str>,
    ) -> Result<PathBuf> {
        use std::process::Command;

        let mut output_file = NamedTempFile::new()?;
        let output_path = output_file.path().to_path_buf();
        drop(output_file); // Close to allow ffmpeg to write

        // Build filter chain
        let mut filters = vec![];

        // Scale to 9:16 with letterbox
        filters.push("scale=1080:1920:force_original_aspect_ratio=decrease".to_string());
        filters.push("pad=1080:1920:(ow-iw)/2:(oh-ih)/2:color=black".to_string());

        // Add watermark overlay (if enabled)
        if watermark {
            // Simple watermark: draw text "Qvora" in bottom-right, 10% opacity
            filters.push(
                "drawtext=text='Qvora':fontsize=48:fontcolor=white@0.1:x=w-200:y=h-100:fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"
                    .to_string(),
            );
        }

        // Add captions (if script provided)
        if add_captions && script.is_some() {
            tracing::warn!(variant_id = %self.variant_id, "caption burn-in prepared but implementation deferred to Phase 3 MVP");
            // TODO: Implement SRT parsing and drawtext for each caption
            // filters.push(format!("subtitles='{}'", srt_file));
        }

        let filter_chain = filters.join(",");

        // Construct ffmpeg command
        let mut cmd = Command::new("ffmpeg");
        cmd.arg("-i")
            .arg(input_path)
            .arg("-filter:v")
            .arg(&filter_chain)
            .arg("-c:v")
            .arg("libx264") // H.264 codec
            .arg("-b:v")
            .arg("2500k") // 2.5 Mbps video bitrate
            .arg("-preset")
            .arg("faster") // Balance speed/quality (faster = less encode time)
            .arg("-c:a")
            .arg("aac") // AAC audio codec
            .arg("-b:a")
            .arg("128k") // 128kbps audio
            .arg("-y") // Overwrite output
            .arg(&output_path);

        info!(
            variant_id = %self.variant_id,
            filter_chain = %filter_chain,
            "running ffmpeg"
        );

        let output = cmd
            .output()
            .map_err(|e| anyhow::anyhow!("ffmpeg spawn failed: {}", e))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            tracing::warn!(variant_id = %self.variant_id, stderr = %stderr, "ffmpeg error");
            return Err(anyhow::anyhow!("ffmpeg failed: {}", stderr));
        }

        Ok(output_path)
    }

    /// Upload processed video to R2 and return presigned URL
    async fn upload_to_r2(&self, local_path: &PathBuf, r2_key: &str) -> Result<String> {
        let body = tokio::fs::read(local_path)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to read processed file: {}", e))?;

        self.s3_client
            .put_object()
            .bucket(&self.r2_bucket)
            .key(r2_key)
            .body(bytes::Bytes::from(body).into())
            .content_type("video/mp4")
            .send()
            .await
            .map_err(|e| anyhow::anyhow!("R2 upload failed: {}", e))?;

        // Generate presigned URL for Mux to download from (valid for 1 hour)
        use aws_sdk_s3::presigning::PresigningConfig;
        use std::time::Duration;

        let presigned_request = self
            .s3_client
            .get_object()
            .bucket(&self.r2_bucket)
            .key(r2_key)
            .presigned(
                PresigningConfig::builder()
                    .expires_in(Duration::from_secs(3600)) // 1 hour
                    .build()
                    .map_err(|e| anyhow::anyhow!("Presigned config failed: {}", e))?,
            )
            .await
            .map_err(|e| anyhow::anyhow!("Presigned URL generation failed: {}", e))?;

        Ok(presigned_request.uri().to_string())
    }
}

/// Result of a successful postprocess job
#[derive(Debug)]
pub struct ProcessResult {
    pub variant_id: String,
    pub output_r2_key: String,
    pub mux_asset_id: String,
    pub mux_playable_id: String,
    #[allow(dead_code)]
    pub duration_ms: u64,
}
