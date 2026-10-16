# Qvora Postprocessor

Rust/Axum service — watermark, captions, transcode, and reframe for generated video variants. CPU-bound work only; see root `README.md` for how this fits into the pipeline.

## The `ffmpeg` Cargo feature

`ffmpeg-next` is an **optional** dependency (see `Cargo.toml`). The Dockerfile builds with `--features ffmpeg`, which is required for `/process` to actually do anything.

```bash
cargo build --features ffmpeg   # required — matches the Docker build
cargo build                     # compiles, but /process always returns an error
```

Without the feature flag, `POST /process` immediately responds with:

```
ffmpeg feature not enabled; rebuild with --features ffmpeg in Docker
```

This is silent at compile time — a plain `cargo build`/`cargo run` succeeds, and the failure only shows up when you actually call `/process`. The feature is optional specifically so the crate still builds on machines without the system ffmpeg libraries available (e.g. local Windows development without the Docker toolchain) — but if you're testing postprocessing locally, always build with `--features ffmpeg`.

## Environment variables

| Variable | Required | Notes |
|---|---|---|
| `R2_BUCKET`, `R2_ENDPOINT` | Yes | Cloudflare R2 (S3-compatible) |
| `MUX_ACCESS_TOKEN`, `MUX_SECRET_TOKEN` | Yes | Mux asset upload |
| `API_BASE_URL`, `INTERNAL_API_KEY` | Yes | Callback to the Go API on completion |
| `POSTPROCESS_ALLOWED_ORIGINS` | No | Comma-separated CORS allowlist. This service is only ever called server-to-server (the Go worker), never from a browser — unset falls back to permissive CORS for zero-config local/dev, but set this in production as defense-in-depth beyond Railway's internal networking. |
| `PORT` | No | Defaults to `3001` |
