use std::io::Write;
use std::path::{Path, PathBuf};
use std::time::Instant;

use anyhow::{anyhow, Result};
use aws_sdk_s3::Client as S3Client;
use tempfile::NamedTempFile;
use tracing::info;

use crate::mux::MuxClient;

#[cfg(feature = "ffmpeg")]
extern crate ffmpeg_next as ffmpeg;

#[cfg(feature = "ffmpeg")]
use ffmpeg::{codec, encoder, filter, format, frame, media, Dictionary, Packet, Rational};

#[cfg(feature = "ffmpeg")]
use anyhow::Context;

#[cfg(feature = "ffmpeg")]
use std::collections::HashMap;

#[cfg(feature = "ffmpeg")]
use std::sync::OnceLock;

#[cfg(feature = "ffmpeg")]
static FFMPEG_INIT: OnceLock<()> = OnceLock::new();

/// Processor config for a single job
pub struct ProcessorConfig {
    pub request_id: String,
    pub job_id: String,
    pub variant_id: String,
    pub workspace_id: String,
    pub watermark: bool,
    pub add_captions: bool,
    pub script: Option<String>,
    pub s3_client: S3Client,
    pub r2_bucket: String,
    pub _r2_endpoint: String,
    pub mux_client: MuxClient,
}

impl ProcessorConfig {
    /// Full pipeline: download -> transcode -> upload
    pub async fn process(
        &self,
        input_r2_key: &str,
        output_r2_key: &str,
    ) -> Result<ProcessResult> {
        let pipeline_start = Instant::now();
        info!(
            request_id = %self.request_id,
            job_id = %self.job_id,
            variant_id = %self.variant_id,
            workspace_id = %self.workspace_id,
            input_r2_key = %input_r2_key,
            "starting postprocess pipeline"
        );

        let download_start = Instant::now();
        let input_path = self.download_from_r2(input_r2_key).await?;
        let download_ms = download_start.elapsed().as_millis() as u64;
        info!(
            request_id = %self.request_id,
            job_id = %self.job_id,
            variant_id = %self.variant_id,
            workspace_id = %self.workspace_id,
            stage = "download",
            duration_ms = download_ms,
            "postprocess stage complete"
        );

        let transcode_start = Instant::now();
        let output_path = self
            .run_transcode(
                &input_path,
                self.watermark,
                self.add_captions,
                self.script.as_deref(),
            )
            .await?;
        let transcode_ms = transcode_start.elapsed().as_millis() as u64;
        info!(
            request_id = %self.request_id,
            job_id = %self.job_id,
            variant_id = %self.variant_id,
            workspace_id = %self.workspace_id,
            stage = "transcode",
            duration_ms = transcode_ms,
            "postprocess stage complete"
        );

        let upload_start = Instant::now();
        let r2_presigned_url = self.upload_to_r2(&output_path, output_r2_key).await?;
        let upload_ms = upload_start.elapsed().as_millis() as u64;
        info!(
            request_id = %self.request_id,
            job_id = %self.job_id,
            variant_id = %self.variant_id,
            workspace_id = %self.workspace_id,
            stage = "upload",
            duration_ms = upload_ms,
            "postprocess stage complete"
        );

        let mux_start = Instant::now();
        let mux_result = self
            .mux_client
            .upload_from_url(&r2_presigned_url, &self.variant_id)
            .await?;
        let mux_ms = mux_start.elapsed().as_millis() as u64;
        info!(
            request_id = %self.request_id,
            job_id = %self.job_id,
            variant_id = %self.variant_id,
            workspace_id = %self.workspace_id,
            asset_id = %mux_result.asset_id,
            playback_id = %mux_result.playback_id,
            stage = "mux",
            duration_ms = mux_ms,
            "postprocess stage complete"
        );

        let _ = std::fs::remove_file(&input_path);
        let _ = std::fs::remove_file(&output_path);

        let total_ms = pipeline_start.elapsed().as_millis() as u64;
        info!(
            request_id = %self.request_id,
            job_id = %self.job_id,
            variant_id = %self.variant_id,
            workspace_id = %self.workspace_id,
            download_ms = download_ms,
            transcode_ms = transcode_ms,
            upload_ms = upload_ms,
            mux_ms = mux_ms,
            total_ms = total_ms,
            "postprocess pipeline complete"
        );

        Ok(ProcessResult {
            variant_id: self.variant_id.clone(),
            output_r2_key: output_r2_key.to_string(),
            mux_asset_id: mux_result.asset_id,
            mux_playable_id: mux_result.playback_id,
            duration_ms: total_ms,
        })
    }

