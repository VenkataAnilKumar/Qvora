FROM rust:1.87-slim AS builder

WORKDIR /app

# Install build dependencies + ffmpeg dev libs
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libavcodec-dev \
    libavformat-dev \
    libavfilter-dev \
    libavutil-dev \
    libswscale-dev \
    libswresample-dev \
    libavdevice-dev \
    clang \
    && rm -rf /var/lib/apt/lists/*

# Cache dependencies
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release --features ffmpeg
RUN rm src/main.rs

COPY src ./src
RUN touch src/main.rs
RUN cargo build --release --features ffmpeg

# ---- Final image ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libavcodec60 \
    libavformat60 \
    libavfilter9 \
    libavutil58 \
    libswscale7 \
    libswresample4 \
    libx264-164 \
    libfdk-aac2 \
    ffmpeg \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/postprocess /usr/local/bin/postprocess

EXPOSE 3001
ENTRYPOINT ["postprocess"]
