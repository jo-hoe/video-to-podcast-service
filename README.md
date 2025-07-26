# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/video-to-podcast-service)](https://goreportcard.com/report/github.com/jo-hoe/video-to-podcast-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/video-to-podcast-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/video-to-podcast-service?branch=main)

Video To Podcast Service is a backend service that downloads video files (currently only from YouTube), extracts and converts them into audio files, and organizes them into podcast feeds accessible via RSS. The service exposes a REST API for adding new videos, listing available podcast feeds, retrieving audio files, and deleting podcast items.

**Key Features:**

- Download videos and convert to audio (MP3)
- Organize audio files into podcast feeds (one feed per channel)
- Serve podcast feeds as RSS for use in podcast apps
- REST API for adding, listing, retrieving, and deleting podcast items
- Dockerized for easy deployment

## How to Use

### Start the Service

You can start the service using `make` (recommended if you have the repository):

```bash
make start
```

#### If you do not have the repository or Docker image

You can pull the latest prebuilt image from GitHub Container Registry:

```bash
docker pull ghcr.io/jo-hoe/video-to-podcast-service:latest
```

#### Running the Service with Docker

```bash
docker run --rm -p 8080:8080 ghcr.io/jo-hoe/video-to-podcast-service:latest
```

#### Or with Docker Compose

If you have a `docker-compose.yml` referencing the image, you can use:

```bash
make start-service
# or
make start-service-rebuild
```

### Environment Variables

The service supports the following environment variables. **All are optional**—if not set, sensible defaults are used:

- `BASE_PATH` (optional): Sets the base directory for resources. **Default:** `resources` directory next to the executable inside the container. Only set this if you want to use a custom location or mount a host directory.
- `CONNECTION_STRING` (optional): Sets the database connection string. **Default:** empty string, which uses a SQLite database file in the resource path. Only set this if you want to use a custom database or location.
- `BASE_URL` (optional): Sets the base URL for generating podcast feed and audio file links. **Default:** Uses the `Host` header from incoming requests with `http://` scheme. Set this to override automatic URL detection, useful for reverse proxies or custom domains. Supports both HTTP and HTTPS schemes (e.g., `https://example.com` or `http://localhost:8080`). If no scheme is provided, defaults to HTTPS.
- `YTDLP_COOKIES_FILE` (optional): Path to a Netscape-format cookie file, e.g. for accessing age-restricted or private YouTube content.

**How to get cookies file:**

For detailed instructions on obtaining cookies (including permanent cookies that don't expire), see the official yt-dlp documentation:
- [YouTube extractor documentation](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#youtube)
- [How do I pass cookies to yt-dlp?](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp)

**Usage with Docker:**

Once you have your cookie file, mount it into the container:

```bash
docker run --rm -p 8080:8080 \
  -v /path/to/your/youtube_cookies.txt:/app/.cookies/youtube_cookies.txt \
  -e YTDLP_COOKIES_FILE=/app/.cookies/youtube_cookies.txt \
  -e BASE_PATH=/data/resources \
  -e CONNECTION_STRING="file:/data/resources/podcast.db" \
  -e BASE_URL="https://your-domain.com" \
  ghcr.io/jo-hoe/video-to-podcast-service:latest
```

**Usage with Docker Compose:**

If using the provided `docker-compose.yml`, place your `youtube_cookies.txt` file in the project root directory and uncomment the relevant lines in `docker-compose.yml`:

```yaml
volumes:
  - "./youtube_cookies.txt:/app/.cookies/youtube_cookies.txt"
environment:
  YTDLP_COOKIES_FILE: "/app/.cookies/youtube_cookies.txt"
```

**Note:** The service will work without cookies for most public YouTube content and in case you are not using a blocked IP (e.g. from the IP range of a hyperscaler).

or on PowerShell:

```powershell
docker run --rm -p 8080:8080 `
  -e YTDLP_COOKIES_FILE=/app/.cookies/youtube_cookies.txt `
  -e BASE_PATH=/data/resources `
  -e CONNECTION_STRING="file:/data/resources/podcast.db" `
  -e BASE_URL="https://your-domain.com" `
  ghcr.io/jo-hoe/video-to-podcast-service:latest
```

Or with Docker Compose, set them in your `docker-compose.yml` under `environment:`.

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