    async fn download_from_r2(&self, r2_key: &str) -> Result<PathBuf> {
        let mut temp_file = NamedTempFile::new()?;
        let temp_path = temp_file.path().to_path_buf();

        if r2_key.starts_with("http://") || r2_key.starts_with("https://") {
            let response = reqwest::Client::new()
                .get(r2_key)
                .send()
                .await
                .map_err(|e| anyhow!("URL download failed: {}", e))?;

            if !response.status().is_success() {
                return Err(anyhow!(
                    "URL download returned HTTP {}",
                    response.status().as_u16()
                ));
            }

            let bytes = response
                .bytes()
                .await
                .map_err(|e| anyhow!("Failed to read URL response body: {}", e))?;

            temp_file.write_all(&bytes)?;
            temp_file.flush()?;
            return Ok(temp_path);
        }

        let body = self
            .s3_client
            .get_object()
            .bucket(&self.r2_bucket)
            .key(r2_key)
            .send()
            .await
            .map_err(|e| anyhow!("R2 download failed: {}", e))?;

        let bytes = body
            .body
            .collect()
            .await
            .map_err(|e| anyhow!("Failed to read R2 stream: {}", e))?
            .into_bytes();

        temp_file.write_all(&bytes)?;
        temp_file.flush()?;

        Ok(temp_path)
    }

    #[allow(unused_variables)]
    async fn run_transcode(
        &self,
        input_path: &Path,
        watermark: bool,
        add_captions: bool,
        script: Option<&str>,
    ) -> Result<PathBuf> {
        #[cfg(feature = "ffmpeg")]
        {
            return self
                .run_transcode_with_bindings(input_path, watermark, add_captions, script)
                .await;
        }

        #[cfg(not(feature = "ffmpeg"))]
        {
            Err(anyhow!(
                "ffmpeg feature not enabled; rebuild with --features ffmpeg in Docker"
            ))
        }
    }

    #[cfg(feature = "ffmpeg")]
    async fn run_transcode_with_bindings(
        &self,
        input_path: &Path,
        watermark: bool,
        add_captions: bool,
        script: Option<&str>,
    ) -> Result<PathBuf> {
        let output_file = NamedTempFile::new()?;
        let output_path = output_file.path().to_path_buf();
        drop(output_file);

        let filter_spec = build_filter_spec(watermark, add_captions, script);

        info!(
            variant_id = %self.variant_id,
            filter_spec = %filter_spec,
            "running FFmpeg bindings pipeline"
        );

        tokio::task::spawn_blocking({
            let input_path = input_path.to_path_buf();
            let output_path = output_path.clone();
            move || transcode_video(&input_path, &output_path, &filter_spec)
        })
        .await
        .map_err(|e| anyhow!("ffmpeg task join failed: {e}"))??;

        Ok(output_path)
    }

    async fn upload_to_r2(&self, local_path: &Path, r2_key: &str) -> Result<String> {
        let body = tokio::fs::read(local_path)
            .await
            .map_err(|e| anyhow!("Failed to read processed file: {}", e))?;

        self.s3_client
            .put_object()
            .bucket(&self.r2_bucket)
            .key(r2_key)
            .body(bytes::Bytes::from(body).into())
            .content_type("video/mp4")
            .send()
            .await
            .map_err(|e| anyhow!("R2 upload failed: {}", e))?;

        use aws_sdk_s3::presigning::PresigningConfig;
        use std::time::Duration;

        let presigned_request = self
            .s3_client
            .get_object()
            .bucket(&self.r2_bucket)
            .key(r2_key)
            .presigned(
                PresigningConfig::builder()
                    .expires_in(Duration::from_secs(3600))
                    .build()
                    .map_err(|e| anyhow!("Presigned config failed: {}", e))?,
            )
            .await
            .map_err(|e| anyhow!("Presigned URL generation failed: {}", e))?;

        Ok(presigned_request.uri().to_string())
    }
}

