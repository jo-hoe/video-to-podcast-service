# Requirements Document

## Introduction

This feature addresses two critical issues in the video-to-podcast service: incorrect feed URLs in the UI and missing cookie storage writability checks in the API service. The feed URLs currently point to the wrong service, and the API service doesn't verify that cookie storage locations are writable, which can cause yt-dlp failures.

## Requirements

### Requirement 1

**User Story:** As a user viewing the podcast items in the UI, I want the RSS feed links to point to the correct API service endpoint, so that I can access the actual RSS feeds.

#### Acceptance Criteria

1. WHEN a user views a podcast item in the UI THEN the RSS feed link SHALL point to the API service endpoint
2. WHEN a user clicks on an RSS feed link THEN the system SHALL serve the RSS feed from the API service at the correct URL format
3. WHEN the feed URL is constructed THEN it SHALL use the API service base URL instead of the UI service host
4. WHEN the feed URL is generated THEN it SHALL follow the format `{api_base_url}/v1/feeds/{feed_title}/rss.xml`

### Requirement 2

**User Story:** As a system administrator, I want the API service to verify that cookie storage locations are writable during startup, so that yt-dlp operations don't fail due to permission issues.

#### Acceptance Criteria

1. WHEN the API service starts up THEN it SHALL check if the cookie storage directory is writable
2. WHEN the cookie storage directory is not writable THEN the system SHALL log a warning message
3. WHEN yt-dlp needs to store cookies THEN the system SHALL have write access to the cookie storage location
4. WHEN the writability check fails THEN the system SHALL provide guidance on fixing permissions
5. WHEN the API service runs in a container THEN the cookie storage check SHALL work correctly with mounted volumes