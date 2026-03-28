package twitch

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
	twitchVodRegex   = `https://(?:www\.)?twitch\.tv/videos/(\d+)`
	twitchClipRegex  = `https://(?:www\.)?twitch\.tv/[A-Za-z0-9_]+/clip/([A-Za-z0-9_-]+)`
	twitchClipsRegex = `https://clips\.twitch\.tv/([A-Za-z0-9_-]+)`
)

var (
	twitchVodPattern   = regexp.MustCompile(twitchVodRegex)
	twitchClipPattern  = regexp.MustCompile(twitchClipRegex)
	twitchClipsPattern = regexp.MustCompile(twitchClipsRegex)
)

type TwitchAudioDownloader struct {
	cookiesConfig *config.Cookies
	mediaConfig   *config.Media
}

func NewTwitchAudioDownloader(cookiesConfig *config.Cookies, mediaConfig *config.Media) *TwitchAudioDownloader {
	return &TwitchAudioDownloader{
		cookiesConfig: cookiesConfig,
		mediaConfig:   mediaConfig,
	}
}

func (t *TwitchAudioDownloader) IsVideoSupported(url string) bool {
	return twitchVodPattern.MatchString(url) ||
		twitchClipPattern.MatchString(url) ||
		twitchClipsPattern.MatchString(url)
}

func (t *TwitchAudioDownloader) IsVideoAvailable(url string) bool {
	slog.Info("checking video availability", "url", url)

	args := t.buildBaseArgs(true)
	args = append(args, "--print", downloader.LiveStatusKey, url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error checking video availability", "err", err)
		return false
	}

	if downloader.IsLiveFromOutput(output) {
		slog.Warn("stream is currently live; treating as unavailable", "url", url)
		return false
	}

	return true
}

func (t *TwitchAudioDownloader) ListIndividualVideoURLs(url string) ([]string, error) {
	return []string{url}, nil
}

func (t *TwitchAudioDownloader) Download(url string, targetPath string) (string, error) {
	tempPath, err := os.MkdirTemp(t.mediaConfig.TempPath, "twitch-download-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(tempPath); err != nil {
			slog.Warn("error removing temp directory", "err", err)
		}
	}()

	slog.Info("downloading", "url", url, "tempPath", tempPath)
	filePaths, err := t.download(tempPath, url)
	if err != nil {
		return "", err
	}
	if len(filePaths) == 0 {
		return "", fmt.Errorf("no audio files downloaded for url %s", url)
	}
	filePath := filePaths[0]
	slog.Info("done downloading file", "filePath", filePath)

	slog.Info("setting metadata", "filePath", filePath)
	if err = t.setMetadata(filePath, url); err != nil {
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

func (t *TwitchAudioDownloader) download(targetDirectory string, url string) ([]string, error) {
	tempFilenameTemplate := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, "%(uploader)s/%(title)s_%(id)s.%(ext)s")

	args := t.buildBaseArgs(false)
	args = append(args,
		"--extract-audio",
		"--audio-format", "mp3",
		"--embed-metadata",
		"--no-progress",
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

func (t *TwitchAudioDownloader) setMetadata(fullFilePath string, sourceURL string) error {
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

	thumbnailURL, err := t.getThumbnailURL(sourceURL)
	if err != nil {
		return err
	}
	metadata[downloader.ThumbnailUrlTag] = thumbnailURL

	if ts, tsErr := t.getTimestamp(sourceURL); tsErr == nil {
		metadata["date"] = time.Unix(ts, 0).UTC().Format("2006-01-02T15:04:05")
	} else {
		slog.Warn("could not get timestamp, will fall back to date tag", "err", tsErr)
	}

	return mp3joiner.SetFFmpegMetadataTag(fullFilePath, metadata, chapters)
}

func (t *TwitchAudioDownloader) getThumbnailURL(url string) (string, error) {
	args := t.buildBaseArgs(true)
	args = append(args, "--print", "thumbnail", url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error getting thumbnail url", "url", url, "err", err)
		return "", err
	}

	return downloader.FirstHTTPSLineFromOutput(output), nil
}

func (t *TwitchAudioDownloader) getTimestamp(url string) (int64, error) {
	args := t.buildBaseArgs(true)
	args = append(args, "--print", "timestamp", url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("yt-dlp timestamp fetch failed: %w", err)
	}

	ts := strings.TrimSpace(string(output))
	if ts == "" || ts == "NA" {
		return 0, fmt.Errorf("timestamp not available for %s", url)
	}
	return strconv.ParseInt(ts, 10, 64)
}

// buildBaseArgs creates base arguments for yt-dlp command.
// When simulate is true, adds --simulate and --quiet flags for dry-run operations.
func (t *TwitchAudioDownloader) buildBaseArgs(simulate bool) []string {
	args := downloader.AppendCookieArgs(make([]string, 0), t.cookiesConfig)

	if simulate {
		args = append(args, "--simulate", "--quiet")
	}

	return args
}