#[cfg(feature = "ffmpeg")]
fn transcode_video(input_path: &Path, output_path: &Path, filter_spec: &str) -> Result<()> {
    init_ffmpeg()?;

    let input_str = input_path.to_string_lossy().into_owned();
    let output_str = output_path.to_string_lossy().into_owned();

    let mut ictx = format::input(&input_str)
        .with_context(|| format!("failed to open input {}", input_path.display()))?;
    let mut octx = format::output(&output_str)
        .with_context(|| format!("failed to open output {}", output_path.display()))?;

    let mut stream_mapping: Vec<isize> = vec![0; ictx.nb_streams() as _];
    let mut ist_time_bases = vec![Rational(0, 0); ictx.nb_streams() as _];
    let mut ost_time_bases = vec![Rational(0, 0); ictx.nb_streams() as _];
    let mut transcoders = HashMap::new();
    let mut ost_index = 0;

    for (ist_index, ist) in ictx.streams().enumerate() {
        let medium = ist.parameters().medium();

        if medium != media::Type::Audio
            && medium != media::Type::Video
            && medium != media::Type::Subtitle
        {
            stream_mapping[ist_index] = -1;
            continue;
        }

        stream_mapping[ist_index] = ost_index;
        ist_time_bases[ist_index] = ist.time_base();

        if medium == media::Type::Video {
            transcoders.insert(
                ist_index,
                VideoTranscoder::new(&ist, &mut octx, ost_index as usize, filter_spec)
                    .with_context(|| {
                        format!("failed to initialize video transcoder for stream {ist_index}")
                    })?,
            );
        } else {
            let mut ost = octx
                .add_stream(encoder::find(codec::Id::None))
                .context("failed to add passthrough stream")?;
            ost.set_parameters(ist.parameters());
            unsafe {
                (*ost.parameters().as_mut_ptr()).codec_tag = 0;
            }
        }

        ost_index += 1;
    }

    octx.set_metadata(ictx.metadata().to_owned());
    octx.write_header().context("failed to write output header")?;

    for (index, _) in octx.streams().enumerate() {
        ost_time_bases[index] = octx
            .stream(index)
            .ok_or_else(|| anyhow!("missing output stream {index}"))?
            .time_base();
    }

    for (stream, mut packet) in ictx.packets() {
        let ist_index = stream.index();
        let mapped_index = stream_mapping[ist_index];

        if mapped_index < 0 {
            continue;
        }

        let ost_time_base = ost_time_bases[mapped_index as usize];

        if let Some(transcoder) = transcoders.get_mut(&ist_index) {
            transcoder.send_packet_to_decoder(&packet)?;
            transcoder.receive_and_process_decoded_frames(&mut octx, ost_time_base)?;
        } else {
            packet.rescale_ts(ist_time_bases[ist_index], ost_time_base);
            packet.set_position(-1);
            packet.set_stream(mapped_index as usize);
            packet
                .write_interleaved(&mut octx)
                .context("failed to write passthrough packet")?;
        }
    }

    for (ist_index, transcoder) in transcoders.iter_mut() {
        let ost_time_base = ost_time_bases[*ist_index];
        transcoder.send_eof_to_decoder()?;
        transcoder.receive_and_process_decoded_frames(&mut octx, ost_time_base)?;
        transcoder.flush_filter()?;
        transcoder.get_and_process_filtered_frames(&mut octx, ost_time_base)?;
        transcoder.send_eof_to_encoder()?;
        transcoder.receive_and_process_encoded_packets(&mut octx, ost_time_base)?;
    }

    octx.write_trailer().context("failed to finalize output")?;
    Ok(())
}

#[cfg(feature = "ffmpeg")]
struct VideoTranscoder {
    ost_index: usize,
    decoder: ffmpeg::decoder::Video,
    encoder: ffmpeg::encoder::Video,
    filter: filter::Graph,
    input_time_base: Rational,
}

