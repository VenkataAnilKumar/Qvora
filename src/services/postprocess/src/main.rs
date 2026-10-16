use axum::{Router, routing::{get, post}};
use axum::http::HeaderValue;
use std::net::SocketAddr;
use tower_http::cors::CorsLayer;
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

mod handler;
mod model;
mod processor;
mod error;
mod mux;

/// This service is only ever called server-to-server (the Go worker invoking
/// `/process`) — no browser is meant to hit it directly. CORS here is
/// defense-in-depth, not the primary access control (that's Railway internal
/// networking). Set `POSTPROCESS_ALLOWED_ORIGINS` to a comma-separated list to
/// restrict it explicitly; falling back to permissive when unset keeps local
/// dev/docker-compose zero-config, but logs a warning so the fallback is visible.
fn build_cors_layer() -> CorsLayer {
    match std::env::var("POSTPROCESS_ALLOWED_ORIGINS") {
        Ok(origins) if !origins.trim().is_empty() => {
            let parsed: Vec<HeaderValue> = origins
                .split(',')
                .map(str::trim)
                .filter(|o| !o.is_empty())
                .filter_map(|o| o.parse::<HeaderValue>().ok())
                .collect();
            CorsLayer::new().allow_origin(parsed)
        }
        _ => {
            tracing::warn!(
                "POSTPROCESS_ALLOWED_ORIGINS not set — falling back to permissive CORS (dev only; set this in prod)"
            );
            CorsLayer::permissive()
        }
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Structured JSON logging
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "postprocess=info,tower_http=info".into()))
        .with(tracing_subscriber::fmt::layer().json())
        .init();

    let app = Router::new()
        .route("/health", get(handler::health))
        .route("/process", post(handler::process))
        .layer(build_cors_layer())
        .layer(TraceLayer::new_for_http());

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(3001);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("qvora postprocessor listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
