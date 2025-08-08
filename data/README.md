# Data Directory Structure

This directory contains all persistent data for the video-to-podcast service, organized into logical subdirectories for better management and easier backup/restore operations.

## Directory Structure

```
data/
├── resources/          # Audio files and media resources
├── database/          # SQLite database files
├── cookies/           # YouTube cookies for authentication
└── config/            # Configuration files
```

## Directory Descriptions

### `resources/`
- **Purpose**: Stores downloaded audio files and converted media resources
- **Contents**: MP3 files, temporary download files, processed audio content
- **Environment Variable**: `BASE_PATH=/app/data/resources`

### `database/`
- **Purpose**: Contains SQLite database files for persistent data storage
- **Contents**: `podcast_items.db` and other database files
- **Environment Variable**: `DATABASE_PATH=/app/data/database`

### `cookies/`
- **Purpose**: Stores YouTube authentication cookies for accessing private/restricted content
- **Contents**: `youtube_cookies.txt` and other authentication files
- **Environment Variable**: `YTDLP_COOKIES_FILE=/app/data/cookies/youtube_cookies.txt`

### `config/`
- **Purpose**: Reserved for future configuration files
- **Contents**: Application configuration files, user preferences, etc.

## Docker Volume Configuration

The entire `data/` directory is mounted as a single Docker volume named `podcast_data`:

```yaml
volumes:
  - podcast_data:/app/data
```

This approach provides:
- **Unified Management**: All persistent data in one location
- **Easy Backup**: Single volume to backup/restore
- **Better Organization**: Logical separation of different data types
- **Simplified Permissions**: Consistent ownership across all subdirectories

## Environment Variables

The following environment variables control the data directory paths:

- `BASE_PATH`: Path to resources directory (default: `/app/data/resources`)
- `DATABASE_PATH`: Path to database directory (default: `/app/data/database`)
- `YTDLP_COOKIES_FILE`: Path to YouTube cookies file (optional)

## Local Development

For local development, you can create this directory structure in your project root:

```bash
mkdir -p data/{resources,database,cookies,config}
```

The application will automatically detect and use this structure when running locally.

## Migration from Old Structure

If you're migrating from the old structure where resources were stored directly in the `resources/` directory:

1. **Backup your data**: Copy existing files to a safe location
2. **Update Docker Compose**: The new configuration uses `podcast_data:/app/data`
3. **Restart services**: Run `docker compose up -d --build`
4. **Restore data**: Copy your backed-up files to the appropriate subdirectories

The application will automatically create the directory structure on first run.