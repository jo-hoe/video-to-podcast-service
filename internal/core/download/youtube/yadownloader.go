package youtube

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
)

const (
	playlistRegex        = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
	youtubeVideoRegex    = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`
	youtubeTinyLinkRegex = `https://youtu\.be/([A-Za-z0-9_-]*)`
	youtubeShortsRegex   = `https://(?:.+)?youtube.com/shorts/([A-Za-z0-9_-]*)`
	// types taken from API description
	// https://wiki.sponsor.ajay.app/w/Types
	sponsorBlockCategories = "sponsor,selfpromo,interaction,intro,outro,preview,music_offtopic,filler,hook"
	// ID3 tag youtube-dl uses to store the video URL
	VideoUrlID3KeyAttribute = "purl"

	liveStatusKeyAttribute = "live_status"
	liveStatusLiveValue    = "is_live"
)

type YoutubeAudioDownloader struct {
	cookiesConfig *config.Cookies
	mediaConfig   *config.Media
}

func NewYoutubeAudioDownloader(cookiesConfig *config.Cookies, mediaConfig *config.Media) *YoutubeAudioDownloader {
	return &YoutubeAudioDownloader{
		cookiesConfig: cookiesConfig,
		mediaConfig:   mediaConfig,
	}
}

func (y *YoutubeAudioDownloader) Download(urlString string, targetPath string) (string, error) {
	var result string

	// Create a unique subdirectory within the configured temp path for download processing
	tempPath, err := os.MkdirTemp(y.mediaConfig.TempPath, "youtube-download-")
	if err != nil {
		return result, err
	}
	defer func() {
		if err := os.RemoveAll(tempPath); err != nil {
			slog.Warn("error removing temp directory", "err", err)
		}
	}()

	slog.Info("downloading", "url", urlString, "tempPath", tempPath)
	tempResults, err := y.download(tempPath, urlString)
	if err != nil {
		return "", err
	}
	if len(tempResults) == 0 {
		return "", fmt.Errorf("no audio files downloaded for url %s", urlString)
	}
	// Expect single file for a single video URL, but pick the first if multiple are found
	filePath := tempResults[0]
	slog.Info("done downloading file", "filePath", filePath)

	slog.Info("setting metadata", "filePath", filePath)
	err = y.setMetadata(filePath)
	if err != nil {
		return "", err
	}
	slog.Info("set metadata", "filePath", filePath)

	slog.Info("moving file to target folder")
	movedItem, err := moveToTarget(filePath, targetPath)
	if err != nil {
		return "", err
	}
	result = movedItem
	slog.Info("completed moving file", "targetPath", result)

	return result, nil
}

func (y *YoutubeAudioDownloader) setMetadata(fullFilePath string) (err error) {
	metadata, err := mp3joiner.GetFFmpegMetadataTag(fullFilePath)
	if err != nil {
		return err
	}
	chapters, err := mp3joiner.GetChapterMetadata(fullFilePath)
	if err != nil {
		return err
	}

	metadata[downloader.PodcastDescriptionTag] = strings.ReplaceAll(metadata["synopsis"], "\n", "<br>")
	metadata[downloader.DateTag] = metadata["date"]
	metadata[downloader.VideoDownloadLink] = metadata[VideoUrlID3KeyAttribute]

	videoUrl := metadata["purl"]
	thumbnailUrl, err := y.getThumbnailUrl(videoUrl)
	if err != nil {
		return err
	}
	metadata[downloader.ThumbnailUrlTag] = thumbnailUrl

	return mp3joiner.SetFFmpegMetadataTag(fullFilePath, metadata, chapters)
}

func (y *YoutubeAudioDownloader) getThumbnailUrl(videoUrl string) (result string, err error) {
	args := y.buildBaseArgs(true)
	args = append(args, "--print", "thumbnail", videoUrl)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error getting thumbnail url", "videoUrl", videoUrl, "err", err)
		return result, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "https") {
			result = line
		}
	}

	return result, err
}

func moveToTarget(sourcePath, targetRootPath string) (results string, err error) {
	// move file to target directory
	// the target directory is created based on the source file's parent directory
	//
	// e.g.:
	// sourcePath = /tmp/1234/5678/file.mp3
	// targetRootPath = /podcasts
	// results = /podcasts/5678/file.mp3
	directoryName := filepath.Base(filepath.Dir(sourcePath))
	targetSubDirectory := filepath.Join(targetRootPath, directoryName)
	err = os.MkdirAll(targetSubDirectory, os.ModePerm)
	if err != nil {
		return results, err
	}

	targetFilename := filepath.Base(sourcePath)
	targetPath := filepath.Join(targetSubDirectory, targetFilename)
	err = filemanagement.MoveFile(sourcePath, targetPath)
	if err != nil {
		return results, err
	}
	return targetPath, err
}

