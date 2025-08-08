#!/bin/bash

# Entrypoint script for the API service
# Handles permission checks for mounted volumes and starts the application

# Fix permissions for mounted volumes
if [ -d "/app/resources" ]; then
    # Check if we can write to the directory
    if [ ! -w "/app/resources" ]; then
        echo "Warning: /app/resources is not writable by appuser. This may cause issues."
        echo "Consider using a named volume or fixing host directory permissions."
    fi
fi

# Check cookie file directory permissions if configured (optional feature)
if [ -n "$YTDLP_COOKIES_FILE" ]; then
    echo "Cookie file configured: $YTDLP_COOKIES_FILE"
    COOKIE_DIR=$(dirname "$YTDLP_COOKIES_FILE")
    
    # Check if the cookie file exists
    if [ -f "$YTDLP_COOKIES_FILE" ]; then
        # Cookie file exists, check if its directory is writable
        if [ ! -w "$COOKIE_DIR" ]; then
            echo "Warning: Cookie file exists but directory $COOKIE_DIR is not writable by appuser."
            echo "This may prevent cookie updates. Consider fixing host directory permissions."
        else
            echo "Cookie file is accessible and directory is writable."
        fi
    else
        # Cookie file doesn't exist, check if we can create it
        if [ -d "$COOKIE_DIR" ]; then
            if [ ! -w "$COOKIE_DIR" ]; then
                echo "Warning: Cookie directory $COOKIE_DIR exists but is not writable by appuser."
                echo "Cookie file cannot be created. Consider fixing host directory permissions."
            else
                echo "Cookie directory is ready for cookie file creation if needed."
            fi
        else
            echo "Warning: Cookie directory $COOKIE_DIR does not exist."
            echo "Cookie file will not be accessible. Consider creating the directory with proper permissions."
        fi
    fi
else
    echo "No cookie file configured (optional feature not enabled)."
fi

# Start the application
exec "$@"