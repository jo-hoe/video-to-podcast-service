# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/video-to-podcast-service)](https://goreportcard.com/report/github.com/jo-hoe/video-to-podcast-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/video-to-podcast-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/video-to-podcast-service?branch=main)

Video To Podcast Service is a backend service that downloads video files (currently only from YouTube), extracts and converts them into audio files, and organizes them into podcast feeds accessible via RSS. The service exposes a REST API for adding new videos, listing available podcast feeds, retrieving audio files, and deleting podcast items.

**Key Features:**

- Download videos (YouTube supported) and convert to audio (MP3)
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

#### YouTube Cookie Configuration

For accessing age-restricted or private YouTube content, you can provide cookies to yt-dlp:

- `YTDLP_COOKIES_FILE` (optional): Path to a Netscape-format cookie file.

**How to get cookies:**

For detailed instructions on obtaining cookies (including permanent cookies that don't expire), see the official yt-dlp documentation:
- [YouTube extractor documentation](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#youtube)
- [How do I pass cookies to yt-dlp?](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp)

**Usage with Docker:**

Once you have your cookie file, mount it into the container:

```bash
docker run --rm -p 8080:8080 \
  -v /path/to/your/youtube_cookies.txt:/home/appuser/.cookies/youtube_cookies.txt \
  -e YTDLP_COOKIES_FILE=/home/appuser/.cookies/youtube_cookies.txt \
  -e BASE_PATH=/data/resources \
  -e CONNECTION_STRING="file:/data/resources/podcast.db" \
  ghcr.io/jo-hoe/video-to-podcast-service:latest
```

**Note:** The service will work without cookies for most public YouTube content and in case you are not using a blocked IP (e.g. from the IP range of a hyperscaler).

or on PowerShell:

```powershell
docker run --rm -p 8080:8080 `
  -e YTDLP_COOKIES_FILE=/home/appuser/.cookies/youtube_cookies.txt `
  -e BASE_PATH=/data/resources `
  -e CONNECTION_STRING="file:/data/resources/podcast.db" `
  ghcr.io/jo-hoe/video-to-podcast-service:latest
```

Or with Docker Compose, set them in your `docker-compose.yml` under `environment:`.

### Resources

All downloaded resources are placed in the `resources` directory. Podcasts are organized in subdirectories named after the channel the video belongs to. Each feed has its own directory containing audio files and the RSS XML.

### Host URL and Network Access

The service generates podcast feed URLs and audio file links using the `Host` header from incoming HTTP requests. This means:

- If you access the API or RSS feed using `localhost` (e.g., `http://localhost:8080/v1/feeds`), the generated feed and audio URLs will also use `localhost`.
- If you access the API using your machine's external IP or hostname (e.g., `http://192.168.1.100:8080/v1/feeds`), the generated URLs will use that IP or hostname.

**Why does this matter?**

- If you want to subscribe to the podcast feed from another device (e.g., your phone or another computer), you must use the external IP or hostname in the URL, not `localhost`. Otherwise, the generated links in the RSS feed will not be accessible from other devices.
- When running in Docker, make sure to publish the port (e.g., `-p 8080:8080`) and use your host's IP address to access the service externally.

**Example:**

- Accessing `http://localhost:8080/v1/feeds` from your browser on the same machine will generate feed URLs with `localhost`.
- Accessing `http://192.168.1.100:8080/v1/feeds` from another device on your network will generate feed URLs with `192.168.1.100`.

If you want to share feeds or audio links, always use the address that matches how other devices will connect to your server.

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

## Future Work

- Add iTunes to
  - add tags for podcast images (requires custom XML generation)
  - set length in podcast metadata
- Provide ticketing/progress feedback via API
- Auto-chapterize videos without chapters

## Relevant Links

- [ID3 Tags](https://www.exiftool.org/TagNames/ID3.html)
- [Example podcast](https://feeds.libsyn.com/230510/rss)
