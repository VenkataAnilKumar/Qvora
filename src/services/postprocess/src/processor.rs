// Placeholder — ffmpeg processing logic will be implemented in Phase 3
// Handles: watermark overlay, caption burn-in, 9:16 crop/pad, H.264 transcode
pub struct ProcessorConfig {
    pub watermark_r2_key: Option<String>,
    pub add_captions: bool,
    pub script: Option<String>,
}
