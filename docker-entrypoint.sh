#!/bin/sh
set -e

# When YT_DLP_UPDATE_TO_NIGHTLY is true, switch yt-dlp to the nightly build before
# starting the app. Useful when the stable release is broken and the fix is only
# available in nightly. The binary lives in /home/appuser/bin (owned by appuser)
# so no elevated permissions are needed.
#
# Alternative considered: an initContainer running as root to pre-populate the
# binary into a shared volume before the main container starts. Rejected because
# emptyDir volumes are ephemeral (re-download on every pod restart) and using the
# data PVC adds unnecessary coupling between the binary and app data.
if [ "${YT_DLP_UPDATE_TO_NIGHTLY}" = "true" ]; then
    yt-dlp --update-to nightly
fi

exec ./app
