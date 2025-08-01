# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/video-to-podcast-service)](https://goreportcard.com/report/github.com/jo-hoe/video-to-podcast-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/video-to-podcast-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/video-to-podcast-service?branch=main)

Video To Podcast Service is a microservices-based application that downloads video files (currently only from YouTube), extracts and converts them into audio files, and organizes them into podcast feeds accessible via RSS. The system consists of two separate services: an API service that handles all backend operations and a UI service that provides a web interface for users.

**Key Features:**

- Download videos and convert to audio (MP3)
- Organize audio files into podcast feeds (one feed per channel)
- Serve podcast feeds as RSS for use in podcast apps
- REST API for adding, listing, retrieving, and deleting podcast items
- Web interface for easy interaction with the system
- Microservices architecture for better scalability and maintainability
- Dockerized services with Docker Compose orchestration

## Architecture

The application follows a microservices architecture with two main components:

### API Service (Port 8080)
- Handles all backend operations including video downloading and conversion
- Manages the SQLite database and file storage
- Serves REST API endpoints for programmatic access
- Serves RSS feeds and audio files directly
- Contains all business logic for video processing

### UI Service (Port 3000)
- Provides a web interface for users to interact with the system
- Handles HTMX requests for dynamic content updates
- Communicates with the API service via HTTP requests
- Focuses solely on presentation logic and user experience

### Service Communication
- **UI → API**: HTTP requests over Docker network using service names
- **External Access**: Users can access both the web interface (UI service) and REST API (API service) directly
- **Network**: Both services run in the same Docker network for internal communication
- **Data Flow**: User → UI Service → API Service → Database/File System

## How to Use

### Running with Docker Compose (Recommended)

The microservices architecture is designed to run with Docker Compose, which orchestrates both the API and UI services:

```bash
# Start both services
docker-compose up

# Or run in detached mode
docker-compose up -d

# Build and start (if you've made changes)
docker-compose up --build
```

This will start:
- **API Service** on `http://localhost:8080` - REST API endpoints
- **UI Service** on `http://localhost:3000` - Web interface

### Using Make Commands

If you have the repository, you can use the provided Make targets:

```bash
# Start both services
make start

# Start services and rebuild if needed
make start-service-rebuild
```

### Accessing the Services

Once running, you can access:

- **Web Interface**: `http://localhost:3000` - User-friendly web interface
- **REST API**: `http://localhost:8080` - Direct API access for programmatic use
- **API Documentation**: See [`openapi.yaml`](./openapi.yaml) for full API specification

### Service Dependencies

The Docker Compose configuration ensures proper startup order:
- UI service waits for API service to be ready
- Both services share a Docker network for internal communication
- External access is available through published ports

### Environment Variables

The services support the following environment variables. **All are optional**—if not set, sensible defaults are used:

#### API Service Environment Variables
- `PORT` (optional): API service port. **Default:** `8080`
- `BASE_PATH` (optional): Sets the base directory for resources. **Default:** `resources` directory next to the executable inside the container. Only set this if you want to use a custom location or mount a host directory.
- `CONNECTION_STRING` (optional): Sets the database connection string. **Default:** empty string, which uses a SQLite database file in the resource path. Only set this if you want to use a custom database or location.
- `BASE_URL` (optional): Sets the base URL for generating podcast feed and audio file links. **Default:** Uses the `Host` header from incoming requests with `http://` scheme. Set this to override automatic URL detection, useful for reverse proxies or custom domains. Supports both HTTP and HTTPS schemes (e.g., `https://example.com` or `http://localhost:8080`). If no scheme is provided, defaults to HTTPS.
- `YTDLP_COOKIES_FILE` (optional): Path to a Netscape-format cookie file, e.g. for accessing age-restricted or private YouTube content.

#### UI Service Environment Variables
- `UI_PORT` (optional): UI service port. **Default:** `3000`
- `API_BASE_URL` (optional): Base URL for API service communication. **Default:** `http://api-service:8080` (uses Docker service name)
- `API_TIMEOUT` (optional): Timeout for API requests. **Default:** `30s`

**How to get cookies file:**

