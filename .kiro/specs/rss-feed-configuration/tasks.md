# Implementation Plan

- [x] 1. Add feed configuration to config structure
  - Add FeedConfig struct to config package with mode field
  - Update APIConfig struct to include Feed field
  - Update default configuration to include per_directory mode
  - Write unit tests for configuration loading with feed settings
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 2. Update configuration loading and validation
  - Modify getDefaultConfig to include default feed configuration
  - Add validation for feed mode values in LoadAPIConfig
  - Ensure backward compatibility when feed config is missing
  - Write tests for invalid feed mode handling
  - _Requirements: 1.1, 1.2_

- [x] 3. Enhance FeedService to support configuration modes
  - Add feedConfig field to FeedService struct
  - Update NewFeedService constructor to accept feed configuration
  - Modify GetFeeds method to route to appropriate mode-specific logic
  - Write unit tests for FeedService initialization with config
  - _Requirements: 1.3, 1.4_

- [x] 4. Implement per-directory feed generation (extract existing logic)
  - Extract current GetFeeds logic into getPerDirectoryFeeds method
  - Ensure existing behavior is preserved exactly
  - Write unit tests to verify per-directory mode functionality
  - _Requirements: 1.3, 2.1_

- [x] 5. Implement unified feed generation
  - Create getUnifiedFeed method that aggregates all podcast items
  - Set unified feed title to "All Podcast Items" and appropriate metadata
  - Sort items by creation date for consistent ordering
  - Write unit tests for unified feed generation
  - _Requirements: 1.4, 2.3, 3.3_

- [x] 6. Update API service to us
e feed configuration
  - Modify APIService to load and pass feed configuration to FeedService
  - Update getFeedService method to include feed config parameter
  - Write integration tests for API service with different feed modes
  - _Requirements: 1.1, 1.3, 1.4_

- [x] 7. Implement mode-specific feed endpoint routing
  - Update feedsHandler to return appropriate feed list based on mode
  - Modify feedHandler to handle unified mode requests (/v1/feeds/all/rss.xml)
  - Add 404 responses for invalid requests in each mode
  - Write integration tests for endpoint behavior in both modes
  - _Requirements: 2.1, 2.2, 2.4, 3.1, 3.2_

- [x] 8. Add comprehensive integration tests
  - Test complete flow for per-directory mode (default behavior)
  - Test complete flow for unified mode with multiple subdirectories
  - Verify backward compatibility with existing configurations
  - Test error scenarios and HTTP status codes
  - _Requirements: 1.2, 2.1, 2.2, 2.3, 2.4, 3.1, 3.2_

- [x] 9. Update configuration files with feed settings
  - Add feed configuration section to config.yaml
  - Add feed configuration section to config.docker.yaml
  - Include comments explaining the available modes
  - _Requirements: 1.1, 1.2_