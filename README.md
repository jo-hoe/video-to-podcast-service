# go-audio-rss-feeder

```bash
ffprobe -i .\video.mp4 -v quiet -of json -show_entries format
```

example podcast: <https://feeds.libsyn.com/230510/rss>
ID3: <https://www.exiftool.org/TagNames/ID3.html>


There might be an issue with setting the video URL.
Replace the audio.mp3 with a real converted YouTube file to check if when the tag is set via software, it is also readable on feed construction. You can just use the integration test to create such a video