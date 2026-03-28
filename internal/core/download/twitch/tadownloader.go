package twitch

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
	twitchVodRegex     = `https://(?:www\.)?twitch\.tv/videos/(\d+)`
	twitchClipRegex    = `https://(?:www\.)?twitch\.tv/[A-Za-z0-9_]+/clip/([A-Za-z0-9_-]+)`
	twitchClipsRegex   = `https://clips\.twitch\.tv/([A-Za-z0-9_-]+)`
	liveStatusLiveValue = "is_live"
	liveStatusKeyAttribute = "live_status"
	// ID3 tag yt-dlp uses to store the source video URL
	videoUrlID3KeyAttribute = "purl"
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
	return regexp.MustCompile(twitchVodRegex).MatchString(url) ||
		regexp.MustCompile(twitchClipRegex).MatchString(url) ||
		regexp.MustCompile(twitchClipsRegex).MatchString(url)
}

func (t *TwitchAudioDownloader) IsVideoAvailable(url string) bool {
	slog.Info("checking video availability", "url", url)

	args := t.buildBaseArgs(true)
	args = append(args, "--print", liveStatusKeyAttribute, url)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("error checking video availability", "err", err)
		return false
	}

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		v := strings.TrimSpace(line)
		if v == "" {
			continue
		}
		if v == liveStatusLiveValue {
			slog.Warn("stream is currently live; treating as unavailable", "url", url, liveStatusKeyAttribute, v)
			return false
		}
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
	result, err := moveToTarget(filePath, targetPath)
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
	metadata[downloader.VideoDownloadLink] = metadata[videoUrlID3KeyAttribute]

	thumbnailURL, err := t.getThumbnailURL(sourceURL)
	if err != nil {
		return err
	}
	metadata[downloader.ThumbnailUrlTag] = thumbnailURL

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

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.HasPrefix(line, "https") {
			return line, nil
		}
	}

	return "", nil
}

func moveToTarget(sourcePath, targetRootPath string) (string, error) {
	directoryName := filepath.Base(filepath.Dir(sourcePath))
	targetSubDirectory := filepath.Join(targetRootPath, directoryName)
	if err := os.MkdirAll(targetSubDirectory, os.ModePerm); err != nil {
		return "", err
	}

	targetPath := filepath.Join(targetSubDirectory, filepath.Base(sourcePath))
	if err := filemanagement.MoveFile(sourcePath, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}

func (t *TwitchAudioDownloader) buildBaseArgs(simulate bool) []string {
	args := make([]string, 0)

	if t.cookiesConfig != nil && t.cookiesConfig.Enabled && t.cookiesConfig.CookiePath != "" {
		if _, err := os.Stat(t.cookiesConfig.CookiePath); err == nil {
			slog.Info("using cookie file path", "path", t.cookiesConfig.CookiePath)
			args = append(args, "--cookies", t.cookiesConfig.CookiePath)
		} else {
			slog.Warn("cookie file path specified but not found", "path", t.cookiesConfig.CookiePath)
		}
	}

	if simulate {
		args = append(args, "--simulate", "--quiet")
	}

	return args
}
