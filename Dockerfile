# Build stage: build the Go application against musl (Alpine) for CGO compatibility
FROM golang:1.26.4-alpine3.23 AS build

# Set the working directory
WORKDIR /src

# Copy go.mod and go.sum to leverage caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install build dependencies for CGO on Alpine (musl)
RUN apk add --no-cache build-base sqlite-dev

# Copy the rest of the code and build the application
COPY . ./

# Build with CGO (required for sqlite dependency) and output to a specific location
RUN CGO_ENABLED=1 go build -o /out/app -ldflags="-s -w" ./

# Runtime stage: use Alpine directly for a minimal image
FROM alpine:3.24

# Install required runtime dependencies:
# - ca-certificates: TLS trust store
# - ffmpeg: audio/video conversion
# - python3: runtime for yt-dlp zipapp
# - deno: JS runtime required by yt-dlp to solve YouTube's n-challenge (>= 2.3.0)
# - sqlite-libs: runtime SQLite shared library (CGO dependency)
# - wget: to fetch yt-dlp standalone binary
RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    python3 \
    deno \
    sqlite-libs \
    wget && \
    update-ca-certificates && \
    wget https://github.com/yt-dlp/yt-dlp/releases/download/2026.06.09/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp && \
    # Verify installations
    deno --version && \
    yt-dlp --version && \
    # Cleanup
    apk del wget && \
    rm -rf /root/.cache /var/cache/apk/*

# Create a non-root user and set up directories (BusyBox adduser)
RUN adduser -D -h /home/appuser appuser && \
    mkdir -p /home/appuser/app/resources && \
    mkdir -p /home/appuser/.cache/deno && \
    chown -R appuser:appuser /home/appuser && \
    chmod 755 /home/appuser/.cache/deno

# Set the working directory
WORKDIR /home/appuser/app

# Copy the built application from the build stage
COPY --from=build /out/app ./app

# Ensure the executable has execute permissions and set ownership
RUN chmod +x ./app && \
    chown -R appuser:appuser /home/appuser/app

# Set the user to the non-root user
USER appuser

# Set HOME environment variable explicitly
ENV HOME=/home/appuser

# ENTRYPOINT should point to the executable
ENTRYPOINT ["./app"]