func (y *YoutubeAudioDownloader) download(targetDirectory string, urlString string) ([]string, error) {
	result := make([]string, 0)

	// set download behavior
	tempFilenameTemplate := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, "%(channel)s/%(title)s_%(id)s.%(ext)s")

	args := y.buildBaseArgs(false)
	args = append(args,
		"--extract-audio",
		"--audio-format", "mp3",
		"--embed-metadata",
		"--no-progress",
		"--sponsorblock-remove", sponsorBlockCategories,
		// Workaround: using lower resolution to avoid issues with download of videos
		// Remove when after upstream fix of
		// https://github.com/yt-dlp/yt-dlp/issues/12482
		// is available and integration tests pass without this code.
		"--format", "bestaudio/best[height<=360]",
		"--output", tempFilenameTemplate,
		urlString,
	)

	cmd := exec.Command("yt-dlp", args...)
	slog.Info("constructed yt-dlp command", "args", args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return result, fmt.Errorf("yt-dlp command failed: %w", err)
	}

	// get file names
	result, err = filemanagement.GetAudioFiles(targetDirectory)
	return result, err
}

func (y *YoutubeAudioDownloader) IsVideoSupported(url string) bool {
	return regexp.MustCompile(playlistRegex).MatchString(url) ||
		regexp.MustCompile(youtubeVideoRegex).MatchString(url) ||
		regexp.MustCompile(youtubeTinyLinkRegex).MatchString(url)
}

func (y *YoutubeAudioDownloader) IsVideoAvailable(urlString string) bool {
	slog.Info("checking video availability", "url", urlString)

	// Use yt-dlp to print live_status in a dry run.
	// Treat videos that are currently livestreaming ("is_live") as not available.
	args := y.buildBaseArgs(true)
	args = append(args, "--print", liveStatusKeyAttribute, urlString)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error checking video availability", "err", err)
		return false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		v := strings.TrimSpace(line)
		if v == "" {
			continue
		}
		// If any line indicates "is_live", consider unavailable immediately
		if v == liveStatusLiveValue {
			slog.Warn("video is currently live; treating as unavailable", "url", urlString, liveStatusKeyAttribute, v)
			return false
		}
	}

	// If not live, consider available (dry run succeeded)
	return true
}

// buildBaseArgs creates base arguments for yt-dlp command
// simulate: if true, adds --simulate and --quiet flags for dry-run operations
func (y *YoutubeAudioDownloader) buildBaseArgs(simulate bool) []string {
	args := make([]string, 0)

	// Add cookie configuration if enabled
	if y.cookiesConfig != nil && y.cookiesConfig.Enabled && y.cookiesConfig.CookiePath != "" {
		if _, err := os.Stat(y.cookiesConfig.CookiePath); err == nil {
			slog.Info("using cookie file path", "path", y.cookiesConfig.CookiePath)
			args = append(args, "--cookies", y.cookiesConfig.CookiePath)
		} else {
			slog.Warn("cookie file path specified but not found", "path", y.cookiesConfig.CookiePath)
		}
	}

	// Workaround: use web_safari client
	// Remove when after upstream fix of https://github.com/yt-dlp/yt-dlp/issues/12482 is available and integration tests pass without this code.
	args = append(args, "--extractor-args", "youtube:player_client=default,web_safari;player_js_version=actual")

	if simulate {
		args = append(args, "--simulate", "--quiet")
	}

	return args
}

// ListIndividualVideoURLs returns individual video URLs for a given input URL.
// For playlist URLs, it returns all video URLs in the playlist.
// For single video URLs, it returns a slice containing the original URL.
func (y *YoutubeAudioDownloader) ListIndividualVideoURLs(urlString string) ([]string, error) {
	entries := make([]string, 0)

	// Detect playlist URLs
	if regexp.MustCompile(playlistRegex).MatchString(urlString) {
		args := y.buildBaseArgs(true)
		// Use flat playlist to avoid resolving each entry and just print the URL
		args = append(args, "--flat-playlist", "--print", "url", urlString)

		cmd := exec.Command("yt-dlp", args...)
		output, err := cmd.Output()
		if err != nil {
			slog.Error("error listing playlist entries", "url", urlString, "err", err)
			return nil, err
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				entries = append(entries, line)
			}
		}
		return entries, nil
	}

	// Not a playlist, return the original URL
	return []string{urlString}, nil
}
