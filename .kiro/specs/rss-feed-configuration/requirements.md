# Requirements Document

## Introduction

This feature enables users to configure how RSS feeds are provided by the video-to-podcast service. Currently, the system creates one RSS feed per subdirectory (1:1 mapping), where each subdirectory represents a separate channel/feed. This feature will add a configuration option to allow users to choose between the current behavior (one feed per subdirectory) or a new unified feed option (all items in one feed regardless of subdirectory).

## Requirements

### Requirement 1

**User Story:** As a system administrator, I want to configure RSS feed generation behavior, so that I can choose between per-directory feeds or a unified feed based on my use case.

#### Acceptance Criteria

1. WHEN the system starts THEN it SHALL read the RSS feed configuration from the configuration file
2. WHEN no RSS feed configuration is specified THEN the system SHALL default to the current behavior (one feed per subdirectory)
3. WHEN the RSS feed configuration is set to "per_directory" THEN the system SHALL create one RSS feed per subdirectory
4. WHEN the RSS feed configuration is set to "unified" THEN the system SHALL create one RSS feed containing all podcast items regardless of subdirectory

### Requirement 2

**User Story:** As a user, I want to access RSS feeds according to the configured behavior, so that I can subscribe to feeds in my preferred format.

#### Acceptance Criteria

1. WHEN RSS feed mode is "per_directory" THEN the system SHALL serve individual feeds at `/v1/feeds/{feedTitle}/rss.xml`
2. WHEN RSS feed mode is "unified" THEN the system SHALL serve a single unified feed at `/v1/feeds/Video-To-Podcast%20Feed/rss.xml`
3. WHEN RSS feed mode is "unified" THEN the system SHALL include all podcast items from all subdirectories in the unified feed
4. WHEN RSS feed mode is "unified" AND individual feed URLs are accessed THEN the system SHALL return a 404 Not Found response

### Requirement 3

**User Story:** As a user, I want the feed listing API to reflect the configured RSS feed behavior, so that I can discover available feeds correctly.

#### Acceptance Criteria

1. WHEN RSS feed mode is "per_directory" THEN the `/v1/feeds` endpoint SHALL return a list of all individual feed URLs
2. WHEN RSS feed mode is "unified" THEN the `/v1/feeds` endpoint SHALL return a single unified feed URL
3. WHEN RSS feed mode is "unified" THEN the unified feed SHALL have a descriptive title like "Video-To-Podcast Feed"

