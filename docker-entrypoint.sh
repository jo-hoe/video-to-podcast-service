#!/bin/sh
set -e

if [ "${YT_DLP_UPDATE_TO_NIGHTLY}" = "true" ]; then
    yt-dlp --update-to nightly
fi

exec ./app
