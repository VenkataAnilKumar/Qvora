use axum::{Router, routing::{get, post}};
use std::net::SocketAddr;
use tower_http::cors::CorsLayer;
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

mod handler;
mod model;
mod processor;
mod error;
mod mux;

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
        .layer(CorsLayer::permissive()) // restricted in prod via Railway internal networking
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