For detailed instructions on obtaining cookies (including permanent cookies that don't expire), see the official yt-dlp documentation:

- [YouTube extractor documentation](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#youtube)
- [How do I pass cookies to yt-dlp?](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp)

**Note:** The service will work without cookies for most public YouTube content and in case you are not using a blocked IP (e.g. from the IP range of a hyperscaler).

**Usage with Docker Compose:**

If using the provided `docker-compose.yml`, place your `youtube_cookies.txt` file in the project root directory and configure it in the API service section:

```yaml
services:
  api-service:
    build:
      context: .
      dockerfile: Dockerfile.api
    ports:
      - "8080:8080"
    volumes:
      - "./data:/app/resources"
      - "./youtube_cookies.txt:/app/.cookies/youtube_cookies.txt"  # Add this line
    environment:
      PORT: "8080"
      BASE_PATH: "/app/resources"
      YTDLP_COOKIES_FILE: "/app/.cookies/youtube_cookies.txt"  # Add this line
      BASE_URL: "https://your-domain.com"  # Optional: set your domain
    
  ui-service:
    build:
      context: .
      dockerfile: Dockerfile.ui
    ports:
      - "3000:3000"
    environment:
      UI_PORT: "3000"
      API_BASE_URL: "http://api-service:8080"
      API_TIMEOUT: "30s"
    depends_on:
      - api-service
```

### Resources

All downloaded resources are placed in the `resources` directory. Podcasts are organized in subdirectories named after the channel the video belongs to. Each feed has its own directory containing audio files and the RSS XML.

### Host URL and Network Access

The service generates podcast feed URLs and audio file links in two ways:

1. **Using BASE_URL environment variable (recommended)**: Set the `BASE_URL` environment variable to explicitly define how URLs should be generated. This is the preferred method for production deployments, reverse proxies, or when you need consistent URLs regardless of how clients access the service.

2. **Using Host header (fallback)**: If `BASE_URL` is not set, the service uses the `Host` header from incoming HTTP requests with an `http://` scheme.

**BASE_URL Examples:**

- `BASE_URL=https://podcasts.example.com` - All generated URLs will use this domain with HTTPS
- `BASE_URL=http://192.168.1.100:8080` - All generated URLs will use this IP and port with HTTP
- `BASE_URL=my-domain.com` - Defaults to HTTPS, equivalent to `https://my-domain.com`

**Host Header Behavior (when BASE_URL is not set):**

- Accessing `http://localhost:8080/v1/feeds` generates feed URLs with `http://localhost:8080`
- Accessing `http://192.168.1.100:8080/v1/feeds` generates feed URLs with `http://192.168.1.100:8080`

**Why does this matter?**

- If you want to subscribe to podcast feeds from other devices, the generated URLs must be accessible from those devices
- Using `BASE_URL` ensures consistent URLs regardless of how clients access the API
- When running behind a reverse proxy or load balancer, `BASE_URL` should match your public domain
- When running in Docker, make sure to publish the port (e.g., `-p 8080:8080`) if not using `BASE_URL`

**Recommendation:** Always set `BASE_URL` in production environments to ensure reliable, consistent URLs for podcast clients.

## Deployment

### Development Deployment

For development, use Docker Compose to run both services locally:

```bash
# Clone the repository
git clone <repository-url>
cd video-to-podcast-service

# Start both services
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
```

### Production Deployment

For production deployment, you have several options:

#### Option 1: Docker Compose (Simple Production)

1. **Prepare your environment:**
   ```bash
   # Create data directory for persistent storage
   mkdir -p ./data
   
   # Copy your cookies file if needed
   cp /path/to/your/youtube_cookies.txt ./youtube_cookies.txt
   ```

2. **Configure environment variables:**
   Create a `.env` file or modify `docker-compose.yml` with production settings:
   ```bash
   # Set your domain for proper URL generation
   BASE_URL=https://your-domain.com
   
   # Optional: Configure custom paths
   BASE_PATH=/app/resources
   CONNECTION_STRING=file:/app/resources/podcast.db
   ```

3. **Deploy:**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

#### Option 2: Container Orchestration (Kubernetes, Docker Swarm)

For larger deployments, you can use container orchestration platforms:

1. **Build and push images:**
   ```bash
   # Build API service
   docker build -f Dockerfile.api -t your-registry/video-podcast-api:latest .
   
   # Build UI service  
   docker build -f Dockerfile.ui -t your-registry/video-podcast-ui:latest .
   
   # Push to your container registry
   docker push your-registry/video-podcast-api:latest
   docker push your-registry/video-podcast-ui:latest
   ```

2. **Deploy using your orchestration platform** with appropriate service definitions, load balancers, and persistent volumes.

#### Option 3: Reverse Proxy Setup

When deploying behind a reverse proxy (nginx, traefik, etc.):

1. **Configure your reverse proxy** to route traffic:
   - `/` → UI Service (port 3000)
   - `/v1/*` → API Service (port 8080)
   - `/feeds/*` → API Service (port 8080)

2. **Set BASE_URL** to match your public domain:
   ```yaml
   environment:
     BASE_URL: "https://your-domain.com"
   ```

### Port Configuration

The microservices use the following default ports:

- **API Service**: `8080` (configurable via `PORT` environment variable)
- **UI Service**: `3000` (configurable via `UI_PORT` environment variable)

**Important:** When deploying, ensure these ports are properly mapped and accessible according to your network setup.

### Health Checks and Monitoring

Both services include health check endpoints:

- **API Service**: `GET /v1/health`
- **UI Service**: Health checks via Docker Compose configuration

Configure your monitoring and alerting systems to check these endpoints for service availability.

## API Usage

The service exposes a REST API. See [`openapi.yaml`](./openapi.yaml) for the full OpenAPI/Swagger specification.

## Linting

The project uses `golangci-lint` for linting. See <https://golangci-lint.run/usage/install/> for installation instructions.

To run linting locally:

```bash
golangci-lint run ./...
```

## Limitations

- Only YouTube is supported as a video source.
- Google may block certain IPs (e.g., from cloud providers), resulting in errors like `403` or age restriction issues. See [this GitHub issue](https://github.com/kkdai/youtube/issues/343#issuecomment-2347950479) for more details.
