# Build stage: build the Go application
FROM golang:1.24.5 AS build
# Set the working directory
WORKDIR /app
# Copy go.mod and go.sum to leverage caching for dependencies
COPY go.mod go.sum ./
RUN go mod download
# Install build dependencies for CGO
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*
# Copy the rest of the code and build the application
COPY . ./
# Build with CGO (required for sqlite dependency) and output to a specific location
RUN CGO_ENABLED=1 go build -o /go/bin/app ./

# Runtime stage: use a minimal base image for runtime
FROM jrottenberg/ffmpeg:7.1-ubuntu

# Install required dependencies
RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    wget \
    python3-minimal && \
    update-ca-certificates && \
    # Download yt-dlp binary
    wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp && \
    # Verify yt-dlp works
    /usr/local/bin/yt-dlp --version && \
    # Cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p /app/resources

# Copy the built application from the build stage
COPY --from=build /go/bin/app .

# Ensure the executable has execute permissions
RUN chmod +x /app/app

# ENTRYPOINT should point to the executable
ENTRYPOINT ["./app"]