package youtube

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
)

const (
	playlistRegex        = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
	youtubeVideoRegex    = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`
	youtubeTinyLinkRegex = `https://youtu\.be/([A-Za-z0-9_-]*)`
	// types taken from API description
	// https://wiki.sponsor.ajay.app/w/Types
	sponsorBlockCategories = "sponsor,selfpromo,interaction,intro,outro,preview,music_offtopic,filler,hook"
)

var (
	playlistPattern     = regexp.MustCompile(playlistRegex)
	youtubeVideoPattern = regexp.MustCompile(youtubeVideoRegex)
	youtubeTinyPattern  = regexp.MustCompile(youtubeTinyLinkRegex)
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

func (y *YoutubeAudioDownloader) Download(url string, targetPath string) (string, error) {
	// Create a unique subdirectory within the configured temp path for download processing
	tempPath, err := os.MkdirTemp(y.mediaConfig.TempPath, "youtube-download-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(tempPath); err != nil {
			slog.Warn("error removing temp directory", "err", err)
		}
	}()

	slog.Info("downloading", "url", url, "tempPath", tempPath)
	tempResults, err := y.download(tempPath, url)
	if err != nil {
		return "", err
	}
	if len(tempResults) == 0 {
		return "", fmt.Errorf("no audio files downloaded for url %s", url)
	}
	// Expect single file for a single video URL, but pick the first if multiple are found
	filePath := tempResults[0]
	slog.Info("done downloading file", "filePath", filePath)

	slog.Info("setting metadata", "filePath", filePath)
	if err = y.setMetadata(filePath); err != nil {
		return "", err
	}
	slog.Info("set metadata", "filePath", filePath)

	slog.Info("moving file to target folder")
	result, err := filemanagement.MoveToTarget(filePath, targetPath)
	if err != nil {
		return "", err
	}
	slog.Info("completed moving file", "targetPath", result)

	return result, nil
}

func (y *YoutubeAudioDownloader) setMetadata(fullFilePath string) error {
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
	metadata[downloader.VideoDownloadLink] = metadata[downloader.VideoURLID3Key]

	videoURL := metadata[downloader.VideoURLID3Key]

	thumbnailURL, err := y.getThumbnailURL(videoURL)
	if err != nil {
		return err
	}
	metadata[downloader.ThumbnailUrlTag] = thumbnailURL

	if ts, tsErr := y.getTimestamp(videoURL); tsErr == nil {
		metadata["date"] = time.Unix(ts, 0).UTC().Format("2006-01-02T15:04:05")
	} else {
		slog.Warn("could not get timestamp, will fall back to date tag", "err", tsErr)
	}

	return mp3joiner.SetFFmpegMetadataTag(fullFilePath, metadata, chapters)
}

func (y *YoutubeAudioDownloader) getThumbnailURL(videoURL string) (string, error) {
	args := y.buildBaseArgs(true)
	args = append(args, "--print", "thumbnail", videoURL)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error getting thumbnail url", "videoURL", videoURL, "err", err)
		return "", err
	}

	return downloader.FirstHTTPSLineFromOutput(output), nil
}

func (y *YoutubeAudioDownloader) getTimestamp(videoURL string) (int64, error) {
	args := y.buildBaseArgs(true)
	args = append(args, "--print", "timestamp", videoURL)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("yt-dlp timestamp fetch failed: %w", err)
	}

	ts := strings.TrimSpace(string(output))
	if ts == "" || ts == "NA" {
		return 0, fmt.Errorf("timestamp not available for %s", videoURL)
	}
	return strconv.ParseInt(ts, 10, 64)
}

func (y *YoutubeAudioDownloader) download(targetDirectory string, url string) ([]string, error) {
	// set download behavior
	tempFilenameTemplate := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, "%(channel)s/%(title)s_%(id)s.%(ext)s")

	args := y.buildBaseArgs(false)
	args = append(args,
		"--extract-audio",
		"--audio-format", "mp3",
		"--embed-metadata",
		"--no-progress",
		"--sponsorblock-remove", sponsorBlockCategories,
		// Ignore SponsorBlock API errors so downloads continue even when the API is unavailable
		"--ignore-errors",
	)

	args = append(args,
		// Workaround: using lower resolution to avoid issues with download of videos
		// Remove when after upstream fix of
		// https://github.com/yt-dlp/yt-dlp/issues/12482
		// is available and integration tests pass without this code.
		"--format", "bestaudio/best[height<=360]",
		"--output", tempFilenameTemplate,
		url,
	)

	cmd := exec.Command("yt-dlp", args...)
	slog.Info("constructed yt-dlp command", "args", args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp command failed: %w", err)
	}

	return filemanagement.GetAudioFiles(targetDirectory)
}

func (y *YoutubeAudioDownloader) IsVideoSupported(url string) bool {
	return playlistPattern.MatchString(url) ||
		youtubeVideoPattern.MatchString(url) ||
		youtubeTinyPattern.MatchString(url)
}

func (y *YoutubeAudioDownloader) IsVideoAvailable(url string) bool {
	slog.Info("checking video availability", "url", url)

	// Use yt-dlp to print live_status in a dry run.
	// Treat videos that are currently livestreaming ("is_live") as not available.
	args := y.buildBaseArgs(true)
	args = append(args, "--print", downloader.LiveStatusKey, url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error checking video availability", "err", err)
		return false
	}

	if downloader.IsLiveFromOutput(output) {
		slog.Warn("video is currently live; treating as unavailable", "url", url)
		return false
	}

	return true
}

// buildBaseArgs creates base arguments for yt-dlp command.
// When simulate is true, adds --simulate and --quiet flags for dry-run operations.
func (y *YoutubeAudioDownloader) buildBaseArgs(simulate bool) []string {
	args := downloader.AppendCookieArgs(make([]string, 0), y.cookiesConfig)

	// Workaround: use web_safari client
	// Remove after upstream fix of https://github.com/yt-dlp/yt-dlp/issues/12482
	// is available and integration tests pass without this code.
	args = append(args, "--extractor-args", "youtube:player_client=default,web_safari;player_js_version=actual")

	if simulate {
		args = append(args, "--simulate", "--quiet")
	}

	return args
}

// ListIndividualVideoURLs returns individual video URLs for a given input URL.
// For playlist URLs, it returns all video URLs in the playlist.
// For single video URLs, it returns a slice containing the original URL.
func (y *YoutubeAudioDownloader) ListIndividualVideoURLs(url string) ([]string, error) {
	if !playlistPattern.MatchString(url) {
		return []string{url}, nil
	}

	args := y.buildBaseArgs(true)
	// Use flat playlist to avoid resolving each entry and just print the URL
	args = append(args, "--flat-playlist", "--print", "url", url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error listing playlist entries", "url", url, "err", err)
		return nil, err
	}

	entries := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			entries = append(entries, line)
		}
	}
	return entries, nil
}