#[cfg(feature = "ffmpeg")]
impl VideoTranscoder {
    fn new(
        ist: &format::stream::Stream,
        octx: &mut format::context::Output,
        ost_index: usize,
        filter_spec: &str,
    ) -> Result<Self> {
        let global_header = octx.format().flags().contains(format::Flags::GLOBAL_HEADER);
        let decoder = ffmpeg::codec::context::Context::from_parameters(ist.parameters())?
            .decoder()
            .video()?;

        let codec = encoder::find(codec::Id::H264)
            .ok_or_else(|| anyhow!("H.264 encoder not available"))?;
        let mut ost = octx
            .add_stream(Some(codec))
            .context("failed to add video output stream")?;

        let mut encoder = ffmpeg::codec::context::Context::new_with_codec(codec)
            .encoder()
            .video()?;
        ost.set_parameters(&encoder);
        encoder.set_width(1080);
        encoder.set_height(1920);
        encoder.set_aspect_ratio((1, 1));
        encoder.set_format(ffmpeg::format::Pixel::YUV420P);
        encoder.set_time_base(ist.time_base());
        encoder.set_frame_rate(decoder.frame_rate());
        encoder.set_bit_rate(2_500_000);
        encoder.set_max_b_frames(2);
        encoder.set_gop(48);

        if global_header {
            encoder.set_flags(codec::Flags::GLOBAL_HEADER);
        }

        let mut options = Dictionary::new();
        options.set("preset", "faster");
        let opened_encoder = encoder.open_as_with(codec, options)?;
        ost.set_parameters(&opened_encoder);

        let filter = build_video_filter(filter_spec, &decoder, &opened_encoder)?;

        Ok(Self {
            ost_index,
            decoder,
            encoder: opened_encoder,
            filter,
            input_time_base: ist.time_base(),
        })
    }

    fn send_packet_to_decoder(&mut self, packet: &Packet) -> Result<()> {
        self.decoder
            .send_packet(packet)
            .context("failed to feed packet to decoder")
    }

    fn send_eof_to_decoder(&mut self) -> Result<()> {
        self.decoder.send_eof().context("failed to send decoder EOF")
    }

    fn send_frame_to_encoder(&mut self, frame: &frame::Video) -> Result<()> {
        self.encoder
            .send_frame(frame)
            .context("failed to feed frame to encoder")
    }

    fn send_eof_to_encoder(&mut self) -> Result<()> {
        self.encoder.send_eof().context("failed to send encoder EOF")
    }

    fn add_frame_to_filter(&mut self, frame: &frame::Video) -> Result<()> {
        self.filter
            .get("in")
            .ok_or_else(|| anyhow!("missing filter input"))?
            .source()
            .add(frame)
            .context("failed to push frame into filter graph")
    }

    fn flush_filter(&mut self) -> Result<()> {
        self.filter
            .get("in")
            .ok_or_else(|| anyhow!("missing filter input"))?
            .source()
            .flush()
            .context("failed to flush filter graph")
    }

    fn get_and_process_filtered_frames(
        &mut self,
        octx: &mut format::context::Output,
        ost_time_base: Rational,
    ) -> Result<()> {
        let mut filtered = frame::Video::empty();

        loop {
            let result = self
                .filter
                .get("out")
                .ok_or_else(|| anyhow!("missing filter output"))?
                .sink()
                .frame(&mut filtered);

            match result {
                Ok(()) => {
                    self.send_frame_to_encoder(&filtered)?;
                    self.receive_and_process_encoded_packets(octx, ost_time_base)?;
                }
                Err(ffmpeg::Error::Other { errno }) if errno == ffmpeg::util::error::EAGAIN => {
                    break
                }
                Err(ffmpeg::Error::Eof) => break,
                Err(err) => return Err(anyhow!(err)).context("failed to pull filtered frame"),
            }
        }

        Ok(())
    }

    fn receive_and_process_decoded_frames(
        &mut self,
        octx: &mut format::context::Output,
        ost_time_base: Rational,
    ) -> Result<()> {
        let mut decoded = frame::Video::empty();

        loop {
            match self.decoder.receive_frame(&mut decoded) {
                Ok(()) => {
                    let timestamp = decoded.timestamp();
                    decoded.set_pts(timestamp);
                    self.add_frame_to_filter(&decoded)?;
                    self.get_and_process_filtered_frames(octx, ost_time_base)?;
                }
                Err(ffmpeg::Error::Other { errno }) if errno == ffmpeg::util::error::EAGAIN => {
                    break
                }
                Err(ffmpeg::Error::Eof) => break,
                Err(err) => return Err(anyhow!(err)).context("failed to decode frame"),
            }
        }

        Ok(())
    }

