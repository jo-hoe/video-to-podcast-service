# Configuration

The Video to Podcast Service uses YAML-based configuration for easy management and deployment flexibility.

## Configuration Files

### Default Configuration (`config.yaml`)
The default configuration provides sensible defaults for local development:

```yaml
api:
  server:
    port: "8080"
    base_url: "" # Auto-detect based on request
  database:
    connection_string: ":memory:" # In-memory database (no persistence)
  storage:
    base_path: "/tmp/video-to-podcast" # Temporary storage
  external:
    ytdlp_cookies_file: "" # No cookies file by default

ui:
  server:
    port: "3000"
  api:
    base_url: "http://localhost:8080" # For local development
    timeout: "30s"
```

### Docker Configuration (`config.docker.yaml`)
Optimized for containerized deployment with Docker Compose:

```yaml
api:
  server:
    port: "8080"
    base_url: ""
  database:
    connection_string: ":memory:" # No persistence beyond container lifetime
  storage:
    base_path: "/tmp/video-to-podcast"
  external:
    ytdlp_cookies_file: ""

ui:
  server:
    port: "3000"
  api:
    base_url: "http://api-service:8080" # Docker Compose service name
    timeout: "30s"
```

## Configuration Options

### API Service Configuration

#### Server Settings
- `api.server.port`: Port for the API service (default: "8080")
- `api.server.base_url`: Base URL for the API service (empty = auto-detect)

#### Database Settings
- `api.database.connection_string`: Database connection string
  - `:memory:` - In-memory database (no persistence)
  - `path/to/database.db` - SQLite file database
  - Empty string - Uses default location

#### Storage Settings
- `api.storage.base_path`: Directory for storing downloaded audio files
  - `/tmp/video-to-podcast` - Temporary storage (default)
  - `/app/data/resources` - Persistent storage in containers

#### External Services
- `api.external.ytdlp_cookies_file`: Path to YouTube cookies file (optional)
  - Empty string - No cookies (default)
  - `/path/to/cookies.txt` - Use cookies for authentication

### UI Service Configuration

#### Server Settings
- `ui.server.port`: Port for the UI service (default: "3000")

#### API Client Settings
- `ui.api.base_url`: URL of the API service
  - `http://localhost:8080` - Local development
  - `http://api-service:8080` - Docker Compose
- `ui.api.timeout`: Timeout for API requests (default: "30s")

## Usage

### Local Development
1. The service automatically creates `config.yaml` with defaults if it doesn't exist
2. Modify `config.yaml` to customize settings
3. Run services: `go run cmd/api/main.go` and `go run cmd/ui/main.go`

### Docker Compose
1. Use the provided `config.docker.yaml` for containerized deployment
2. Customize settings in `config.docker.yaml` if needed
3. Run: `docker compose up`

### Custom Configuration
Create your own configuration file and specify its path:
```bash
# Set custom config path (optional)
export CONFIG_PATH="/path/to/custom-config.yaml"
go run cmd/api/main.go
```

## Configuration Behavior

### No Persistence by Default
The default configuration uses:
- In-memory database (`:memory:`)
- Temporary storage (`/tmp/video-to-podcast`)

This means:
- ✅ Works out of the box without setup
- ✅ No cleanup required
- ⚠️ Data is lost when services restart

### Adding Persistence
To enable persistence, modify the configuration:

```yaml
api:
  database:
    connection_string: "/app/data/database/podcast_items.db"
  storage:
    base_path: "/app/data/resources"
```

And mount volumes in Docker Compose:
```yaml
volumes:
  - podcast_data:/app/data
```

### YouTube Cookies
To access private or age-restricted videos, add a cookies file:

1. Export cookies from your browser to `cookies/youtube_cookies.txt`
2. Update configuration:
   ```yaml
   api:
     external:
       ytdlp_cookies_file: "/app/cookies/youtube_cookies.txt"
   ```
3. Mount the cookies file in Docker Compose:
   ```yaml
   volumes:
     - "./cookies/youtube_cookies.txt:/app/cookies/youtube_cookies.txt:ro"
   ```

## Environment Variables (Legacy)
The service no longer uses environment variables for configuration. All settings are now managed through YAML files for better organization and maintainability.