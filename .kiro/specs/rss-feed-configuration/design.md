# Design Document

## Overview

This design implements configurable RSS feed generation modes for the video-to-podcast service. The system will support two modes:
1. **Per-directory mode** (default): Creates one RSS feed per subdirectory, maintaining current behavior
2. **Unified mode**: Creates a single RSS feed containing all podcast items regardless of subdirectory

The design extends the existing configuration system and modifies the feed service to support both modes while maintaining backward compatibility.

## Architecture

### Configuration Layer
- Extend the existing `APIConfig` struct to include RSS feed configuration
- Add new `FeedConfig` struct with mode selection
- Update configuration loading to handle the new feed settings
- Maintain backward compatibility by defaulting to per-directory mode

### Feed Service Layer
- Modify `FeedService` to accept and use feed configuration
- Implement mode-specific feed generation logic
- Create unified feed generation method for single-feed mode
- Maintain existing per-directory logic for backward compatibility

### API Layer
- Update API service to pass feed configuration to feed service
- Modify feed endpoints to handle both modes appropriately
- Ensure proper HTTP responses for each mode (404 for invalid unified mode requests)

## Components and Interfaces

### Configuration Components

#### FeedConfig Struct
```go
type FeedConfig struct {
    Mode string `yaml:"mode"` // "per_directory" or "unified"
}
```

#### Updated APIConfig
```go
type APIConfig struct {
    Server   APIServerConfig `yaml:"server"`
    Database DatabaseConfig  `yaml:"database"`
    Storage  StorageConfig   `yaml:"storage"`
    External ExternalConfig  `yaml:"external"`
    Feed     FeedConfig      `yaml:"feed"`
}
```

### Feed Service Components

#### Enhanced FeedService
```go
type FeedService struct {
    coreservice  *core.CoreService
    feedBasePort string
    feedItemPath string
    feedConfig   *FeedConfig
}
```

#### Feed Generation Methods
- `GetFeeds(host string) ([]*feeds.Feed, error)` - Modified to handle both modes
- `getPerDirectoryFeeds(host string) ([]*feeds.Feed, error)` - Current logic extracted
- `getUnifiedFeed(host string) (*feeds.Feed, error)` - New unified feed logic

### API Service Components

#### Modified Endpoints
- `GET /v1/feeds` - Returns appropriate feed list based on mode
- `GET /v1/feeds/{feedTitle}/rss.xml` - Handles per-directory mode
- `GET /v1/feeds/all/rss.xml` - Handles unified mode (when enabled)

## Data Models

### Feed Configuration
```yaml
api:
  feed:
    mode: "per_directory" # or "unified"
```

### Unified Feed Structure
- **Title**: "All Podcast Items"
- **Description**: "Unified podcast feed containing all items"
- **Author**: "Video to Podcast Service"
- **Items**: All podcast items from all subdirectories, sorted by creation date

## Error Handling

### Configuration Errors
- Invalid feed mode values default to "per_directory" with warning log
- Missing feed configuration defaults to "per_directory"

### Runtime Errors
- Unified mode: Return 404 for individual feed requests (`/v1/feeds/{feedTitle}/rss.xml` where feedTitle != "Video-To-Podcast%20Feed")
- Database errors propagated as internal server errors

### Validation
- Feed mode validation during configuration loading
- Appropriate HTTP status codes for invalid requests in each mode

## Testing Strategy

### Unit Tests
- Configuration loading with various feed mode values
- Feed service behavior in both modes
- Feed generation logic for unified and per-directory modes
- Error handling for invalid configurations and requests

### Integration Tests
- End-to-end API testing for both feed modes
- Configuration file loading and application
- Feed endpoint responses in different modes
- Web interface feed link generation (should work automatically)

### Test Cases
1. **Per-directory mode tests**:
   - Multiple feeds generated correctly
   - Individual feed access works
   - Unified feed access returns 404

2. **Unified mode tests**:
   - Single unified feed generated
   - All items included regardless of subdirectory
   - Individual feed access returns 404
   - Unified feed access works

3. **Configuration tests**:
   - Default behavior (per-directory)
   - Invalid mode handling
   - Configuration file parsing