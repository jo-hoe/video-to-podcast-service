# Video To Podcast Service

[![Test Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/test/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/video-to-podcast-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/video-to-podcast-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/video-to-podcast-service)](https://goreportcard.com/report/github.com/jo-hoe/video-to-podcast-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/video-to-podcast-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/video-to-podcast-service?branch=main)

Convert YouTube videos to podcast feeds that you can subscribe to in any podcast app.

## What It Does

- Downloads YouTube videos and converts them to audio (MP3)
- Creates podcast feeds organized by channel
- Serves RSS feeds that work with any podcast app
- Provides a simple web interface for adding videos

## Quick Start

1. **Download and run:**

   ```bash
   git clone https://github.com/jo-hoe/video-to-podcast-service.git
   cd video-to-podcast-service
   
   # Create required directories and set permissions (Linux/Mac)
   mkdir -p cookies data/{cookies,resources,database,config} resources
   docker compose up
   ```

2. **Access the service:**
   - Web interface: <http://localhost:3000>
   - API: <http://localhost:8080>

3. **Add a video:**
   - Open <http://localhost:3000> in your browser
   - Paste a YouTube URL and click "Add"
   - The video will be downloaded and converted to audio

4. **Subscribe to feeds:**
   - Copy the RSS feed URL from the web interface
   - Add it to your favorite podcast app

## How It Works

The service runs two components:

- **API Service** (port 8080): Downloads videos, converts audio, serves RSS feeds
- **UI Service** (port 3000): Web interface for easy interaction

## Configuration

The service uses YAML configuration files with sensible defaults. No setup required for basic usage.

### Default Behavior

- **No persistence**: Data is stored temporarily and cleared when containers restart
- **In-memory database**: No permanent storage of metadata
- **Temporary files**: Audio files stored in `/tmp` and cleaned up automatically

### Custom Configuration

Create or modify `config.yaml` to customize settings:

```yaml
api:
  server:
    port: "8080"
  database:
    connection_string: ":memory:"  # In-memory (no persistence)
  storage:
    base_path: "/tmp/video-to-podcast"  # Temporary storage
  external:
    ytdlp_cookies_file: ""  # Optional: path to cookies file

ui:
  server:
    port: "3000"
  api:
    base_url: "http://localhost:8080"
    timeout: "30s"
```

### Adding Persistence (Optional)

To keep data between restarts, modify the configuration:

```yaml
api:
  database:
    connection_string: "/app/data/podcast.db"
  storage:
    base_path: "/app/data/resources"
```

And add volumes to `docker-compose.yml`:

```yaml
volumes:
  - podcast_data:/app/data
```

### YouTube Cookies (Optional)

For private or age-restricted videos, add a cookies file:

1. Export cookies from your browser to `cookies/youtube_cookies.txt`
2. Update configuration:

   ```yaml
   api:
     external:
       ytdlp_cookies_file: "/app/cookies/youtube_cookies.txt"
   ```

3. Mount the cookies file:

   ```yaml
   volumes:
     - "./cookies/youtube_cookies.txt:/app/cookies/youtube_cookies.txt:ro"
   ```

See [yt-dlp documentation](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp) for cookie extraction instructions.

## API Usage

The service provides a REST API for programmatic access:

- `GET /v1/feeds` - List all podcast feeds
- `POST /v1/addItems` - Add videos to convert
- `GET /v1/items` - List all podcast items
- `DELETE /v1/feeds/{feed}/{item}` - Delete a podcast item
- `GET /v1/feeds/{feed}/rss.xml` - Get RSS feed for a channel

## Accessing from Other Devices

To access podcast feeds from other devices on your network:

1. **Find your computer's IP address:**

   ```bash
   # On Linux/Mac
   ip addr show | grep inet
   
   # On Windows
   ipconfig
   ```

2. **Update the configuration** to use your IP:

   ```yaml
   api:
     server:
       base_url: "http://192.168.1.100:8080"  # Use your actual IP
   ```

3. **Subscribe to feeds** using URLs like:
   `http://192.168.1.100:8080/v1/feeds/ChannelName/rss.xml`

## Limitations

- Only YouTube videos are supported
- Some videos may be blocked due to geographic restrictions or IP blocking
- No persistence by default (data cleared on restart)

## Troubleshooting

**Service won't start:**

- Make sure ports 3000 and 8080 are available
- Check Docker is running: `docker --version`

**Permission denied errors (Linux):**

- Use Docker to fix permissions: `docker run --rm -v $(pwd):/workspace alpine chown -R 1001:1001 /workspace/cookies /workspace/data /workspace/resources`
- Or run with user mapping: `docker compose run --user $(id -u):$(id -g) api-service`

**Can't download videos:**

- Try adding YouTube cookies for authentication
- Some IPs (especially cloud providers) may be blocked by YouTube

**Can't access from other devices:**

- Configure `base_url` with your computer's IP address
- Make sure firewall allows connections on ports 3000 and 8080