    fn receive_and_process_encoded_packets(
        &mut self,
        octx: &mut format::context::Output,
        ost_time_base: Rational,
    ) -> Result<()> {
        let mut encoded = Packet::empty();

        loop {
            match self.encoder.receive_packet(&mut encoded) {
                Ok(()) => {
                    encoded.set_stream(self.ost_index);
                    encoded.rescale_ts(self.input_time_base, ost_time_base);
                    encoded
                        .write_interleaved(octx)
                        .context("failed to write encoded packet")?;
                }
                Err(ffmpeg::Error::Other { errno }) if errno == ffmpeg::util::error::EAGAIN => {
                    break
                }
                Err(ffmpeg::Error::Eof) => break,
                Err(err) => {
                    return Err(anyhow!(err)).context("failed to receive encoded packet")
                }
            }
        }

        Ok(())
    }
}

#[cfg(feature = "ffmpeg")]
fn build_video_filter(
    filter_spec: &str,
    decoder: &ffmpeg::decoder::Video,
    encoder: &ffmpeg::encoder::Video,
) -> Result<filter::Graph> {
    let mut graph = filter::Graph::new();

    let pixel_format = decoder
        .format()
        .descriptor()
        .ok_or_else(|| anyhow!("missing pixel format descriptor"))?
        .name();

    let args = format!(
        "video_size={}x{}:pix_fmt={}:time_base={}:pixel_aspect={}",
        decoder.width(),
        decoder.height(),
        pixel_format,
        decoder.time_base(),
        decoder.aspect_ratio(),
    );

    graph
        .add(
            &filter::find("buffer").ok_or_else(|| anyhow!("buffer filter not available"))?,
            "in",
            &args,
        )
        .context("failed to create filter input")?;
    graph
        .add(
            &filter::find("buffersink")
                .ok_or_else(|| anyhow!("buffersink filter not available"))?,
            "out",
            "",
        )
        .context("failed to create filter output")?;

    {
        let mut out = graph
            .get("out")
            .ok_or_else(|| anyhow!("missing filter output context"))?;
        out.set_pixel_format(encoder.format());
    }

    graph
        .output("in", 0)?
        .input("out", 0)?
        .parse(filter_spec)
        .context("failed to parse video filter graph")?;
    graph
        .validate()
        .context("failed to validate video filter graph")?;

    Ok(graph)
}

#[cfg(feature = "ffmpeg")]
fn build_filter_spec(watermark: bool, add_captions: bool, script: Option<&str>) -> String {
    let mut filters = vec![
        "scale=1080:1920:force_original_aspect_ratio=decrease".to_string(),
        "pad=1080:1920:(ow-iw)/2:(oh-ih)/2:color=black".to_string(),
    ];

    if watermark {
        filters.push(
            "drawtext=text='Qvora':fontsize=48:fontcolor=white@0.1:x=w-200:y=h-100:fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"
                .to_string(),
        );
    }

    if add_captions {
        if let Some(script) = script.and_then(normalize_caption_text) {
            filters.push(format!(
                "drawtext=text='{}':fontsize=42:fontcolor=white:borderw=3:bordercolor=black@0.75:line_spacing=12:x=(w-text_w)/2:y=h-text_h-140:fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
                escape_drawtext(&script)
            ));
        }
    }

    filters.join(",")
}

#[cfg(feature = "ffmpeg")]
fn normalize_caption_text(script: &str) -> Option<String> {
    let collapsed = script
        .split_whitespace()
        .collect::<Vec<_>>()
        .chunks(6)
        .map(|chunk| chunk.join(" "))
        .collect::<Vec<_>>()
        .join("\n");

    let trimmed = collapsed.trim();
    if trimmed.is_empty() {
        None
    } else {
        Some(trimmed.to_string())
    }
}

#[cfg(feature = "ffmpeg")]
fn escape_drawtext(value: &str) -> String {
    value
        .replace('\\', "\\\\")
        .replace(':', "\\:")
        .replace('\'', "\\'")
        .replace('%', "\\%")
        .replace(',', "\\,")
        .replace('\n', "\\n")
}

#[cfg(feature = "ffmpeg")]
fn init_ffmpeg() -> Result<()> {
    if FFMPEG_INIT.get().is_none() {
        ffmpeg::init()
            .map_err(|err| anyhow!(err))
            .context("failed to initialize ffmpeg")?;
        let _ = FFMPEG_INIT.set(());
    }

    Ok(())
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
